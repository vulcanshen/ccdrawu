package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
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

// [2] of sidebar and inline has Items, the ones shown, and Order, the same in
// the drawer's order; Enter on Items, or e, opens every item in one checkbox
// popup, where Space checks, Esc drops it all and Enter applies it all
func TestEditPicksTheItemsInOnePopup(t *testing.T) {
	d, loc := newTestDrawer(t)
	// [1] starts on sidebar; Enter goes into [2] (tdp K3, details)
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if d.app.Screen().Focused() != d.detail {
		t.Fatalf("Enter on [1] left the focus there:\n%s", frameText(d))
	}
	text := frameText(d)
	for _, want := range []string{"ITEMS", "ORDER", "Model, Mode, Context", "Enter:edit"} {
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

	toPageField(t, d, rowSidebar, "Items")
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // on Items: the popup of every item
	if text := frameText(d); !strings.Contains(text, "EDIT: sidebar") || !strings.Contains(text, "Mode     Permission mode") {
		t.Fatalf("Enter on Items did not open the popup of every item with their keys:\n%s", text)
	}
	press(d, space) // Model off
	press(d, tea.KeyMsg{Type: tea.KeyEsc})
	if _, err := os.Stat(loc.ConfigFile); err == nil {
		t.Fatal("Esc kept what was checked")
	}

	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	press(d, space) // Model off
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if got, want := dock(), []string{"mode", "context"}; !slices.Equal(got, want) {
		t.Fatalf("dock = %v, want %v", got, want)
	}

	press(d, key("e")) // the same popup
	for range 3 {      // Mode, Context, Git
		press(d, key("j"))
	}
	press(d, space) // Git on
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if got, want := dock(), []string{"mode", "context", "git"}; !slices.Equal(got, want) {
		t.Fatalf("dock = %v after e, want %v", got, want)
	}
	if text := frameText(d); !strings.Contains(text, "Mode, Context, Git") {
		t.Fatalf("[2] does not show what was applied:\n%s", text)
	}
}

// The drawer's on and off heads each of its pages, its own there, on until it
// is turned off: off is written as drawer.sidebarOff or drawer.inlineOff, and
// on leaves it out again; 0.1.1's drawer.off reads as off in both
func TestDrawerTurnsOffAndOn(t *testing.T) {
	d, loc := newTestDrawer(t)
	sidebar, inline := d.drawerOn[rowSidebar], d.drawerOn[rowInline]
	if d.pages[rowSidebar].Fields[0] != sidebar || d.pages[rowInline].Fields[0] != inline || sidebar == inline || sidebar.Selected != 1 || inline.Selected != 1 {
		t.Fatalf("the drawer is not on atop each page: %+v", d.pages[rowSidebar].Fields[0])
	}
	if text := frameText(d); !strings.Contains(text, "SETTING") || !strings.Contains(text, "VALUE") {
		t.Fatalf("[2] has no titles over its columns:\n%s", text)
	}
	raw := func() string {
		t.Helper()
		data, err := os.ReadFile(loc.ConfigFile)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	// Off above the prompt alone: beside the transcript it stays on
	press(d, key("j"))
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // into inline, on off
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // off
	if c, _ := config.ReadConfig(loc.ConfigFile); !c.Drawer.InlineOff || c.Drawer.SidebarOff || !strings.Contains(raw(), `"inlineOff": true`) {
		t.Fatalf("not written off above the prompt alone:\n%s", raw())
	}
	if sidebar.Selected != 1 {
		t.Fatal("the sidebar's turned off with inline")
	}
	press(d, key("j"))                       // on
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // on again
	if strings.Contains(raw(), "drawer") {
		t.Fatalf("on is still written:\n%s", raw())
	}

	if err := os.WriteFile(loc.ConfigFile, []byte(`{"drawer": {"off": true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	again, err := newDrawer(loc, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	if again.drawerOn[rowSidebar].Selected != 0 || again.drawerOn[rowInline].Selected != 0 {
		t.Fatal("0.1.1's drawer off reads on")
	}
}

// Order is the drawer's order, the config's list as it is; Enter opens it in
// a reorder popup, where D picks an item up, j|k move it, Enter drops it and
// Enter again writes the order back. With one item there is nothing to order.
func TestOrderIsReordered(t *testing.T) {
	d, loc := newTestDrawer(t)
	d.cfg.Dock = []string{"context", "model"}
	fields := d.pages[rowSidebar].Fields
	if f := fields[1]; f.Label != "Items" || f.Value() != "Model, Context" {
		t.Fatalf("%q shows %q, want Items showing Model, Context", f.Label, f.Value())
	}
	if f := fields[2]; f.Label != "Order" || f.Value() != "Context, Model" {
		t.Fatalf("%q shows %q, want Order showing Context, Model", f.Label, f.Value())
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	toPageField(t, d, rowSidebar, "Order")
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if text := frameText(d); !strings.Contains(text, "ORDER: sidebar") || !strings.Contains(text, "Context") {
		t.Fatalf("Enter on Order did not open the reorder popup:\n%s", text)
	}
	press(d, key("D"))                       // Context
	press(d, key("j"))                       // below Model
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // dropped
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // written back
	if c, _ := config.ReadConfig(loc.ConfigFile); !slices.Equal(c.Dock, []string{"model", "context"}) {
		t.Fatalf("dock = %v, want model, context", c.Dock)
	}
	if f := fields[2]; f.Value() != "Model, Context" {
		t.Fatalf("Order shows %q, want Model, Context", f.Value())
	}

	d.cfg.Dock = []string{"model"}
	if !fields[2].Disabled() {
		t.Fatal("Order is on with one item to order")
	}
}

func TestNoCatalogSaysWhatToDo(t *testing.T) {
	dir := t.TempDir()
	d, err := newDrawer(config.Locations{Data: filepath.Join(dir, "data"), ConfigFile: filepath.Join(dir, "config.json")}, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	d.app.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	// tdp components/layout/screen, the empty state: a fact, and a hint that
	// names a key
	if !strings.Contains(terminu.StripStyles(strings.Join(d.app.Frame(), "\n")), "Press [Enter] on Plugin") {
		t.Fatal("an empty panel says nothing of the key that installs the plugin")
	}
}

// tdp components/layout/panel, groups and details: [1] in groups, the cursor
// skipping their titles; [2] is the row under it, named after it
func TestDetailFollowsTheSidebar(t *testing.T) {
	d, _ := newTestDrawer(t)
	text := frameText(d)
	for _, want := range []string{"[1] PREFERENCES", "DRAWER", "STATUSLINE", "SETTINGS", "[2] SIDEBAR", "Model, Mode"} {
		if !strings.Contains(text, want) {
			t.Fatalf("no %q in\n%s", want, text)
		}
	}
	d.app.Update(key("j"))
	if d.detail.Label != rowInline {
		t.Fatalf("[2] is %q after j, want inline", d.detail.Label)
	}
	d.app.Update(key("j")) // over the STATUSLINE title
	if d.detail.Label != rowStatusline {
		t.Fatalf("[2] is %q after j j, want statusline", d.detail.Label)
	}
	d.app.Update(key("j")) // over the SETTINGS title
	text = frameText(d)
	if d.detail.Label != rowPreferences || !strings.Contains(text, "[2] PREFERENCES") {
		t.Fatalf("[2] is %q after j j j, want preferences:\n%s", d.detail.Label, text)
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
	press(d, key("j"))
	press(d, key("j")) // sidebar, inline, statusline, preferences; its first field is the alias
	press(d, tea.KeyMsg{Type: tea.KeyTab})
	edit := func(v string) {
		press(d, tea.KeyMsg{Type: tea.KeyEnter})
		press(d, key(v))
		press(d, tea.KeyMsg{Type: tea.KeyEnter})
	}
	// tdp K3: the value it opened on is not written back, though typed again
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	press(d, key("a"))
	press(d, tea.KeyMsg{Type: tea.KeyBackspace})
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if _, err := os.Stat(loc.ConfigFile); err == nil {
		t.Fatal("the alias it opened on was written back")
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
	for range 3 { // sidebar, inline, statusline, preferences
		press(d, key("j"))
	}
	if d.detail.Label != rowPreferences {
		t.Fatalf("[2] is %q, want preferences", d.detail.Label)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	press(d, key("j")) // past the alias
	// More than five themes: a select (tdp components/input/radio), on off
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if text := frameText(d); !strings.Contains(text, "CLAUDE THEME") || !strings.Contains(text, "light-ansi") {
		t.Fatalf("Enter on the theme did not open a select of them:\n%s", text)
	}
	for range 5 { // dark, light, dark-daltonized, light-daltonized, dark-ansi
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

	if text := frameText(d); !strings.Contains(text, "dark-ansi") {
		t.Fatalf("the theme row does not read dark-ansi:\n%s", text)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // the select again, on dark-ansi
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
	for range 3 { // to preferences
		press(d, key("j"))
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	for range 2 { // past the alias and the theme
		press(d, key("j"))
	}
	if d.pages[rowPreferences].Current() != d.background {
		t.Fatalf("the cursor is not on the background:\n%s", frameText(d))
	}
	// tdp components/input/color-picker: a swatch, then the hex
	if text := frameText(d); !strings.Contains(text, "███ #FF0000") {
		t.Fatalf("the background is not a swatch and its hex:\n%s", text)
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

// toField goes into preferences and down to the field label, on its first
// option when it has options.
func toField(t *testing.T, d *drawer, label string) {
	t.Helper()
	if d.detail.Label != rowPreferences {
		for range 3 {
			press(d, key("j"))
		}
		press(d, tea.KeyMsg{Type: tea.KeyEnter})
	}
	for range 30 {
		if d.pages[rowPreferences].Current().Label == label {
			return
		}
		press(d, key("j"))
	}
	t.Fatalf("no %s field:\n%s", label, frameText(d))
}

func TestBarIsSquaresUntilTheLineIsPicked(t *testing.T) {
	d, loc := newTestDrawer(t)
	toField(t, d, "Bar")
	if d.pages[rowPreferences].Current().Selected != 0 {
		t.Fatal("the bar starts on the squares")
	}
	if text := frameText(d); !strings.Contains(text, "\uF445\uF445\uF445\uF445\uF51D") || !strings.Contains(text, "━━━━━━━━") {
		t.Fatalf("the options do not show their bars:\n%s", text)
	}
	press(d, key("j"))
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Bar != "line" {
		t.Fatalf("bar = %q, want line", c.Bar)
	}
}

func TestPetIsPicked(t *testing.T) {
	d, loc := newTestDrawer(t)
	toField(t, d, "Pet") // on the cat
	press(d, key("j"))
	press(d, key("j"))
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Pet.Species != "rabbit" {
		t.Fatalf("pet = %+v, want the rabbit", c.Pet)
	}
}

// tdp components/layout/panel, one value per row: a value set by hand in
// config.json that ccdrawu does not take turns its label Red
func TestWrongValuesAreMarked(t *testing.T) {
	d, loc := newTestDrawer(t)
	if err := os.MkdirAll(filepath.Dir(loc.ConfigFile), 0o755); err != nil {
		t.Fatal(err)
	}
	wrong := `{"alias": "a b!", "theme": {"base": "dark-ansi", "background": "zzz"}, "statusline": {"refresh": 999, "separator": "\u001b"}}`
	if err := os.WriteFile(loc.ConfigFile, []byte(wrong), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := newDrawer(loc, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	marked := map[string]bool{}
	for _, f := range append(d.pages[rowPreferences].Fields, d.statuslineFields()...) {
		if f.Err != nil && f.Err() {
			marked[f.Label] = true
		}
	}
	for _, label := range []string{"Alias", "Drawer background", "Refresh", "Separator"} {
		if !marked[label] {
			t.Errorf("%s is not marked wrong: %v", label, marked)
		}
	}
	if marked["Plugin"] {
		t.Error("the plugin is marked wrong before it is looked for")
	}
	d.pluginChecked(pluginChecked{err: errors.New("no claude")})
	if f := d.pages[rowPreferences].Fields[len(d.pages[rowPreferences].Fields)-1]; f.Label != "Plugin" || !f.Err() {
		t.Error("the plugin row is not marked wrong when it can't tell")
	}
}

// A change that can't be written is not shown as if it were: the config, and
// the radio, go back to the file
func TestUnsavedChangeGoesBack(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a folder can't be made read-only there")
	}
	d, loc := newTestDrawer(t)
	toField(t, d, "Pet") // on the cat
	dir := filepath.Dir(loc.ConfigFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	press(d, key("j"))
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // the dog, not saved
	if text := frameText(d); !strings.Contains(text, "CAN'T SAVE") {
		t.Fatalf("no error popup:\n%s", text)
	}
	if f := d.pages[rowPreferences].Current(); d.cfg.Pet.Species != "" || f.Selected != 0 {
		t.Fatalf("pet %q, the radio on %d, want both back on the cat", d.cfg.Pet.Species, f.Selected)
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
	for range 3 { // to preferences
		press(d, key("j"))
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // into [2]
	press(d, key("G"))                       // the last field: the plugin
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // the confirm
	if !f.installed && slices.Contains(f.ran, "plugin install ccdrawu --marketplace https://github.com/vulcanshen/ccdrawu.git --json") {
		t.Fatal("installed before the confirm was answered")
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // install
	if !slices.Contains(f.ran, "plugin install ccdrawu --marketplace https://github.com/vulcanshen/ccdrawu.git --json") {
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
