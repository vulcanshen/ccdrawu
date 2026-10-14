package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vulcanshen/ccdrawu/internal/claude"
	"github.com/vulcanshen/ccdrawu/internal/config"
	terminu "github.com/vulcanshen/terminu-go"
)

var testCatalog = config.Catalog{Version: 1, Items: []config.Item{
	{ID: "model", Label: "Model", Name: "Model", Group: "Model", Dock: true, Inline: true},
	{ID: "mode", Label: "Mode", Name: "Permission mode", Group: "Model", Dock: true, Inline: false},
	{ID: "context", Label: "Context", Name: "Context", Group: "Usage", Dock: true, Inline: true},
	{ID: "git", Label: "Git", Name: "Git branch and changes", Group: "Workspace", Dock: false, Inline: true},
}}

// claudeFake answers `claude plugin ...` as Claude Code does, and keeps what ran
type claudeFake struct {
	installed bool
	ran       []string
}

func (f *claudeFake) run(name string, args ...string) ([]byte, error) {
	line := strings.Join(args, " ")
	f.ran = append(f.ran, line)
	switch {
	case strings.HasPrefix(line, "plugin list"):
		if f.installed {
			return []byte(`[{"id": "ccdrawu@ccdrawu", "enabled": true}]`), nil
		}
		return []byte(`[]`), nil
	case strings.HasPrefix(line, "plugin install"):
		f.installed = true
		return []byte(`{"outcome": "ok", "message": "Installed ccdrawu"}`), nil
	case strings.HasPrefix(line, "plugin uninstall"), strings.HasPrefix(line, "plugin marketplace remove"):
		f.installed = false
		return []byte(`{"outcome": "ok", "message": "Uninstalled ccdrawu"}`), nil
	}
	return nil, errors.New("no answer for " + line)
}

// newTestDrawer opens the TUI on files in a temporary folder: a catalog, and a
// Claude Code folder whose settings.json selects the dark theme
func newTestDrawer(t *testing.T) (*drawer, config.Locations) {
	t.Helper()
	dir := t.TempDir()
	loc := config.Locations{Data: filepath.Join(dir, "data"), ConfigFile: filepath.Join(dir, "config", "config.json"), Claude: filepath.Join(dir, "claude")}
	for _, d := range []string{loc.Data, loc.Claude} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	data, _ := json.Marshal(testCatalog)
	if err := os.WriteFile(filepath.Join(loc.Data, "catalog.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(loc.Claude, "settings.json"), []byte("{\n  \"theme\": \"dark\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := newDrawer(loc, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	d.app.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	return d, loc
}

// tdp L1 and L4: at 80 x 40 every row is exactly the terminal's width
func TestL4EveryRowFullWidthAt80x40(t *testing.T) {
	d, _ := newTestDrawer(t)
	frame := d.app.Frame()
	if len(frame) != 40 {
		t.Fatalf("%d rows, want 40", len(frame))
	}
	for i, line := range frame {
		if w := terminu.Width(line); w != 80 {
			t.Fatalf("row %d is %d cells wide, want 80", i, w)
		}
	}
}

func key(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

// press sends msg and plays out what it starts: a popup opens and closes in
// frames the app sends itself (tdp F2). A command still waiting after 100ms,
// a timer, is dropped.
func press(d *drawer, msg tea.Msg) {
	update := func(m tea.Msg) tea.Cmd {
		_, cmd := d.app.Update(m)
		return cmd
	}
	queue := []tea.Cmd{update(msg)}
	for n := 0; len(queue) > 0 && n < 500; n++ {
		cmd := queue[0]
		queue = queue[1:]
		if cmd == nil {
			continue
		}
		got := make(chan tea.Msg, 1)
		go func() { got <- cmd() }()
		select {
		case m := <-got:
			if batch, ok := m.(tea.BatchMsg); ok {
				queue = append(queue, batch...)
			} else if m != nil {
				queue = append(queue, update(m))
			}
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func frameText(d *drawer) string { return terminu.StripStyles(strings.Join(d.app.Frame(), "\n")) }

var space = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}

// [2] of sidebar and inline lists what each group shows; Enter on any row, or
// e, opens every item in one checkbox popup, where Space checks, Esc drops it
// all and Enter applies it all
func TestEditPicksTheItemsInOnePopup(t *testing.T) {
	d, loc := newTestDrawer(t)
	// [1] starts on sidebar; Enter goes into [2] (tdp K3, details)
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if d.app.Screen().Focused() != d.detail {
		t.Fatalf("Enter on [1] left the focus there:\n%s", frameText(d))
	}
	text := frameText(d)
	for _, want := range []string{"Model Mode", "Context", "none", "Enter:edit"} {
		if !strings.Contains(text, want) {
			t.Fatalf("no %q on the sidebar page:\n%s", want, text)
		}
	}
	if strings.Contains(text, "e:edit") {
		t.Fatalf("the hint names e as well as Enter:\n%s", text)
	}

	dock := func() []string {
		t.Helper()
		c, err := config.ReadConfig(loc.ConfigFile)
		if err != nil {
			t.Fatal(err)
		}
		if c.Inline != nil {
			t.Fatalf("inline = %v, want it left on its defaults", c.Inline)
		}
		return c.Dock
	}

	press(d, key("j"))
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // on Usage: the popup of every item, on Usage's first
	if text := frameText(d); !strings.Contains(text, "EDIT: sidebar") || !strings.Contains(text, "Mode     Permission mode") {
		t.Fatalf("Enter on a row did not open the popup of every item with their keys:\n%s", text)
	}
	press(d, space) // Context off
	press(d, tea.KeyMsg{Type: tea.KeyEsc})
	if _, err := os.Stat(loc.ConfigFile); err == nil {
		t.Fatal("Esc kept what was checked")
	}

	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	press(d, space) // Context off
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if got, want := dock(), []string{"model", "mode"}; !slices.Equal(got, want) {
		t.Fatalf("dock = %v from Usage, want %v", got, want)
	}

	press(d, key("e")) // the same popup, on the first item of all
	press(d, space)    // Model off
	for range 3 {      // Mode, Context, Git
		press(d, key("j"))
	}
	press(d, space) // Git on
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if got, want := dock(), []string{"mode", "git"}; !slices.Equal(got, want) {
		t.Fatalf("dock = %v after e, want %v", got, want)
	}
	if !strings.Contains(frameText(d), "Git") {
		t.Fatalf("[2] does not show Git once applied:\n%s", frameText(d))
	}
}

func TestNoCatalogSaysWhatToDo(t *testing.T) {
	dir := t.TempDir()
	d, err := newDrawer(config.Locations{Data: filepath.Join(dir, "data"), ConfigFile: filepath.Join(dir, "config.json")}, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	d.app.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	if !strings.Contains(terminu.StripStyles(strings.Join(d.app.Frame(), "\n")), "ccdrawu install") {
		t.Fatal("an empty panel names ccdrawu install")
	}
}

// tdp components/layout/panel, groups and details: [1] in groups, the cursor
// skipping their titles; [2] is the row under it, named after it
func TestDetailFollowsTheSidebar(t *testing.T) {
	d, _ := newTestDrawer(t)
	text := frameText(d)
	for _, want := range []string{"[1] PREFERENCES", "DRAWER", "SETTINGS", "[2] SIDEBAR", "Model Mode"} {
		if !strings.Contains(text, want) {
			t.Fatalf("no %q in\n%s", want, text)
		}
	}
	d.app.Update(key("j"))
	if d.detail.Label != rowInline {
		t.Fatalf("[2] is %q after j, want inline", d.detail.Label)
	}
	d.app.Update(key("j")) // over the SETTINGS title
	text = frameText(d)
	if d.detail.Label != rowPreferences || !strings.Contains(text, "[2] PREFERENCES") {
		t.Fatalf("[2] is %q after j j, want preferences:\n%s", d.detail.Label, text)
	}
	// Every setting on that one page
	for _, want := range []string{"ALIAS", "none", "CLAUDE THEME", "DRAWER BACKGROUND", "PLUGIN"} {
		if !strings.Contains(text, want) {
			t.Fatalf("no %q on the preferences page:\n%s", want, text)
		}
	}
}

func TestAliasIsCheckedAndSaved(t *testing.T) {
	d, loc := newTestDrawer(t)
	press(d, key("j"))
	press(d, key("j")) // sidebar, inline, preferences; its first field is the alias
	press(d, tea.KeyMsg{Type: tea.KeyTab})
	edit := func(v string) {
		press(d, tea.KeyMsg{Type: tea.KeyEnter})
		press(d, key(v))
		press(d, tea.KeyMsg{Type: tea.KeyEnter})
	}
	edit("two words")
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Alias != "" {
		t.Fatalf("alias %q saved, want a name the mod takes or nothing", c.Alias)
	}
	if !strings.Contains(frameText(d), "letters, digits and -") {
		t.Fatalf("the input says nothing of what is wrong:\n%s", frameText(d))
	}
	press(d, tea.KeyMsg{Type: tea.KeyEsc})
	edit("/ccd")
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Alias != "ccd" {
		t.Fatalf("alias = %q, want ccd", c.Alias)
	}
}

func claudeTheme(t *testing.T, loc config.Locations) string {
	t.Helper()
	var s struct{ Theme string }
	data, _ := os.ReadFile(filepath.Join(loc.Claude, "settings.json"))
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("settings.json: %v\n%s", err, data)
	}
	return s.Theme
}

func TestThemeGoesIntoClaudeCodeAndBack(t *testing.T) {
	d, loc := newTestDrawer(t)
	for range 2 { // sidebar, inline, preferences
		press(d, key("j"))
	}
	if d.detail.Label != rowPreferences {
		t.Fatalf("[2] is %q, want preferences", d.detail.Label)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	for range 6 { // alias, then the theme's off, dark, light, dark-daltonized, light-daltonized, dark-ansi
		press(d, key("j"))
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if got := claudeTheme(t, loc); got != claude.ThemeSetting {
		t.Fatalf("Claude Code's theme = %q, want %q", got, claude.ThemeSetting)
	}
	c, _ := config.ReadConfig(loc.ConfigFile)
	if c.Theme.Base != "dark-ansi" || c.Theme.Before != "dark" {
		t.Fatalf("theme kept as %+v, want dark-ansi over the dark it replaced", c.Theme)
	}
	if _, err := os.Stat(filepath.Join(loc.Claude, "themes", "ccdrawu.json")); err != nil {
		t.Fatal("no themes/ccdrawu.json:", err)
	}

	for range 5 {
		press(d, key("k"))
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if got := claudeTheme(t, loc); got != "dark" {
		t.Fatalf("Claude Code's theme = %q after off, want dark back", got)
	}
}

func TestXResetsTheBackground(t *testing.T) {
	d, loc := newTestDrawer(t)
	d.cfg.Theme = config.Theme{Base: "dark-ansi", Background: "#ff0000"}
	d.applyTheme()
	for range 2 { // to preferences
		press(d, key("j"))
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	for range 8 { // past the alias and the theme's 7 options
		press(d, key("j"))
	}
	if d.pages[rowPreferences].Current() != d.background {
		t.Fatalf("the cursor is not on the background:\n%s", frameText(d))
	}
	press(d, key("x"))
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Theme.Background != "" || c.Theme.Base != "dark-ansi" {
		t.Fatalf("theme kept as %+v, want dark-ansi with no background", c.Theme)
	}
	raw, _ := os.ReadFile(filepath.Join(loc.Claude, "themes", "ccdrawu.json"))
	if strings.Contains(string(raw), "composerSidebarBackground") {
		t.Fatalf("themes/ccdrawu.json still sets the background:\n%s", raw)
	}
}

func TestPluginRowInstallsAfterAsking(t *testing.T) {
	t.Setenv("CCDRAWU__MARKETPLACE", "")
	d, _ := newTestDrawer(t)
	f := &claudeFake{}
	d.run = f.run
	d.app.Update(d.checkPlugin()())
	if d.plugin != "not installed" {
		t.Fatalf("plugin row = %q, want not installed", d.plugin)
	}
	for range 2 { // to preferences
		press(d, key("j"))
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // into [2]
	press(d, key("G"))                       // the last field: the plugin
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // the confirm
	if !f.installed && slices.Contains(f.ran, "plugin install ccdrawu --marketplace vulcanshen/ccdrawu --json") {
		t.Fatal("installed before the confirm was answered")
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // install
	if !slices.Contains(f.ran, "plugin install ccdrawu --marketplace vulcanshen/ccdrawu --json") {
		t.Fatalf("ran %q, want the install:\n%s", f.ran, frameText(d))
	}
	if d.plugin != "installed" {
		t.Fatalf("plugin row = %q after installing, want installed", d.plugin)
	}
}

// tdp E6: the splash draws docs/icon.svg
func TestSplashIconIsDocsIcon(t *testing.T) {
	want, err := os.ReadFile("../../docs/icon.svg")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(icon, want) {
		t.Fatal("internal/ui/icon.svg is not docs/icon.svg: copy docs/icon.svg over it")
	}
}
