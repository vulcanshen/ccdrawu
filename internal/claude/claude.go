// Package claude puts the ccdrawu plugin into Claude Code and takes it out,
// through Claude Code's own `claude plugin` commands.
package claude

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Plugin is the plugin's name; Repo is the marketplace it comes from, cloned
// over HTTPS (owner/repo clones over SSH, which needs a key on GitHub).
const (
	Plugin = "ccdrawu"
	Repo   = "https://github.com/vulcanshen/ccdrawu.git"
)

// SourceFrom is where install takes the plugin from: CCDRAWU__MARKETPLACE
// (a folder, while developing), else Repo at the release tag of this ccdrawu,
// version as internal/version has it, so the plugin and the command go
// together; a dev build takes the default branch. Unpinned, the plugin would
// be whatever the default branch holds, released or not.
func SourceFrom(getenv func(string) string, version string) string {
	if s := getenv("CCDRAWU__MARKETPLACE"); s != "" {
		return s
	}
	if version == "" || version == "dev" {
		return Repo
	}
	return Repo + "#v" + version
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
// marketplace first when Claude Code does not have it. Where Claude Code has
// the plugin from that same source already it updates it instead: Claude Code
// refreshes a marketplace like ours only when asked, and installing again is
// no change. From another source (an older ccdrawu's tag, the default branch)
// it moves it: Claude Code adds no marketplace of a name it has from another
// source, so the marketplace goes, and the plugin with it, and the plugin is
// installed from source. It returns Claude Code's own word on what it did.
func Install(run Run, source string) (string, error) {
	id, _, err := Installed(run)
	if err != nil {
		return "", err
	}
	install := func() (string, error) {
		return result(run("claude", "plugin", "install", Plugin, "--marketplace", source, "--json"))
	}
	if id == "" {
		return install()
	}
	_, market, _ := strings.Cut(id, "@")
	from, err := recorded(run, market)
	if err != nil {
		return "", err
	}
	if from != sourceKey(source) {
		if _, err := result(run("claude", "plugin", "marketplace", "remove", market, "--json")); err != nil {
			return "", err
		}
		return install()
	}
	if _, err := result(run("claude", "plugin", "marketplace", "update", market, "--json")); err != nil {
		return "", err
	}
	return result(run("claude", "plugin", "update", id, "--json"))
}

// recorded is where Claude Code has the marketplace market from, written as
// sourceKey writes a source: a git URL and its #ref, or a folder; "" when it
// does not have it.
func recorded(run Run, market string) (string, error) {
	out, err := run("claude", "plugin", "marketplace", "list", "--json")
	if err != nil {
		return "", err
	}
	var markets []struct {
		Name, Source, URL, Repo, Ref, Path string
	}
	if err := json.Unmarshal(out, &markets); err != nil {
		return "", fmt.Errorf("claude plugin marketplace list: %w", err)
	}
	for _, m := range markets {
		if m.Name != market {
			continue
		}
		at := ""
		if m.Ref != "" {
			at = "#" + m.Ref
		}
		switch m.Source {
		case "git":
			return m.URL + at, nil
		case "github":
			return m.Repo + at, nil
		case "directory":
			return filepath.Clean(m.Path), nil
		}
		return m.Source, nil
	}
	return "", nil
}

// sourceKey is a source as recorded reads one: a folder by its absolute path,
// anything else as it is written.
func sourceKey(source string) string {
	if info, err := os.Stat(source); err == nil && info.IsDir() {
		if abs, err := filepath.Abs(source); err == nil {
			return abs
		}
	}
	return source
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
