package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/vulcanshen/terminu-go"
)

// testLocations is a home of its own with a catalog of a few items and Claude
// Code's settings on its own theme.
func testLocations(t *testing.T) Locations {
	t.Helper()
	dir := t.TempDir()
	loc := Locations{ConfigFile: filepath.Join(dir, "config", "config.yaml"), Cache: filepath.Join(dir, "cache"), Claude: filepath.Join(dir, "claude"), Data: filepath.Join(dir, ".ccdrawu")}
	for _, d := range []string{filepath.Dir(loc.ConfigFile), loc.Cache, loc.Claude} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	cat, _ := json.Marshal(Catalog{Items: []Item{{ID: "pet"}, {ID: "model"}, {ID: "git"}}})
	if err := os.WriteFile(loc.Catalog(), cat, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(loc.Claude, "settings.json"), []byte("{\n  \"theme\": \"dark\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return loc
}

const misfitting = `alias: "a b!"
dock: [model, nope, 3]
inline: model
pet: {species: dragon}
bar: wavy
drawer: {sidebarOff: yes please}
theme: {background: zzz}
statusline:
  on: true
  when: sometimes
  items: [git]
  format: loud
  color: red
  separator: "\x1b"
  refresh: ten
`

// tdp E3-28: values ccdrawu can't use are not a file it can't read: each is
// its default, said with its key, what the file has and what is used now; an
// item the catalog does not have is taken out of its list
func TestValuesThatDoNotFitAreTheirDefaults(t *testing.T) {
	loc := testLocations(t)
	if err := os.WriteFile(loc.ConfigFile, []byte(misfitting), 0o644); err != nil {
		t.Fatal(err)
	}
	c, misfits, err := readFit(loc.ConfigFile, []string{"pet", "model", "git"})
	if err != nil {
		t.Fatal(err)
	}
	want := []terminu.Misfit{
		{Key: `dock "nope"`, Was: "nope", Now: ""},
		{Key: "dock 3", Was: "3", Now: ""},
		{Key: "inline", Was: "model", Now: "default"},
		{Key: "alias", Was: "a b!", Now: "none"},
		{Key: "pet.species", Was: "dragon", Now: "cat"},
		{Key: "bar", Was: "wavy", Now: "squares"},
		{Key: "drawer.sidebarOff", Was: "yes please", Now: "false"},
		{Key: "theme.background", Was: "zzz", Now: "none"},
		{Key: "statusline.when", Was: "sometimes", Now: "always"},
		{Key: "statusline.format", Was: "loud", Now: "label"},
		{Key: "statusline.color", Was: "red", Now: "highlighted"},
		{Key: "statusline.separator", Was: "\x1b", Now: "|"},
		{Key: "statusline.refresh", Was: "ten", Now: "10"},
	}
	if !slices.Equal(misfits, want) {
		t.Fatalf("misfits\n%+v\nwant\n%+v", misfits, want)
	}
	if !slices.Equal(c.Dock, []string{"model"}) || c.Inline != nil || c.Alias != "" || c.Pet.Species != "" || c.Bar != "" ||
		c.Drawer.SidebarOff || c.Theme.Background != "" || !c.Statusline.On || !slices.Equal(c.Statusline.Items, []string{"git"}) ||
		c.Statusline.When != "" || c.Statusline.Format != "" || c.Statusline.Color != "" || c.Statusline.Separator != "" || c.Statusline.Refresh != 0 {
		t.Fatalf("read as %+v", c)
	}
	// Past 60 seconds, below 0, and a theme that is not a built-in one
	if err := os.WriteFile(loc.ConfigFile, []byte("theme: {base: neon}\nstatusline: {refresh: 61}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, misfits, _ = readFit(loc.ConfigFile, nil); len(misfits) != 2 || misfits[0].Now != "off" || misfits[1].Was != "61" {
		t.Fatalf("misfits %+v", misfits)
	}
	// With no catalog there is nothing to know the items by: none is taken out
	if c, misfits, _ = readFit(loc.ConfigFile, nil); len(misfits) != 2 {
		t.Fatalf("misfits %+v", misfits)
	}
	if err := os.WriteFile(loc.ConfigFile, []byte("dock: [model, nope]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if c, misfits, _ = readFit(loc.ConfigFile, nil); len(misfits) != 0 || !slices.Equal(c.Dock, []string{"model", "nope"}) {
		t.Fatalf("with no catalog: %+v, %+v", c, misfits)
	}
	// Not YAML, or not a mapping: a file that can't be read (tdp E3-17)
	for _, text := range []string{"alias: [\n", "- a list\n"} {
		if err := os.WriteFile(loc.ConfigFile, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadConfig(loc.ConfigFile); err == nil {
			t.Errorf("%q read", text)
		}
	}
}

// tdp E3-29, E3-30: at a start the file is copied to config.yaml.bak (.bak.1
// with one there) and written again with the defaults; the next start finds
// nothing to say
func TestSettleKeepsTheFileAsideAndWritesItAgain(t *testing.T) {
	loc := testLocations(t)
	if err := os.WriteFile(loc.ConfigFile, []byte(misfitting), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(loc.ConfigFile+".bak", []byte("an older one"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, s, err := Settle(loc)
	if err != nil || s.Err != nil || len(s.Misfits) != 13 || s.Copy != loc.ConfigFile+".bak.1" {
		t.Fatalf("settled %+v, %v", s, err)
	}
	if kept, _ := os.ReadFile(s.Copy); string(kept) != misfitting {
		t.Fatalf("kept aside:\n%s", kept)
	}
	again, err := ReadConfig(loc.ConfigFile)
	if err != nil || !slices.Equal(again.Dock, c.Dock) || !again.Statusline.On {
		t.Fatalf("written again as %+v, %v", again, err)
	}
	if _, s, _ = Settle(loc); len(s.Misfits) != 0 || s.Copy != "" {
		t.Fatalf("a second start: %+v", s)
	}
	// Said with no screen, a line on its own (tdp E3-31)
	three := []terminu.Misfit{{Key: "bar", Was: "wavy", Now: "squares"}, {Key: "pet.species", Was: "dragon", Now: "cat"}, {Key: "alias", Was: "a b!", Now: "none"}}
	lines := (Settled{Misfits: three, Copy: loc.ConfigFile + ".bak"}).Lines(loc.ConfigFile)
	if len(lines) != 1 || lines[0] != "3 values in config.yaml did not fit: defaults used, yours kept in config.yaml.bak" {
		t.Fatalf("lines %q", lines)
	}
}

// A theme that does not fit is off, and off gives Claude Code its own back, as
// Claude theme off does in the TUI
func TestSettleGivesClaudeCodeItsThemeBack(t *testing.T) {
	loc := testLocations(t)
	if err := os.WriteFile(filepath.Join(loc.Claude, "settings.json"), []byte("{\n  \"theme\": \"custom:ccdrawu\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteOriginals(loc.Originals(), Originals{Theme: "light"}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(loc.ConfigFile, []byte("theme: {base: neon}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, s, err := Settle(loc); err != nil || s.Err != nil {
		t.Fatalf("%+v, %v", s, err)
	}
	settings, _ := os.ReadFile(filepath.Join(loc.Claude, "settings.json"))
	orig, _ := ReadOriginals(loc.Originals(), Config{})
	if !strings.Contains(string(settings), `"light"`) || orig.Theme != "" {
		t.Fatalf("settings %s, originals %+v", settings, orig)
	}
}

// tdp E3-32: the copy can't be made: the file is not written again, the
// defaults for this run only, and why it says
func TestSettleWithNoCopyLeavesTheFile(t *testing.T) {
	if runtime.GOOS == "windows" || os.Getuid() == 0 {
		t.Skip("a folder that can't be written to")
	}
	loc := testLocations(t)
	if err := os.WriteFile(loc.ConfigFile, []byte("bar: wavy\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(loc.ConfigFile)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	c, s, err := Settle(loc)
	if err != nil || c.Bar != "" || s.Copy != "" || s.Err == nil || !strings.Contains(s.Err.Error(), "config.yaml left as it was") {
		t.Fatalf("%+v, %+v, %v", c, s, err)
	}
	if data, _ := os.ReadFile(loc.ConfigFile); string(data) != "bar: wavy\n" || fileThere(loc.ConfigFile+".bak") {
		t.Fatalf("the file touched:\n%s", data)
	}
	if lines := s.Lines(loc.ConfigFile); len(lines) != 1 || !strings.Contains(lines[0], "could not be made") {
		t.Fatalf("lines %q", lines)
	}
}

// The config.json of 0.1.4 and before moves as it is, values that do not fit
// and all, for Settle to say
func TestMoveKeepsValuesThatDoNotFit(t *testing.T) {
	loc := testLocations(t)
	if err := os.WriteFile(filepath.Join(filepath.Dir(loc.ConfigFile), "config.json"), []byte(`{"bar": "wavy", "theme": {"before": "light"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Move(loc); err != nil {
		t.Fatal(err)
	}
	if _, s, err := Settle(loc); err != nil || len(s.Misfits) != 1 || s.Misfits[0].Key != "bar" {
		t.Fatalf("after the move: %+v, %v", s, err)
	}
	if orig, _ := ReadOriginals(loc.Originals(), Config{}); orig.Theme != "light" {
		t.Fatalf("the theme it kept: %+v", orig)
	}
}
