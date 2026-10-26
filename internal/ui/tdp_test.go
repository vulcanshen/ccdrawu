package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/vulcanshen/ccdrawu/internal/config"
	terminu "github.com/vulcanshen/terminu-go"
	"github.com/vulcanshen/terminu-go/tdptest"
)

// TestTDP launches the TUI as for the first time, in a sandbox, on many sizes
// and icon widths, and walks what every family app has (tdptest): its
// catalog seeded, its cache and Claude Code's folder in the sandbox, and
// `claude plugin list` answered by the fake, so nothing outside is touched.
func TestTDP(t *testing.T) {
	tdptest.Check(t, app(t, func(string) {}))
}

// TestTDPTricky gives values that have broken screens (tdptest.Tricky) to
// what ccdrawu draws from a config set by hand: the alias, the separator and
// the drawer background. Every frame is checked for what reaches the
// terminal (tdp components.5).
func TestTDPTricky(t *testing.T) {
	for _, v := range []string{"\x1b[31mred\x1b[0m", "\x1b]52;c;aGk=\x07", "line\nbreak", "漢字かな한국어 wide"} {
		tdptest.Check(t, app(t, func(home string) {
			cfg := config.Config{Alias: v, Theme: config.Theme{Base: "dark", Background: v}, Statusline: config.Statusline{Separator: v}}
			if err := config.WriteConfig(filepath.Join(home, "config", "config.yaml"), cfg); err != nil {
				t.Fatal(err)
			}
		}))
	}
}

// app is ccdrawu for tdptest, seed putting more in its sandbox.
func app(t *testing.T, seed func(home string)) tdptest.App {
	return tdptest.App{
		Name: "ccdrawu",
		New: func() *terminu.App {
			loc, err := config.UserLocations()
			if err != nil {
				t.Fatal(err)
			}
			d, err := newDrawer(loc, (&claudeFake{}).run)
			if err != nil {
				t.Fatal(err)
			}
			return d.app
		},
		Env: func(home string) map[string]string {
			return map[string]string{
				"CCDRAWU__CONFIG":   filepath.Join(home, "config"),
				"CCDRAWU__CACHE":    filepath.Join(home, "cache"),
				"CLAUDE_CONFIG_DIR": filepath.Join(home, ".claude"),
			}
		},
		Seed: func(home string) {
			data, _ := json.Marshal(testCatalog)
			if err := os.MkdirAll(filepath.Join(home, "cache"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(home, "cache", "catalog.json"), data, 0o644); err != nil {
				t.Fatal(err)
			}
			seed(home)
		},
	}
}
