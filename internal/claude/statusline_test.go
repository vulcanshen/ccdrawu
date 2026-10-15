package claude

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeSettings(t *testing.T, text string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func readBack(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestStatuslineReplacesAnotherAndPutsItBack(t *testing.T) {
	const theirs = `{
  "env": {
    "A": "1"
  },
  "statusLine": {
    "type": "command",
    "command": "npx -y ccstatusline@latest",
    "padding": 0
  },
  "theme": "dark"
}
`
	dir := writeSettings(t, theirs)
	if cmd, ours, err := CurrentStatusline(dir); err != nil || ours || cmd != "npx -y ccstatusline@latest" {
		t.Fatalf("current: %q %v %v", cmd, ours, err)
	}
	before, switched, err := ApplyStatusline(dir, 10)
	if err != nil || !switched {
		t.Fatalf("apply: %v %v", switched, err)
	}
	got := readBack(t, dir)
	if !strings.Contains(got, `"statusLine": {"type":"command","command":"ccdrawu statusline","refreshInterval":10},`) ||
		!strings.Contains(got, `"theme": "dark"`) || !strings.HasPrefix(got, "{\n  \"env\": {\n    \"A\": \"1\"\n  },\n") {
		t.Fatalf("applied:\n%s", got)
	}
	if _, ours, _ := CurrentStatusline(dir); !ours {
		t.Fatal("ccdrawu's is not the one set")
	}
	// Set again, it takes the refresh alone and has nothing to put back
	again, switched, err := ApplyStatusline(dir, 30)
	if err != nil || switched || again != nil || !strings.Contains(readBack(t, dir), `"refreshInterval":30`) {
		t.Fatalf("again: %s %v %v", again, switched, err)
	}
	// The one it replaced comes back as it was written, byte for byte
	if err := RestoreStatusline(dir, before); err != nil {
		t.Fatal(err)
	}
	if got := readBack(t, dir); got != theirs {
		t.Fatalf("restored:\n%s", got)
	}
	// Owner-only permissions survive the writes (Windows has no 0600 to keep)
	if info, _ := os.Stat(filepath.Join(dir, "settings.json")); runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("mode %v", info.Mode().Perm())
	}
}

func TestStatuslineWhereThereWasNone(t *testing.T) {
	const none = "{\n  \"theme\": \"dark\"\n}\n"
	dir := writeSettings(t, none)
	before, switched, err := ApplyStatusline(dir, 10)
	if err != nil || !switched || string(before) != "null" {
		t.Fatalf("apply: %s %v %v", before, switched, err)
	}
	if err := RestoreStatusline(dir, before); err != nil {
		t.Fatal(err)
	}
	if got := readBack(t, dir); got != none {
		t.Fatalf("restored:\n%q", got)
	}
	// No settings.json at all: one is written with the statusLine alone
	empty := t.TempDir()
	if _, _, err := ApplyStatusline(empty, 10); err != nil {
		t.Fatal(err)
	}
	if _, ours, _ := CurrentStatusline(empty); !ours {
		t.Fatal("not set in a new settings.json")
	}
}

func TestStatuslineSetSinceIsLeftAlone(t *testing.T) {
	dir := writeSettings(t, "{}\n")
	before, _, err := ApplyStatusline(dir, 10)
	if err != nil {
		t.Fatal(err)
	}
	const since = "{\n  \"statusLine\": {\"type\": \"command\", \"command\": \"mine\"}\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(since), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RestoreStatusline(dir, before); err != nil {
		t.Fatal(err)
	}
	if got := readBack(t, dir); got != since {
		t.Fatalf("changed what the user set:\n%s", got)
	}
}
