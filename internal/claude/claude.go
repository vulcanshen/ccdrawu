// Package claude puts the ccdrawu plugin into Claude Code and takes it out,
// through Claude Code's own `claude plugin` commands.
package claude

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Plugin is the plugin's name; Source is the marketplace it comes from,
// unless CCDRAWU__MARKETPLACE names another (a folder, while developing).
const (
	Plugin = "ccdrawu"
	Source = "vulcanshen/ccdrawu"
)

// SourceFrom is where install takes the plugin from: CCDRAWU__MARKETPLACE,
// else Source.
func SourceFrom(getenv func(string) string) string {
	if s := getenv("CCDRAWU__MARKETPLACE"); s != "" {
		return s
	}
	return Source
}

// Run runs a command and returns what it printed on stdout.
type Run func(name string, args ...string) ([]byte, error)

// Exec runs the command for real; a failure carries what it said on stderr.
func Exec(name string, args ...string) ([]byte, error) {
	if _, err := exec.LookPath(name); err != nil {
		return nil, errors.New("Claude Code's claude command is not on PATH: install Claude Code first")
	}
	out, err := exec.Command(name, args...).Output()
	var exit *exec.ExitError
	if errors.As(err, &exit) && len(exit.Stderr) > 0 {
		err = fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exit.Stderr)))
	}
	return out, err
}

// Install installs the plugin for the user from source, adding its
// marketplace first when Claude Code does not have it. Installing it again is
// no change. It returns Claude Code's own word on what it did.
func Install(run Run, source string) (string, error) {
	return result(run("claude", "plugin", "install", Plugin, "--marketplace", source, "--json"))
}

// Uninstall removes the plugin and then its marketplace. ok is false when
// Claude Code does not have the plugin.
func Uninstall(run Run) (msg string, ok bool, err error) {
	id, _, err := Installed(run)
	if err != nil || id == "" {
		return "", false, err
	}
	if msg, err = result(run("claude", "plugin", "uninstall", id, "--json")); err != nil {
		return "", true, err
	}
	// The marketplace is ours alone: the part of the id after @
	if _, market, found := strings.Cut(id, "@"); found {
		if _, err := result(run("claude", "plugin", "marketplace", "remove", market, "--json")); err != nil {
			return msg, true, err
		}
	}
	return msg, true, nil
}

// Installed reports the plugin's id in Claude Code ("ccdrawu@ccdrawu"), ""
// when it is not there, and whether it is enabled.
func Installed(run Run) (id string, enabled bool, err error) {
	out, err := run("claude", "plugin", "list", "--json")
	if err != nil {
		return "", false, err
	}
	var plugins []struct {
		ID      string `json:"id"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.Unmarshal(out, &plugins); err != nil {
		return "", false, fmt.Errorf("claude plugin list: %w", err)
	}
	for _, p := range plugins {
		if strings.HasPrefix(p.ID, Plugin+"@") {
			return p.ID, p.Enabled, nil
		}
	}
	return "", false, nil
}

// result reads the one JSON line a `claude plugin ... --json` prints last.
func result(out []byte, err error) (string, error) {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var r struct {
		Outcome string `json:"outcome"`
		Message string `json:"message"`
	}
	if jerr := json.Unmarshal([]byte(lines[len(lines)-1]), &r); jerr != nil {
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("unexpected answer from claude: %s", strings.TrimSpace(string(out)))
	}
	if r.Outcome != "ok" {
		return "", errors.New(r.Message)
	}
	return r.Message, nil
}
