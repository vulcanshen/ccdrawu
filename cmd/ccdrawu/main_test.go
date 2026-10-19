package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vulcanshen/ccdrawu/internal/claude"
	"github.com/vulcanshen/ccdrawu/internal/config"
)

// uninstall claude puts back the user's own theme and status line, which it
// does whether or not the plugin was still there to take out
func TestPutBackGivesClaudeCodeItsOwnBack(t *testing.T) {
	dir := t.TempDir()
	loc := config.Locations{Cache: filepath.Join(dir, "data"), ConfigFile: filepath.Join(dir, "config", "config.yaml"), Claude: filepath.Join(dir, "claude")}
	const theirs = "{\n  \"statusLine\": {\"type\": \"command\", \"command\": \"npx -y ccstatusline@latest\", \"padding\": 0},\n  \"theme\": \"dark\"\n}\n"
	settings := filepath.Join(loc.Claude, "settings.json")
	if err := os.MkdirAll(loc.Claude, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settings, []byte(theirs), 0o600); err != nil {
		t.Fatal(err)
	}

	var cfg config.Config
	before, _, err := claude.ApplyTheme(loc.Claude, "dark-ansi", "#11111b")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Theme.Base, cfg.Theme.Before = "dark-ansi", before
	line, _, err := claude.ApplyStatusline(loc.Claude, 10)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Statusline.On, cfg.Statusline.Before = true, config.Kept(line)
	if err := config.WriteConfig(loc.ConfigFile, cfg); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := putBack(loc, &out); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(settings); string(got) != theirs {
		t.Fatalf("settings.json not put back:\n%s", got)
	}
	for _, want := range []string{"its own theme back", "its own status line back"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("no %q in %q", want, out.String())
		}
	}
	if c, _ := config.ReadConfig(loc.ConfigFile); c.Theme.Base != "" || c.Statusline.On || c.Statusline.Before != nil {
		t.Fatalf("config still has ccdrawu's: %+v %+v", c.Theme, c.Statusline)
	}

	// Again, there is nothing of ccdrawu's left to put back
	out.Reset()
	if err := putBack(loc, &out); err != nil || out.Len() != 0 {
		t.Fatalf("again: %q %v", out.String(), err)
	}
}
