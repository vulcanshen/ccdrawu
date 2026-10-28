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
	loc := config.Locations{Cache: filepath.Join(dir, "data"), ConfigFile: filepath.Join(dir, "config", "config.yaml"), Claude: filepath.Join(dir, "claude"), Data: filepath.Join(dir, ".ccdrawu")}
	for _, d := range []string{loc.Cache, loc.Claude} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	data, _ := json.Marshal(testCatalog)
	if err := os.WriteFile(filepath.Join(loc.Cache, "catalog.json"), data, 0o644); err != nil {
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
	if c, _ := config.ReadConfig(loc.ConfigFile); !c.Drawer.InlineOff || c.Drawer.SidebarOff || !strings.Contains(raw(), "inlineOff: true") {
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

// tdp E3, a place changed: what 0.1.4 and before kept elsewhere moves on the
// first start, and an Info toast says from where to where
func TestOldPlacesMoveOnTheFirstStart(t *testing.T) {
	dir := t.TempDir()
	loc := config.Locations{ConfigFile: filepath.Join(dir, "config", "config.yaml"), Cache: filepath.Join(dir, "cache"),
		Claude: filepath.Join(dir, "claude"), Data: filepath.Join(dir, "old")}
	for _, d := range []string{filepath.Dir(loc.ConfigFile), loc.Data} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(loc.ConfigFile), "config.json"), []byte(`{"alias": "ccd", "theme": {"base": "dark-ansi", "before": "dark"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(testCatalog)
	if err := os.WriteFile(filepath.Join(loc.Data, "catalog.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := newDrawer(loc, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	d.app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	run(d, d.app.Start()) // what moved when it opened, said once the screen is up
	moved, err := config.Move(loc)
	if err != nil || moved != nil {
		t.Fatalf("moved again: %+v, %v", moved, err)
	}
	text := frameText(d)
	if _, err := os.Stat(loc.ConfigFile); err != nil || !strings.Contains(text, "Moved") || d.cfg.Alias != "ccd" || len(d.cat.Items) == 0 {
		t.Fatalf("no toast of the move, or the config or the catalog not read after it:\n%s", text)
	}
	if _, err := os.Stat(filepath.Join(loc.Cache, "catalog.json")); err != nil {
		t.Fatal("the catalog did not move to the cache:", err)
	}
	// The theme it replaced, data now, in the data folder (tdp E3)
	if d.orig.Theme != "dark" || d.cfg.Theme.Before != "" {
		t.Fatalf("the theme kept as %+v, the config's as %+v", d.orig, d.cfg.Theme)
	}
	press(d, key("!"))
	if text := frameText(d); !strings.Contains(text, "theme.before") || !strings.Contains(text, "originals.yaml") {
		t.Fatalf("the App Log does not say theme.before moved to originals.yaml:\n%s", text)
	}
}

// tdp E3, the new place has it already: a config.json left beside
// config.yaml is named, with it, in an error popup each start
func TestConfigJSONBesideConfigYAMLSaysSo(t *testing.T) {
	_, loc := newTestDrawer(t)
	if err := config.WriteConfig(loc.ConfigFile, config.Config{Alias: "new"}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(loc.ConfigFile), "config.json"), []byte(`{"alias": "old"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := newDrawer(loc, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	d.app.Update(tea.WindowSizeMsg{Width: 200, Height: 40})
	run(d, d.app.Start())
	if text := frameText(d); !strings.Contains(text, "CAN'T MOVE") || !strings.Contains(text, "ccdrawu reads config.yaml") || strings.Contains(text, `\n`) || d.cfg.Alias != "new" {
		t.Fatalf("no error popup for config.json beside config.yaml, or config.yaml not read:\n%s", text)
	}
}

// tdp E3, can't be read: originals.yaml, data, is asked about as the config
// is; moved aside, nothing is kept
func TestOriginalsCantBeRead(t *testing.T) {
	_, loc := newTestDrawer(t)
	if err := os.MkdirAll(loc.Data, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(loc.Originals(), []byte("theme: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := newDrawer(loc, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	d.app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	run(d, d.app.Start())
	if text := frameText(d); !strings.Contains(text, "MOVE ASIDE") || !strings.Contains(text, "originals.yaml.bak") {
		t.Fatalf("no MOVE ASIDE for originals.yaml:\n%s", text)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if _, err := os.Stat(loc.Originals() + ".bak"); err != nil || !d.orig.Equal(config.Originals{}) {
		t.Fatalf("originals.yaml not moved aside (%v), or still kept: %+v", err, d.orig)
	}
}

// tdp E3-18: Enter on MOVE ASIDE starts as on the first start, where nothing
// moved while the question was there: a config.json beside the config.yaml
// moved aside is taken where config.yaml is
func TestMovedAsideStartsAsTheFirstStart(t *testing.T) {
	_, loc := newTestDrawer(t)
	if err := os.MkdirAll(filepath.Dir(loc.ConfigFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(loc.ConfigFile, []byte("alias: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := filepath.Join(filepath.Dir(loc.ConfigFile), "config.json")
	if err := os.WriteFile(old, []byte(`{"alias": "old"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := newDrawer(loc, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	d.app.Update(tea.WindowSizeMsg{Width: 200, Height: 40})
	run(d, d.app.Start())
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	c, _ := config.ReadConfig(loc.ConfigFile)
	if text := frameText(d); fileThere(old) || c.Alias != "old" || d.cfg.Alias != "old" || !strings.Contains(text, "Moved") || strings.Contains(text, "CAN'T MOVE") {
		t.Fatalf("config.json not taken where config.yaml is, as on a first start (%+v):\n%s", c, text)
	}
}

// tdp E3-18: with originals.yaml moved aside, a status line set in Claude
// Code since turns the statusline off, as a first start does
func TestMovedAsideOriginalsStartAsTheFirstStart(t *testing.T) {
	d, loc := newTestDrawer(t)
	d.cfg.Statusline = config.Statusline{On: true}
	if err := config.WriteConfig(loc.ConfigFile, d.cfg); err != nil {
		t.Fatal(err)
	}
	setClaudeSettings(t, loc, mine)
	if err := os.MkdirAll(loc.Data, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(loc.Originals(), []byte("statusline: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	again, err := newDrawer(loc, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	again.app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	run(again, again.app.Start())
	press(again, tea.KeyMsg{Type: tea.KeyEnter})
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Statusline.On || again.statusOn.Selected != 0 {
		t.Fatalf("on over a status line set since: %+v", c.Statusline)
	}
	if got := claudeSettings(t, loc); got != mine {
		t.Fatalf("changed what the user set:\n%s", got)
	}
}

func fileThere(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// tdp E3, can't be read: a config that can't be read is asked about first
// and moved aside on Enter; a catalog, a cache, is thrown away, in the App Log
func TestWhatCantBeRead(t *testing.T) {
	d, loc := newTestDrawer(t)
	if err := os.MkdirAll(filepath.Dir(loc.ConfigFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(loc.ConfigFile, []byte("alias: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(loc.Cache, "catalog.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := newDrawer(loc, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	d.app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	run(d, d.app.Start())
	// The question, not "can't be read": after a temp folder's long path, as
	// on Windows, that wraps
	if text := frameText(d); !strings.Contains(text, "MOVE ASIDE") || !strings.Contains(text, "Move it to config.yaml.bak") {
		t.Fatalf("no MOVE ASIDE for the config:\n%s", text)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if _, err := os.Stat(loc.ConfigFile + ".bak"); err != nil {
		t.Fatal("the config was not moved aside:", err)
	}
	if _, err := os.Stat(filepath.Join(loc.Cache, "catalog.json")); err == nil {
		t.Fatal("the catalog that can't be read is still there")
	}
	press(d, key("!"))
	if text := frameText(d); !strings.Contains(text, "WARN Threw away") {
		t.Fatalf("the App Log says nothing of the catalog:\n%s", text)
	}
}

func TestNoCatalogSaysWhatToDo(t *testing.T) {
	dir := t.TempDir()
	d, err := newDrawer(config.Locations{Cache: filepath.Join(dir, "data"), ConfigFile: filepath.Join(dir, "config.yaml"), Data: filepath.Join(dir, ".ccdrawu")}, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	d.app.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	// tdp screen.8, the empty state: a fact, and a hint that
	// names a key
	if !strings.Contains(terminu.StripStyles(strings.Join(d.app.Frame(), "\n")), "Press [Enter] on Plugin") {
		t.Fatal("an empty panel says nothing of the key that installs the plugin")
	}
}

// tdp panel.11 and panel.13, groups and details: [1] in groups, the cursor
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

// tdp M3 and tdp K6: what Enter does on [2]'s value is an item operation, in
// the Space menu and in ?, named as the hint names it, and does the same
func TestEnterOnAValueIsInTheMenu(t *testing.T) {
	d, loc := newTestDrawer(t)
	toField(t, d, "Alias")
	press(d, space)
	if text := frameText(d); !strings.Contains(text, "[Enter] Edit") {
		t.Fatalf("no [Enter] Edit in the Space menu on the alias:\n%s", text)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEsc})
	press(d, key("?"))
	if text := frameText(d); !strings.Contains(text, "Edit — this value") {
		t.Fatalf("? does not list Enter:\n%s", text)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEsc})
	toField(t, d, "Pet") // on the cat
	press(d, key("j"))   // the dog
	press(d, space)
	if text := frameText(d); !strings.Contains(text, "[Enter] Choose") {
		t.Fatalf("no [Enter] Choose in the Space menu on a radio:\n%s", text)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // runs it from the menu
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Pet.Species != "dog" {
		t.Fatalf("pet = %q after Choose from the menu, want the dog", c.Pet.Species)
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
	// More than five themes: a select (tdp radio.1), on off
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
	// The dark it replaced kept as data, in originals.yaml (tdp E3)
	c, _ := config.ReadConfig(loc.ConfigFile)
	o, err := config.ReadOriginals(loc.Originals(), config.Config{})
	if c.Theme.Base != "dark-ansi" || c.Theme.Before != "" || o.Theme != "dark" || err != nil {
		t.Fatalf("theme kept as %+v and %+v (%v), want dark-ansi over the dark it replaced", c.Theme, o, err)
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
	// Nothing kept, nothing left in the data folder
	if _, err := os.Stat(loc.Data); err == nil {
		t.Fatalf("%s still there after off", loc.Data)
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
	// tdp color-picker.4: a swatch, then the hex
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

// tdp panel.12, one value per row: a value set by hand in
// config.yaml that ccdrawu does not take turns its label Red
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
	// An entry starts with a capital, as a toast (tdp note-78)
	d.app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	press(d, key("!"))
	if text := frameText(d); !strings.Contains(text, "WARN Can't tell from claude plugin list") {
		t.Errorf("the App Log does not say it can't tell:\n%s", text)
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

	// From an input, the error opens over it, which stays open with what was
	// typed (tdp popup.7)
	press(d, tea.KeyMsg{Type: tea.KeyEsc})
	toField(t, d, "Alias")
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	press(d, key("ccd"))
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if text := frameText(d); !strings.Contains(text, "FAILED") || d.cfg.Alias != "" {
		t.Fatalf("no error over the alias, or the alias taken as saved (%q):\n%s", d.cfg.Alias, text)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEsc}) // the error
	if text := frameText(d); !strings.Contains(text, "One more name for /ccdrawu") || !strings.Contains(text, "ccd") {
		t.Fatalf("the alias input is gone, or what was typed in it:\n%s", text)
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
	if text := frameText(d); !strings.Contains(text, "Claude Code: Installed ccdrawu") {
		t.Fatalf("no toast in Claude Code's own word:\n%s", text)
	}
}

// tdp components-32: several lines from claude are not put whole into the
// error popup: the line saying why goes there, all of it into the App Log
func TestClaudesLinesGoToTheAppLog(t *testing.T) {
	d, _ := newTestDrawer(t)
	d.app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	run(d, d.pluginDone(pluginDone{err: errors.New("exit status 1: Error: no network\n    at fetch (cli.js:1)")}))
	if text := frameText(d); !strings.Contains(text, "Error: no network") || strings.Contains(text, "cli.js") || strings.Contains(text, `\n`) {
		t.Fatalf("the error popup does not hold the one line saying why:\n%s", text)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	press(d, key("!"))
	if text := frameText(d); !strings.Contains(text, "cli.js:1") {
		t.Fatalf("the App Log does not hold all claude said:\n%s", text)
	}
}

// tdp components-31: what could not move, a row each in the error popup
func TestWhatCantMoveARowEach(t *testing.T) {
	d, _ := newTestDrawer(t)
	d.app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	run(d, d.sayMoved(nil, errors.Join(errors.New("The first can't move."), errors.New("The second can't move."))))
	first, second := -1, -1
	for i, line := range strings.Split(frameText(d), "\n") {
		if strings.Contains(line, "The first") {
			first = i
		}
		if strings.Contains(line, "The second") {
			second = i
		}
	}
	if first < 0 || second != first+1 {
		t.Fatalf("not a row each, one under the other:\n%s", frameText(d))
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
