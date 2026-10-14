package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

var testCatalog = Catalog{Version: 1, Items: []Item{
	{ID: "model", Label: "Model", Name: "Model", Group: "Model", Dock: true, Inline: true},
	{ID: "mode", Label: "Mode", Name: "Permission mode", Group: "Model", Dock: true, Inline: false},
	{ID: "context", Label: "Context", Name: "Context", Group: "Usage", Dock: true, Inline: true},
	{ID: "git", Label: "Git", Name: "Git branch and changes", Group: "Workspace", Dock: false, Inline: true},
}}

func TestLocateFollowsE3(t *testing.T) {
	env := func(vars map[string]string) func(string) string { return func(k string) string { return vars[k] } }
	got := locate(env(nil), "/h")
	if got != (Locations{Data: "/h/.ccdrawu", ConfigFile: "/h/.config/ccdrawu/config.json", Claude: "/h/.claude"}) {
		t.Fatalf("defaults: %+v", got)
	}
	got = locate(env(map[string]string{"XDG_CONFIG_HOME": "/x", "CCDRAWU__DATA": ""}), "/h")
	if got != (Locations{Data: "/h/.ccdrawu", ConfigFile: "/x/ccdrawu/config.json", Claude: "/h/.claude"}) {
		t.Fatalf("XDG: %+v", got)
	}
	got = locate(env(map[string]string{"XDG_CONFIG_HOME": "/x", "CCDRAWU__CONFIG": "/c", "CCDRAWU__DATA": "/d"}), "/h")
	if got != (Locations{Data: "/d", ConfigFile: "/c/config.json", Claude: "/h/.claude"}) {
		t.Fatalf("CCDRAWU__: %+v", got)
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

	path := filepath.Join(t.TempDir(), "sub", "config.json")
	if err := WriteConfig(path, c); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
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

// A toggle rewrites the whole file: the alias goes back as it was
func TestConfigKeepsTheAlias(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"dock": null, "alias": "ccd"}`), 0o644); err != nil {
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
	if c, _ = ReadConfig(path); c.Alias != "ccd" {
		t.Fatalf("alias = %q after a toggle, want ccd", c.Alias)
	}
}
