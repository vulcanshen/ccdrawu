package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const settings = `{
  "env": {
    "A": "1"
  },
  "theme": "custom:dark-ansi-crust",
  "enabledPlugins": {
    "x@y": true
  }
}
`

func TestSetStringChangesThatValueAlone(t *testing.T) {
	got, err := setString([]byte(settings), "theme", "custom:ccdrawu")
	if err != nil {
		t.Fatal(err)
	}
	want := `{
  "env": {
    "A": "1"
  },
  "theme": "custom:ccdrawu",
  "enabledPlugins": {
    "x@y": true
  }
}
`
	if string(got) != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}

func TestSetStringPutsAMissingKeyFirst(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"{\n    \"env\": {}\n}\n", "{\n    \"theme\": \"dark\",\n    \"env\": {}\n}\n"},
		{"{}\n", "{\n  \"theme\": \"dark\"\n}\n"},
		{`{"env": {}}`, `{"theme": "dark", "env": {}}`},
	} {
		got, err := setString([]byte(c.in), "theme", "dark")
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != c.want {
			t.Fatalf("setString(%q) = %q, want %q", c.in, got, c.want)
		}
		var v map[string]any
		if err := json.Unmarshal(got, &v); err != nil {
			t.Fatalf("%q is no longer JSON: %v", got, err)
		}
	}
}

func TestRemoveKeyTakesItsLineAndComma(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{settings, "{\n  \"env\": {\n    \"A\": \"1\"\n  },\n  \"enabledPlugins\": {\n    \"x@y\": true\n  }\n}\n"},
		{"{\n  \"env\": {},\n  \"theme\": \"dark\"\n}\n", "{\n  \"env\": {}\n}\n"},
		{"{\n  \"theme\": \"dark\"\n}\n", "{}\n"},
		{`{"a": 1, "theme": "dark", "b": 2}`, `{"a": 1, "b": 2}`},
	} {
		got, err := removeKey([]byte(c.in), "theme")
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != c.want {
			t.Fatalf("removeKey(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestThemeIsAppliedAndPutBack(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(settings), 0o600); err != nil {
		t.Fatal(err)
	}
	before, switched, err := ApplyTheme(dir, "dark-ansi", "#11111b")
	if err != nil || !switched || before != "custom:dark-ansi-crust" {
		t.Fatalf("ApplyTheme = %q, %v, %v", before, switched, err)
	}
	data, _ := os.ReadFile(path)
	if stringAt(data, "theme") != ThemeSetting {
		t.Fatalf("theme = %q, want %q", stringAt(data, "theme"), ThemeSetting)
	}
	var theme struct {
		Base      string            `json:"base"`
		Overrides map[string]string `json:"overrides"`
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "themes", "ccdrawu.json"))
	if err := json.Unmarshal(raw, &theme); err != nil || theme.Base != "dark-ansi" || theme.Overrides["composerSidebarBackground"] != "#11111b" {
		t.Fatalf("themes/ccdrawu.json = %s (%v)", raw, err)
	}
	// Once selected, a change of base or background is no new switch
	if _, switched, _ := ApplyTheme(dir, "dark", ""); switched {
		t.Fatal("switched again while ccdrawu's theme was already the one")
	}
	// Windows has no such mode to keep: a file there reads 0666 unless read-only
	if info, _ := os.Stat(path); runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Fatalf("settings.json is now %v, want its 0600 kept", info.Mode().Perm())
	}

	if err := RestoreTheme(dir, before); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(path); string(data) != settings {
		t.Fatalf("settings.json after restoring:\n%s\nwant it as it was", data)
	}
	if _, err := os.Stat(filepath.Join(dir, "themes", "ccdrawu.json")); err == nil {
		t.Fatal("themes/ccdrawu.json is still there")
	}
}

func TestRestoreLeavesAThemeChosenSince(t *testing.T) {
	dir := t.TempDir()
	chosen := "{\n  \"theme\": \"light\"\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(chosen), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RestoreTheme(dir, "dark"); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(filepath.Join(dir, "settings.json")); string(data) != chosen {
		t.Fatalf("settings.json = %s, want the theme chosen in /theme kept", data)
	}
}

func TestDirFollowsClaudeConfigDir(t *testing.T) {
	env := func(v string) func(string) string { return func(string) string { return v } }
	if got := Dir(env(""), "/h"); got != filepath.Join("/h", ".claude") {
		t.Fatalf("Dir = %q", got)
	}
	if got := Dir(env("/c"), "/h"); got != "/c" {
		t.Fatalf("Dir with CLAUDE_CONFIG_DIR = %q", got)
	}
}
