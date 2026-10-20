package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vulcanshen/ccdrawu/internal/claude"
	"github.com/vulcanshen/ccdrawu/internal/config"
)

// toStatusline goes into the statusline page, on the first field: its on and
// off, at off.
func toStatusline(t *testing.T, d *drawer) {
	t.Helper()
	press(d, key("j"))
	press(d, key("j"))
	if d.detail.Label != rowStatusline {
		t.Fatalf("[2] is %q, want statusline", d.detail.Label)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
}

func claudeSettings(t *testing.T, loc config.Locations) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(loc.Claude, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func setClaudeSettings(t *testing.T, loc config.Locations, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(loc.Claude, "settings.json"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

const mine = "{\n  \"statusLine\": {\"type\": \"command\", \"command\": \"my-line\"},\n  \"theme\": \"dark\"\n}\n"

func TestStatuslineReplacesTheUsersOnlyAfterAsking(t *testing.T) {
	d, loc := newTestDrawer(t)
	setClaudeSettings(t, loc, mine)
	toStatusline(t, d)
	text := frameText(d)
	for _, want := range []string{"[2] STATUSLINE", "ITEMS", "ORDER", "FORMAT", "ctx 42%", "SEPARATOR", "REFRESH     10 seconds"} {
		if !strings.Contains(text, want) {
			t.Fatalf("no %q on the statusline page:\n%s", want, text)
		}
	}
	// The status bar says when the statusline follows (its folder, in a test, leaves no room on screen)
	var got string
	for _, s := range d.status(300) {
		got += s.Text
	}
	if !strings.Contains(got, "the statusline follows within 10 seconds") {
		t.Fatalf("status bar: %q", got)
	}

	press(d, key("j"))                       // on
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // the confirm, naming the status line it replaces
	if text := frameText(d); !strings.Contains(text, "my-line") || !strings.Contains(text, "puts it back") {
		t.Fatalf("the confirm does not say what it replaces:\n%s", text)
	}
	if got := claudeSettings(t, loc); got != mine {
		t.Fatalf("replaced before the confirm was answered:\n%s", got)
	}
	press(d, tea.KeyMsg{Type: tea.KeyEsc}) // no
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Statusline.On || d.statusOn.Selected != 0 || claudeSettings(t, loc) != mine {
		t.Fatalf("a no turned it on: %+v, selected %d", c.Statusline, d.statusOn.Selected)
	}

	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // asked again
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // replace
	if _, ours, _ := claude.CurrentStatusline(loc.Claude); !ours {
		t.Fatalf("Claude Code's status line is not ccdrawu's:\n%s", claudeSettings(t, loc))
	}
	// The one it replaced kept as data, in originals.yaml (tdp E3)
	c, _ := config.ReadConfig(loc.ConfigFile)
	o, _ := config.ReadOriginals(loc.Originals(), config.Config{})
	var before struct{ Command string }
	if err := json.Unmarshal(o.Statusline, &before); !c.Statusline.On || c.Statusline.Before != nil || err != nil || before.Command != "my-line" {
		t.Fatalf("statusline kept as %+v and %+v, want on over my-line", c.Statusline, o)
	}

	press(d, key("k"))                       // off
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // puts mine back as it was
	if got := claudeSettings(t, loc); got != mine {
		t.Fatalf("not put back:\n%s", got)
	}
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Statusline.On || c.Statusline.Before != nil {
		t.Fatalf("statusline kept as %+v after off", c.Statusline)
	}
	if o, err := config.ReadOriginals(loc.Originals(), config.Config{}); o.Statusline != nil || err != nil {
		t.Fatalf("still kept after off: %+v (%v)", o, err)
	}
}

func TestStatuslineOnWithoutAskingWhereThereWasNone(t *testing.T) {
	d, loc := newTestDrawer(t)
	toStatusline(t, d)
	press(d, key("j"))
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if _, ours, _ := claude.CurrentStatusline(loc.Claude); !ours {
		t.Fatalf("not set at once:\n%s", frameText(d))
	}
	if !strings.Contains(claudeSettings(t, loc), `"refreshInterval":10`) {
		t.Fatalf("no refresh:\n%s", claudeSettings(t, loc))
	}
	// The refresh goes into Claude Code's settings at once. It is a number
	// (tdp components/input/number): x does not go in, and 99 is out of range
	toStatusField(t, d, "Refresh")
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	press(d, key("x9"))
	press(d, key("9"))
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Statusline.Refresh != 0 {
		t.Fatalf("refresh %d saved out of range", c.Statusline.Refresh)
	}
	if text := frameText(d); !strings.Contains(text, "Refresh must be 1-60") || !strings.Contains(text, "happens, 1-60") {
		t.Fatalf("the input says nothing of the range:\n%s", text)
	}
	press(d, tea.KeyMsg{Type: tea.KeyBackspace})
	press(d, tea.KeyMsg{Type: tea.KeyBackspace})
	press(d, key("30"))
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Statusline.Refresh != 30 || !strings.Contains(claudeSettings(t, loc), `"refreshInterval":30`) {
		t.Fatalf("refresh %d, settings:\n%s", c.Statusline.Refresh, claudeSettings(t, loc))
	}
}

func TestStatuslineFormatAndItems(t *testing.T) {
	d, loc := newTestDrawer(t)
	if got := d.cfg.Shown(config.Status, d.cat); len(got) != 0 {
		t.Fatalf("on by default in the statusline: %v", got)
	}
	toStatusline(t, d)
	toStatusField(t, d, "Format")
	press(d, key("j")) // short
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Statusline.Format != "short" {
		t.Fatalf("format = %q, want short", c.Statusline.Format)
	}
	toStatusField(t, d, "Color")
	press(d, key("j")) // dimmed
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Statusline.Color != "dimmed" {
		t.Fatalf("color = %q, want dimmed", c.Statusline.Color)
	}
	// The items popup has every item but the pet
	d.cat.Items = append([]config.Item{{ID: petID, Label: "Pet", Name: "Pet", Group: "Pet"}}, d.cat.Items...)
	press(d, key("e"))
	if text := frameText(d); !strings.Contains(text, "EDIT: statusline") || strings.Contains(text, "Pet") {
		t.Fatalf("the statusline's popup:\n%s", text)
	}
}

// A status line set in Claude Code since, in place of ccdrawu's, finds the
// statusline off
func TestStatuslineReplacedSinceReadsOff(t *testing.T) {
	d, loc := newTestDrawer(t)
	d.cfg.Statusline = config.Statusline{On: true, Before: config.Kept("null")}
	if err := config.WriteConfig(loc.ConfigFile, d.cfg); err != nil {
		t.Fatal(err)
	}
	setClaudeSettings(t, loc, mine)
	again, err := newDrawer(loc, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Statusline.On || again.statusOn.Selected != 0 {
		t.Fatalf("on over a status line set since: %+v", c.Statusline)
	}
	if got := claudeSettings(t, loc); got != mine {
		t.Fatalf("changed what the user set:\n%s", got)
	}
}

// toStatusField moves down the statusline page to the field label, on its
// first option when it has options.
func toStatusField(t *testing.T, d *drawer, label string) {
	t.Helper()
	toPageField(t, d, rowStatusline, label)
}

// toPageField moves down the page of row id to the field label, on its first
// option when it has options.
func toPageField(t *testing.T, d *drawer, id, label string) {
	t.Helper()
	for range 40 {
		if f := d.pages[id].Current(); f != nil && f.Label == label {
			return
		}
		press(d, key("j"))
	}
	t.Fatalf("no %s field:\n%s", label, frameText(d))
}

// On when narrow is on too, kept as the config's when; between it and on
// there is nothing to ask or set in Claude Code again
func TestStatuslineOnWhenNarrow(t *testing.T) {
	d, loc := newTestDrawer(t)
	toStatusline(t, d)
	press(d, key("j"))
	press(d, key("j")) // on when narrow
	press(d, tea.KeyMsg{Type: tea.KeyEnter})
	if c, _ := config.ReadConfig(loc.ConfigFile); !c.Statusline.On || c.Statusline.When != config.WhenNarrow {
		t.Fatalf("statusline kept as %+v, want on when narrow", c.Statusline)
	}
	if _, ours, _ := claude.CurrentStatusline(loc.Claude); !ours {
		t.Fatalf("not set in Claude Code:\n%s", claudeSettings(t, loc))
	}
	// Read again, it is on when narrow
	again, err := newDrawer(loc, (&claudeFake{}).run)
	if err != nil {
		t.Fatal(err)
	}
	if again.statusOn.Selected != 2 {
		t.Fatalf("reads as option %d, want on when narrow", again.statusOn.Selected)
	}

	press(d, key("k"))                       // on
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // at once, nothing asked
	if c, _ := config.ReadConfig(loc.ConfigFile); !c.Statusline.On || c.Statusline.When != "" {
		t.Fatalf("statusline kept as %+v, want on", c.Statusline)
	}
	press(d, key("k"))                       // off
	press(d, tea.KeyMsg{Type: tea.KeyEnter}) // and Claude Code has none again
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Statusline.On || c.Statusline.When != "" {
		t.Fatalf("statusline kept as %+v after off", c.Statusline)
	}
	if strings.Contains(claudeSettings(t, loc), "statusLine") {
		t.Fatalf("statusLine left:\n%s", claudeSettings(t, loc))
	}
}
