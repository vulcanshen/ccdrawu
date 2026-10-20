package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var testCatalog = Catalog{Version: 1, Items: []Item{
	{ID: "model", Label: "Model", Name: "Model", Group: "Model", Dock: true, Inline: true},
	{ID: "mode", Label: "Mode", Name: "Permission mode", Group: "Model", Dock: true, Inline: false},
	{ID: "context", Label: "Context", Name: "Context", Group: "Usage", Dock: true, Inline: true},
	{ID: "git", Label: "Git", Name: "Git branch and changes", Group: "Workspace", Dock: false, Inline: true},
}}

// The paths ccdrawu puts together use the platform's separator (\ on
// Windows); a folder a variable names is kept as it is
func TestLocateFollowsE3(t *testing.T) {
	env := func(vars map[string]string) func(string) string { return func(k string) string { return vars[k] } }
	j := filepath.Join
	got := locate(env(nil), "/h", "darwin")
	want := Locations{ConfigFile: j("/h", ".config", "ccdrawu", "config.yaml"), Cache: j("/h", "Library", "Caches", "ccdrawu"),
		Claude: j("/h", ".claude"), OldData: j("/h", ".ccdrawu")}
	if got != want {
		t.Fatalf("defaults on macOS: %+v", got)
	}
	got = locate(env(map[string]string{"XDG_CONFIG_HOME": "/x", "XDG_CACHE_HOME": "/xc", "CCDRAWU__CACHE": ""}), "/h", "linux")
	want = Locations{ConfigFile: j("/x", "ccdrawu", "config.yaml"), Cache: j("/xc", "ccdrawu"), Claude: j("/h", ".claude"), OldData: j("/h", ".ccdrawu")}
	if got != want {
		t.Fatalf("XDG: %+v", got)
	}
	// A relative XDG_CACHE_HOME is not taken, as Go's os.UserCacheDir does not
	if got = locate(env(map[string]string{"XDG_CACHE_HOME": "rel"}), "/h", "linux"); got.Cache != j("/h", ".cache", "ccdrawu") {
		t.Fatalf("relative XDG_CACHE_HOME: %+v", got)
	}
	if got = locate(env(map[string]string{"LOCALAPPDATA": `C:\U\AppData\Local`}), `C:\U`, "windows"); got.Cache != j(`C:\U\AppData\Local`, "ccdrawu", "cache") {
		t.Fatalf("Windows: %+v", got)
	}
	got = locate(env(map[string]string{"XDG_CONFIG_HOME": "/x", "CCDRAWU__CONFIG": "/c", "CCDRAWU__CACHE": "/k", "CCDRAWU__DATA": "/d"}), "/h", "linux")
	want = Locations{ConfigFile: j("/c", "config.yaml"), Cache: "/k", Claude: j("/h", ".claude"), OldData: "/d"}
	if got != want {
		t.Fatalf("CCDRAWU__: %+v", got)
	}
}

// config.yaml holds the config as YAML, in its own order, a list left unset
// as null and the status line kept before as a string; it reads back the same
func TestConfigIsYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	c := Config{Inline: []string{}, Alias: "ccd", Pet: Pet{Species: "dog"},
		Statusline: Statusline{On: true, Items: []string{"model", "cost"}, Refresh: 30, Separator: "#", Before: Kept(`{"type":"command","command":"my-line"}`)}}
	if err := WriteConfig(path, c); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	for _, want := range []string{"dock: null\n", "inline: []\n", "alias: ccd\n", "pet:\n  species: dog\n", "  items:\n    - model\n", "  refresh: 30\n", "  separator: '#'\n"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("no %q in\n%s", want, raw)
		}
	}
	got, err := ReadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dock != nil || got.Inline == nil || got.Alias != "ccd" || got.Pet.Species != "dog" || got.Statusline.Refresh != 30 ||
		got.Statusline.Separator != "#" || !slices.Equal(got.Statusline.Items, c.Statusline.Items) || string(got.Statusline.Before) != string(c.Statusline.Before) {
		t.Fatalf("read back %+v\nfrom\n%s", got, raw)
	}
	// Written by hand, in YAML's own ways
	if err := os.WriteFile(path, []byte("# mine\nalias: 'cc d'\ndock: [model, git]\nstatusline: {on: true}\ninput_tab_width: 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err = ReadConfig(path); err != nil || got.Alias != "cc d" || !slices.Equal(got.Dock, []string{"model", "git"}) || got.InputTabWidth != 2 {
		t.Fatalf("by hand: %+v, %v", got, err)
	}
	// The family's input_tab_width stays through a save (tdp E3)
	if err := WriteConfig(path, got); err != nil {
		t.Fatal(err)
	}
	if raw, _ = os.ReadFile(path); !strings.Contains(string(raw), "input_tab_width: 2\n") {
		t.Errorf("no input_tab_width in\n%s", raw)
	}
}

// 0.1.4 and before kept config.json, and the catalog and the statusline's
// values in ~/.ccdrawu: config.json reads until config.yaml is written, and
// Move takes both to their places now (tdp E3), once
func TestMoveTakesTheOldPlaces(t *testing.T) {
	dir := t.TempDir()
	loc := Locations{ConfigFile: filepath.Join(dir, "config", "config.yaml"), Cache: filepath.Join(dir, "cache", "ccdrawu"), OldData: filepath.Join(dir, ".ccdrawu")}
	old := filepath.Join(dir, "config", "config.json")
	for _, d := range []string{filepath.Dir(old), filepath.Join(loc.OldData, "sessions")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(old, []byte(`{"alias": "ccd", "dock": ["model"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(loc.OldData, "catalog.json"), []byte(`{"version": 1, "items": []}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if c, err := ReadConfig(loc.ConfigFile); err != nil || c.Alias != "ccd" {
		t.Fatalf("config.json not read before the move: %+v, %v", c, err)
	}
	if loc.Catalog() != filepath.Join(loc.OldData, "catalog.json") {
		t.Fatalf("catalog at %s before the move", loc.Catalog())
	}
	moved, err := Move(loc)
	if err != nil || len(moved) != 2 || moved[0] != (Moved{old, loc.ConfigFile}) || moved[1] != (Moved{loc.OldData, loc.Cache}) {
		t.Fatalf("moved %+v, %v", moved, err)
	}
	if fileThere(old) || fileThere(loc.OldData) || !fileThere(filepath.Join(loc.Cache, "sessions")) {
		t.Fatal("the old places are still there, or the sessions did not move")
	}
	if c, err := ReadConfig(loc.ConfigFile); err != nil || c.Alias != "ccd" || !slices.Equal(c.Dock, []string{"model"}) {
		t.Fatalf("config.yaml after the move: %+v, %v", c, err)
	}
	if loc.Catalog() != filepath.Join(loc.Cache, "catalog.json") {
		t.Fatalf("catalog at %s after the move", loc.Catalog())
	}
	if moved, err := Move(loc); err != nil || moved != nil {
		t.Fatalf("moved again: %+v, %v", moved, err)
	}

	// The mod writes the catalog to the cache folder as a session starts,
	// often before ccdrawu first runs: the sessions move,
	// the stale catalog goes, and so does ~/.ccdrawu
	if err := os.MkdirAll(filepath.Join(loc.OldData, "sessions"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{filepath.Join(loc.OldData, "catalog.json"), filepath.Join(loc.OldData, "sessions", "s1.json")} {
		if err := os.WriteFile(f, []byte(`{}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.RemoveAll(filepath.Join(loc.Cache, "sessions")); err != nil {
		t.Fatal(err)
	}
	if moved, err := Move(loc); err != nil || len(moved) != 1 || moved[0] != (Moved{loc.OldData, loc.Cache}) {
		t.Fatalf("moved into a cache folder there already: %+v, %v", moved, err)
	}
	if fileThere(loc.OldData) || !fileThere(filepath.Join(loc.Cache, "sessions", "s1.json")) {
		t.Fatal("~/.ccdrawu still there, or its sessions not moved")
	}
}

func TestConfigDefaultsUntilChanged(t *testing.T) {
	var c Config
	if !c.On(Dock, testCatalog.Items[1]) || c.On(Inline, testCatalog.Items[1]) {
		t.Fatal("an unset placement takes the catalog's defaults")
	}
	c.Set(Inline, testCatalog, "mode", true)
	if want := []string{"model", "mode", "context", "git"}; !slices.Equal(c.Inline, want) {
		t.Fatalf("inline = %v, want the defaults plus mode, in catalog order: %v", c.Inline, want)
	}
	if c.Dock != nil {
		t.Fatal("the other placement stays on its defaults")
	}
}

func TestConfigKeepsUnknownIDsAndWritesAllOff(t *testing.T) {
	c := Config{Dock: []string{"from-newer-mod", "model"}}
	c.Set(Dock, testCatalog, "model", false)
	if want := []string{"from-newer-mod"}; !slices.Equal(c.Dock, want) {
		t.Fatalf("dock = %v, want %v", c.Dock, want)
	}
	c.Set(Dock, testCatalog, "from-newer-mod", false)

	path := filepath.Join(t.TempDir(), "sub", "config.yaml")
	if err := WriteConfig(path, c); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	data, err := jsonFromYAML(data)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	// Every item off is [], which the mod reads as nothing; null would be the defaults
	if list, ok := raw["dock"].([]any); !ok || len(list) != 0 {
		t.Fatalf("dock written as %v, want []", raw["dock"])
	}
	if raw["inline"] != nil {
		t.Fatalf("inline written as %v, want null", raw["inline"])
	}
}

// The drawer's order is the list's: a toggle keeps it, and an item turned on
// goes after the nearest one before it in the catalog
func TestConfigKeepsTheOrder(t *testing.T) {
	c := Config{Dock: []string{"context", "model"}}
	c.Set(Dock, testCatalog, "mode", true)
	if want := []string{"context", "model", "mode"}; !slices.Equal(c.Dock, want) {
		t.Fatalf("dock = %v, want mode after model: %v", c.Dock, want)
	}
	c.Set(Dock, testCatalog, "model", false)
	c.Set(Dock, testCatalog, "model", true)
	if want := []string{"model", "context", "mode"}; !slices.Equal(c.Dock, want) {
		t.Fatalf("dock = %v, want model first, with nothing before it in the catalog: %v", c.Dock, want)
	}
	var ids []string
	for _, it := range c.Shown(Dock, testCatalog) {
		ids = append(ids, it.ID)
	}
	if want := []string{"model", "context", "mode"}; !slices.Equal(ids, want) {
		t.Fatalf("shown = %v, want the list's order %v", ids, want)
	}
	var unset Config
	ids = nil
	for _, it := range unset.Shown(Inline, testCatalog) {
		ids = append(ids, it.ID)
	}
	if want := []string{"model", "context", "git"}; !slices.Equal(ids, want) {
		t.Fatalf("shown = %v, want the catalog's defaults %v", ids, want)
	}
}

// The reorder popup's order goes into the places those items had; the rest
// stay where they were
func TestConfigReordersInPlace(t *testing.T) {
	c := Config{Dock: []string{"model", "from-newer-mod", "context"}}
	c.Reorder(Dock, testCatalog, []string{"context", "model"})
	if want := []string{"context", "from-newer-mod", "model"}; !slices.Equal(c.Dock, want) {
		t.Fatalf("dock = %v, want %v", c.Dock, want)
	}
	var unset Config
	unset.Reorder(Inline, testCatalog, []string{"git", "model", "context"})
	if want := []string{"git", "model", "context"}; !slices.Equal(unset.Inline, want) {
		t.Fatalf("inline = %v, want the defaults written out in the new order %v", unset.Inline, want)
	}
}

// The same names as hooks/format.ts commandNames takes
func TestCleanAliasTakesWhatTheModTakes(t *testing.T) {
	for in, want := range map[string]string{"": "", "  ": "", "ccd": "ccd", " /ccd ": "ccd", "dr-2": "dr-2"} {
		if got, err := CleanAlias(in); err != nil || got != want {
			t.Errorf("CleanAlias(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"ccdrawu", "/ccdrawu", "two words", "-x", "a_b"} {
		if _, err := CleanAlias(in); err == nil {
			t.Errorf("CleanAlias(%q) took it; the mod would not", in)
		}
	}
}

// A toggle rewrites the whole file: the alias and the pet go back as they were
func TestConfigKeepsTheAliasAndPet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("dock: null\nalias: ccd\npet:\n  species: duck\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := ReadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	c.Set(Dock, testCatalog, "git", true)
	if err := WriteConfig(path, c); err != nil {
		t.Fatal(err)
	}
	if c, _ = ReadConfig(path); c.Alias != "ccd" || c.Pet != (Pet{Species: "duck"}) {
		t.Fatalf("alias = %q, pet = %+v after a toggle, want ccd and the duck", c.Alias, c.Pet)
	}
}

// The status line a config keeps to put back comes back as it was written,
// its own spacing and all, however the config is laid out around it; one a
// config.json of ccdrawu 0.1.1 held as JSON reads too
func TestKeptComesBackByteForByte(t *testing.T) {
	dir := t.TempDir()
	path, old := filepath.Join(dir, "config.yaml"), filepath.Join(dir, "config.json")
	for _, before := range []string{
		`{"type": "command", "command": "my-line"}`,
		"{\n    \"type\": \"command\",\n    \"command\": \"npx -y ccstatusline@latest\",\n    \"padding\": 0\n  }",
		"null",
	} {
		if err := WriteConfig(path, Config{Statusline: Statusline{On: true, Before: Kept(before)}}); err != nil {
			t.Fatal(err)
		}
		c, err := ReadConfig(path)
		if err != nil || string(c.Statusline.Before) != before {
			t.Fatalf("kept %q, read back %q (%v)", before, c.Statusline.Before, err)
		}
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(old, []byte("{\"statusline\": {\"on\": true, \"before\": {\n  \"type\": \"command\", \"command\": \"my-line\"\n}}}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if c, err := ReadConfig(path); err != nil || string(c.Statusline.Before) != `{"type":"command","command":"my-line"}` {
		t.Fatalf("0.1.1 object read as %q (%v)", c.Statusline.Before, err)
	}
	if err := os.WriteFile(old, []byte(`{"statusline": {"on": true, "before": null}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if c, err := ReadConfig(path); err != nil || len(c.Statusline.Before) != 0 {
		t.Fatalf("0.1.1 null read as %q (%v)", c.Statusline.Before, err)
	}
}
