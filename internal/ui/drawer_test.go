package ui

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vulcanshen/ccdrawu/internal/config"
	terminu "github.com/vulcanshen/terminu-go"
)

var testCatalog = config.Catalog{Version: 1, Items: []config.Item{
	{ID: "model", Label: "Model", Name: "Model", Group: "Model", Dock: true, Inline: true},
	{ID: "mode", Label: "Mode", Name: "Permission mode", Group: "Model", Dock: true, Inline: false},
	{ID: "context", Label: "Context", Name: "Context", Group: "Usage", Dock: true, Inline: true},
	{ID: "git", Label: "Git", Name: "Git branch and changes", Group: "Workspace", Dock: false, Inline: true},
}}

func newTestDrawer(t *testing.T) (*drawer, config.Locations) {
	t.Helper()
	dir := t.TempDir()
	loc := config.Locations{Data: filepath.Join(dir, "data"), ConfigFile: filepath.Join(dir, "config", "config.json")}
	if err := os.MkdirAll(loc.Data, 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(testCatalog)
	if err := os.WriteFile(filepath.Join(loc.Data, "catalog.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := newDrawer(loc)
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

func TestEnterTogglesAndSavesAtOnce(t *testing.T) {
	d, loc := newTestDrawer(t)
	// [1] starts on sidebar; Enter goes into [2] (tdp K3, details), whose cursor is on Model
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if d.app.Screen().Focused() != d.detail {
		t.Fatalf("Enter on [1] left the focus there:\n%s", frameText(d))
	}
	d.app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	c, err := config.ReadConfig(loc.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"mode", "context"}; !slices.Equal(c.Dock, want) {
		t.Fatalf("dock = %v, want %v", c.Dock, want)
	}
	if c.Inline != nil {
		t.Fatalf("inline = %v, want it left on its defaults", c.Inline)
	}
}

func TestNoCatalogSaysWhatToDo(t *testing.T) {
	dir := t.TempDir()
	d, err := newDrawer(config.Locations{Data: filepath.Join(dir, "data"), ConfigFile: filepath.Join(dir, "config.json")})
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
	for _, want := range []string{"[1] PREFERENCES", "DRAWER", "SETTINGS", "[2] SIDEBAR", "Permission mode"} {
		if !strings.Contains(text, want) {
			t.Fatalf("no %q in\n%s", want, text)
		}
	}
	d.app.Update(key("j"))
	if d.detail.Label != rowInline {
		t.Fatalf("[2] is %q after j, want inline", d.detail.Label)
	}
	d.app.Update(key("j")) // over the SETTINGS title
	if text := frameText(d); d.detail.Label != rowAlias || !strings.Contains(text, "[2] ALIAS") || !strings.Contains(text, "none") {
		t.Fatalf("[2] is %q after j j, want alias, none set:\n%s", d.detail.Label, text)
	}
}

func TestAliasIsCheckedAndSaved(t *testing.T) {
	d, loc := newTestDrawer(t)
	press(d, key("G")) // the last row: alias
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
