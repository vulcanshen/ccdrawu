package statusline

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/vulcanshen/ccdrawu/internal/config"
)

var cat = config.Catalog{Items: []config.Item{
	{ID: "pet", Label: "Pet", Short: "pet"},
	{ID: "model", Label: "Model", Short: "mdl", Statusline: true},
	{ID: "context", Label: "Context", Short: "ctx", Statusline: true},
	{ID: "cost", Label: "Cost", Short: "cost", Statusline: true},
	{ID: "git", Label: "Git", Short: "git"},
}}

var items = map[string][]Seg{
	"model":   {{Text: "Opus 5.5", Color: lavenderColor, Bold: true}},
	"context": {{Text: "42%", Color: "#a6e3a1"}},
	"cost":    {{Text: "$1.20", Color: textColor}},
	"git":     {{Text: "main", Color: textColor}, {Text: "  +12", Color: "#a6e3a1"}, {Text: " -3", Color: "#f38ba8"}},
}

func TestLineFormats(t *testing.T) {
	cases := []struct {
		format, want string
	}{
		{"", "Model Opus 5.5 | Context 42% | Cost $1.20"},
		{"label", "Model Opus 5.5 | Context 42% | Cost $1.20"},
		{"short", "mdl Opus 5.5 | ctx 42% | cost $1.20"},
		{"value", "Opus 5.5 | 42% | $1.20"},
	}
	for _, c := range cases {
		cfg := config.Config{Statusline: config.Statusline{Format: c.format}}
		if got := ansi.Strip(Line(cfg, cat, items, 0)); got != c.want {
			t.Errorf("format %q: %q, want %q", c.format, got, c.want)
		}
	}
}

func TestLineOrderSeparatorAndGaps(t *testing.T) {
	// The config's order and separator; an item with nothing to show (context here) is left out
	cfg := config.Config{Statusline: config.Statusline{Items: []string{"git", "context", "model"}, Separator: "·"}}
	got := ansi.Strip(Line(cfg, cat, map[string][]Seg{"git": items["git"], "model": items["model"]}, 0))
	if want := "Git main  +12 -3 · Model Opus 5.5"; got != want {
		t.Errorf("%q, want %q", got, want)
	}
	// A short label the catalog lacks (from an older mod) falls back on the label
	cfg = config.Config{Statusline: config.Statusline{Items: []string{"model"}, Format: "short"}}
	old := config.Catalog{Items: []config.Item{{ID: "model", Label: "Model"}}}
	if got := ansi.Strip(Line(cfg, old, items, 0)); got != "Model Opus 5.5" {
		t.Errorf("no short label: %q", got)
	}
}

func TestLineColors(t *testing.T) {
	line := Line(config.Config{Statusline: config.Statusline{Items: []string{"model", "context"}}}, cat, items, 0)
	for _, want := range []string{
		"\x1b[38;2;137;180;250mModel \x1b[0m",          // the label in blue
		"\x1b[38;2;180;190;254m\x1b[1mOpus 5.5\x1b[0m", // the value as the mod colored it
		"\x1b[38;2;69;71;90m | \x1b[0m",                // the separator in surface
	} {
		if !strings.Contains(line, want) {
			t.Errorf("%q has no %q", line, want)
		}
	}
	if got := paint("x", "not a color", false); got != "x\x1b[0m" {
		t.Errorf("unreadable color: %q", got)
	}
}

func TestLineDimmedHasNoColorOfItsOwn(t *testing.T) {
	cfg := config.Config{Statusline: config.Statusline{Items: []string{"model", "context", "cost", "git"}, Color: "dimmed"}}
	for _, columns := range []int{0, 20} {
		line := Line(cfg, cat, items, columns)
		if strings.Contains(line, "\x1b") {
			t.Errorf("dimmed, %d columns: %q", columns, line)
		}
	}
	if got := Line(cfg, cat, items, 20); !strings.HasSuffix(got, "…") || ansi.StringWidth(got) != 20 {
		t.Errorf("dimmed, cut: %q", got)
	}
}

func TestLineIsCutToTheWidth(t *testing.T) {
	cfg := config.Config{Statusline: config.Statusline{Items: []string{"model", "context", "cost", "git"}}}
	full := Line(cfg, cat, items, 0)
	cut := Line(cfg, cat, items, 20)
	if w := ansi.StringWidth(cut); w != 20 {
		t.Errorf("cut to %d cells, want 20: %q", w, ansi.Strip(cut))
	}
	if !strings.HasSuffix(ansi.Strip(cut), "…") || ansi.StringWidth(full) <= 20 {
		t.Errorf("cut line %q of %q", ansi.Strip(cut), ansi.Strip(full))
	}
	if got := Line(cfg, cat, items, 500); got != full {
		t.Errorf("a line that fits is cut: %q", ansi.Strip(got))
	}
}

func setup(t *testing.T) (config.Locations, string) {
	t.Helper()
	dir := t.TempDir()
	loc := config.Locations{Data: filepath.Join(dir, "data"), ConfigFile: filepath.Join(dir, "config", "config.json")}
	if err := config.WriteConfig(loc.ConfigFile, config.Config{Statusline: config.Statusline{On: true}}); err != nil {
		t.Fatal(err)
	}
	sessions := filepath.Join(loc.Data, "sessions")
	if err := os.MkdirAll(sessions, 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(cat)
	if err := os.WriteFile(filepath.Join(loc.Data, "catalog.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return loc, sessions
}

func run(t *testing.T, loc config.Locations, input string, columns int, now time.Time) string {
	t.Helper()
	var out bytes.Buffer
	if err := Run(strings.NewReader(input), &out, loc, columns, now); err != nil {
		t.Fatal(err)
	}
	return ansi.Strip(out.String())
}

func TestRunReadsTheSessionsFile(t *testing.T) {
	loc, sessions := setup(t)
	data, _ := json.Marshal(map[string]any{"version": 1, "items": items})
	if err := os.WriteFile(filepath.Join(sessions, "abc.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	input := `{"session_id": "abc", "model": {"id": "claude-sonnet-5-5", "display_name": "Sonnet 5.5"}, "cost": {"total_cost_usd": 9}}`
	if got := run(t, loc, input, 0, time.Now()); got != "Model Opus 5.5 | Context 42% | Cost $1.20\n" {
		t.Errorf("%q", got)
	}
	// Claude Code keeps two cells on either side: the line is cut to the width less those
	if got := run(t, loc, input, 24, time.Now()); ansi.StringWidth(strings.TrimSuffix(got, "\n")) != 20 {
		t.Errorf("cut for 24 columns: %q", got)
	}
}

func TestRunWithoutTheFileShowsWhatTheInputHas(t *testing.T) {
	loc, _ := setup(t)
	input := `{"session_id": "no-file", "model": {"id": "claude-opus-5-5", "display_name": "Opus 5.5"}, "context_window": {"used_percentage": 71.6}, "cost": {"total_cost_usd": 0.5}}`
	if got := run(t, loc, input, 0, time.Now()); got != "Model claude-opus-5-5 | Context 72% | Cost $0.50\n" {
		t.Errorf("%q", got)
	}
	// Before the first reply the input has no percentage: that item is left out
	if got := run(t, loc, `{"session_id": "no-file", "model": {"id": "claude-opus-5-5"}}`, 0, time.Now()); got != "Model claude-opus-5-5\n" {
		t.Errorf("%q", got)
	}
	// An id that would reach out of the folder is not read
	if items, _ := sessionItems(filepath.Join(loc.Data, "sessions"), "../catalog"); items != nil {
		t.Error("read a file outside sessions")
	}
	if got := fromInput(Input{})["context"]; got != nil {
		t.Errorf("context from nothing: %v", got)
	}
}

func TestRunWithNoCatalogOrConfig(t *testing.T) {
	dir := t.TempDir()
	loc := config.Locations{Data: filepath.Join(dir, "data"), ConfigFile: filepath.Join(dir, "none", "config.json")}
	input := `{"model": {"id": "claude-opus-5-5"}, "cost": {"total_cost_usd": 0}}`
	if got := run(t, loc, input, 0, time.Now()); got != "Model claude-opus-5-5 | Cost $0.00\n" {
		t.Errorf("%q", got)
	}
}

func TestRunRemovesFilesOfSessionsLongQuiet(t *testing.T) {
	loc, sessions := setup(t)
	now := time.Now()
	for name, age := range map[string]time.Duration{"old.json": keep + time.Hour, "new.json": time.Hour, "keep.txt": keep * 2} {
		path := filepath.Join(sessions, name)
		if err := os.WriteFile(path, []byte(`{"items": {}}`), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, now.Add(-age), now.Add(-age)); err != nil {
			t.Fatal(err)
		}
	}
	run(t, loc, `{"session_id": "new"}`, 0, now)
	for name, kept := range map[string]bool{"old.json": false, "new.json": true, "keep.txt": true} {
		if _, err := os.Stat(filepath.Join(sessions, name)); (err == nil) != kept {
			t.Errorf("%s kept: %v, want %v", name, err == nil, kept)
		}
	}
}

// On when narrow, the line is empty while the drawer is beside the transcript
// or the terminal is wide enough for it, and the items otherwise; on, it shows
// either way
func TestRunOnWhenNarrow(t *testing.T) {
	loc, sessions := setup(t)
	input := `{"session_id": "abc"}`
	const all = "Model Opus 5.5 | Context 42% | Cost $1.20\n"
	for _, c := range []struct {
		when    string
		docked  bool
		columns int
		want    string
	}{
		{config.WhenNarrow, true, 120, "\n"},
		{config.WhenNarrow, false, 120, all},
		{config.WhenNarrow, false, 143, all},
		{config.WhenNarrow, false, 144, "\n"}, // wide: the drawer closed, or not drawn yet
		{config.WhenNarrow, false, 0, all},    // no width known
		{"", true, 170, all},
	} {
		if err := config.WriteConfig(loc.ConfigFile, config.Config{Statusline: config.Statusline{On: true, When: c.when}}); err != nil {
			t.Fatal(err)
		}
		data, _ := json.Marshal(map[string]any{"version": 1, "docked": c.docked, "items": items})
		if err := os.WriteFile(filepath.Join(sessions, "abc.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
		if got := run(t, loc, input, c.columns, time.Now()); got != c.want {
			t.Errorf("when %q, docked %v, %d columns: %q, want %q", c.when, c.docked, c.columns, got, c.want)
		}
	}
}

// A line hidden, or with nothing to show, is a reset alone, which Claude Code
// draws as no row at all, where an empty line would leave a blank one
func TestRunHidesTheRow(t *testing.T) {
	loc, sessions := setup(t)
	data, _ := json.Marshal(map[string]any{"version": 1, "docked": true, "items": items})
	if err := os.WriteFile(filepath.Join(sessions, "abc.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, s := range []config.Statusline{
		{On: true, When: config.WhenNarrow}, // docked
		{On: true, Items: []string{}},       // nothing picked
	} {
		if err := config.WriteConfig(loc.ConfigFile, config.Config{Statusline: s}); err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		if err := Run(strings.NewReader(`{"session_id": "abc"}`), &out, loc, 100, time.Now()); err != nil {
			t.Fatal(err)
		}
		if got := out.String(); got != hidden+"\n" {
			t.Errorf("%+v: %q, want the reset alone", s, got)
		}
	}
}
