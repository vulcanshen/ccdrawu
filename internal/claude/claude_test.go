package claude

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// fake answers each command by its first words, and keeps what ran
type fake struct {
	answers map[string]string
	ran     []string
}

func (f *fake) run(name string, args ...string) ([]byte, error) {
	line := strings.Join(append([]string{name}, args...), " ")
	f.ran = append(f.ran, line)
	for prefix, answer := range f.answers {
		if strings.HasPrefix(line, prefix) {
			return []byte(answer), nil
		}
	}
	return nil, errors.New("no answer for " + line)
}

const listed = `[{"id": "other@x", "enabled": true}, {"id": "ccdrawu@ccdrawu", "enabled": true}]`

func TestInstallTakesTheMarketplaceItIsGiven(t *testing.T) {
	f := &fake{answers: map[string]string{
		"claude plugin install": `{"command":"install","outcome":"ok","message":"Installed ccdrawu"}`,
	}}
	msg, err := Install(f.run, "/dev/ccdrawu")
	if err != nil || msg != "Installed ccdrawu" {
		t.Fatalf("Install = %q, %v", msg, err)
	}
	if want := "claude plugin install ccdrawu --marketplace /dev/ccdrawu --json"; !slices.Equal(f.ran, []string{want}) {
		t.Fatalf("ran %q, want %q", f.ran, want)
	}
}

func TestInstallSaysWhatClaudeCodeRefused(t *testing.T) {
	f := &fake{answers: map[string]string{
		"claude plugin install": "fetching…\n" + `{"outcome":"error","message":"Marketplace file not found"}`,
	}}
	if _, err := Install(f.run, Source); err == nil || err.Error() != "Marketplace file not found" {
		t.Fatalf("err = %v, want Claude Code's message", err)
	}
}

func TestUninstallRemovesThePluginThenItsMarketplace(t *testing.T) {
	f := &fake{answers: map[string]string{
		"claude plugin list":               listed,
		"claude plugin uninstall":          `{"outcome":"ok","message":"Uninstalled ccdrawu@ccdrawu"}`,
		"claude plugin marketplace remove": `{"outcome":"ok","message":"Removed"}`,
	}}
	msg, ok, err := Uninstall(f.run)
	if err != nil || !ok || msg != "Uninstalled ccdrawu@ccdrawu" {
		t.Fatalf("Uninstall = %q, %v, %v", msg, ok, err)
	}
	want := []string{
		"claude plugin list --json",
		"claude plugin uninstall ccdrawu@ccdrawu --json",
		"claude plugin marketplace remove ccdrawu --json",
	}
	if !slices.Equal(f.ran, want) {
		t.Fatalf("ran %q, want %q", f.ran, want)
	}
}

func TestUninstallWithNothingThere(t *testing.T) {
	f := &fake{answers: map[string]string{"claude plugin list": `[{"id": "other@x", "enabled": true}]`}}
	if _, ok, err := Uninstall(f.run); ok || err != nil {
		t.Fatalf("ok = %v, err = %v; want not there, no error", ok, err)
	}
	if len(f.ran) != 1 {
		t.Fatalf("ran %q, want the list alone", f.ran)
	}
}

func TestInstalledFindsThePlugin(t *testing.T) {
	f := &fake{answers: map[string]string{"claude plugin list": listed}}
	if id, enabled, err := Installed(f.run); id != "ccdrawu@ccdrawu" || !enabled || err != nil {
		t.Fatalf("Installed = %q, %v, %v", id, enabled, err)
	}
}
