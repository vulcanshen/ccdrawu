package main

import (
	"bytes"
	"io"
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
	loc := config.Locations{Cache: filepath.Join(dir, "data"), ConfigFile: filepath.Join(dir, "config", "config.yaml"), Claude: filepath.Join(dir, "claude"), Data: filepath.Join(dir, ".ccdrawu")}
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
	// The theme kept in originals.yaml, the status line where 0.1.4 kept it,
	// in the config, not moved yet (tdp E3)
	cfg.Theme.Base = "dark-ansi"
	if err := config.WriteOriginals(loc.Originals(), config.Originals{Theme: before}); err != nil {
		t.Fatal(err)
	}
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
	if _, err := os.Stat(loc.Data); err == nil {
		t.Fatalf("%s still there with nothing kept", loc.Data)
	}

	// Again, there is nothing of ccdrawu's left to put back
	out.Reset()
	if err := putBack(loc, &out); err != nil || out.Len() != 0 {
		t.Fatalf("again: %q %v", out.String(), err)
	}
}

// tdp E3, can't be read, with no screen: uninstall claude stops on a config or
// originals that can't be read before it touches anything, Claude Code's
// plugin included, the error naming the file and why
func TestUninstallStopsOnAFileThatCantBeRead(t *testing.T) {
	dir := t.TempDir()
	loc := config.Locations{Cache: filepath.Join(dir, "cache"), ConfigFile: filepath.Join(dir, "config", "config.yaml"), Claude: filepath.Join(dir, "claude"), Data: filepath.Join(dir, ".ccdrawu")}
	run := func(name string, args ...string) ([]byte, error) {
		t.Fatalf("ran %s %v", name, args)
		return nil, nil
	}
	write := func(path, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write(loc.ConfigFile, "drawer: [oops\n")
	var out bytes.Buffer
	if err := uninstall(loc, run, &out, io.Discard); err == nil || !strings.HasPrefix(err.Error(), loc.ConfigFile+" can't be read: ") {
		t.Fatalf("config: %v", err)
	}

	write(loc.ConfigFile, "theme:\n  base: dark-ansi\n")
	write(loc.Originals(), "theme: [oops\n")
	if err := uninstall(loc, run, &out, io.Discard); err == nil || !strings.HasPrefix(err.Error(), loc.Originals()+" can't be read: ") {
		t.Fatalf("originals: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("printed %q", out.String())
	}
	if bak, _ := filepath.Glob(filepath.Join(dir, "*", "*.bak*")); len(bak) != 0 {
		t.Errorf("moved aside: %v", bak)
	}
}

// tdp E3-31: values that do not fit, with no screen, are their defaults, the
// file kept aside and written again, said on stderr; uninstall goes on
func TestUninstallGoesOnWithValuesThatDoNotFit(t *testing.T) {
	dir := t.TempDir()
	loc := config.Locations{Cache: filepath.Join(dir, "cache"), ConfigFile: filepath.Join(dir, "config", "config.yaml"), Claude: filepath.Join(dir, "claude"), Data: filepath.Join(dir, ".ccdrawu")}
	if err := os.MkdirAll(filepath.Dir(loc.ConfigFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(loc.ConfigFile, []byte("bar: wavy\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var ran bool
	run := func(name string, args ...string) ([]byte, error) {
		ran = true
		return []byte("[]"), nil
	}
	var out, errs bytes.Buffer
	if err := uninstall(loc, run, &out, &errs); err != nil {
		t.Fatal(err)
	}
	// The toast's row and the value's WARN, in the TUI's words (tdp E3-40 on tdp main)
	want := "ccdrawu: 1 value in config.yaml did not fit: defaults used, yours kept in config.yaml.bak\n" +
		"ccdrawu: config.yaml: bar \"wavy\" did not fit; using \"squares\"\n"
	if !ran || errs.String() != want {
		t.Fatalf("ran %v, stderr %q", ran, errs.String())
	}
	if _, err := os.Stat(loc.ConfigFile + ".bak"); err != nil {
		t.Error("not kept aside:", err)
	}
}

// tdp E1 (v1.3.0): update says so when ccdrawu is the latest release, and
// when it is newer than it, a beta
func TestUpdateSaysItIsUpToDate(t *testing.T) {
	for _, c := range []struct{ version, latest, want string }{
		{"1.4.1", "1.4.1", "ccdrawu 1.4.1 is the latest release."},
		{"1.5.0-beta.1", "1.4.1", "ccdrawu 1.5.0-beta.1 is newer than the latest release, 1.4.1."},
	} {
		if got := upToDate(c.version, c.latest); got != c.want {
			t.Errorf("upToDate(%s, %s) = %q, want %q", c.version, c.latest, got, c.want)
		}
	}
}

// tdp E1 (tdp #24, on tdp main): a wrong use says what is wrong, a blank line
// and the usage, exit status 2; a line of a command that could not be done is
// a toast's row, a capital first unless it starts with a path or a program's
// name
func TestWrongUseAndSentences(t *testing.T) {
	var errs bytes.Buffer
	if status := wrongUse(&errs, "Unknown command \"x\""); status != 2 || errs.String() != "ccdrawu: Unknown command \"x\"\n\n"+usage {
		t.Fatalf("exit %d, stderr %q", status, errs.String())
	}
	for line, want := range map[string]string{
		"looking up the latest release: EOF":      "Looking up the latest release: EOF",
		"invalid character 'x' looking for value": "Invalid character 'x' looking for value",
		"config.yaml can't be read: not a config": "config.yaml can't be read: not a config",
		"/tmp/x can't be read: denied":            "/tmp/x can't be read: denied",
		"~/.ccdrawu/originals.yaml can't be read": "~/.ccdrawu/originals.yaml can't be read",
		"claude plugin list: exit status 1":       "claude plugin list: exit status 1",
		"brew upgrade vulcanshen/tap/ccdrawu: x":  "brew upgrade vulcanshen/tap/ccdrawu: x",
		"This ccdrawu is built from source":       "This ccdrawu is built from source",
		"":                                        "",
	} {
		if got := sentence(line); got != want {
			t.Errorf("%q: %q", line, got)
		}
	}
}

// doctor sets right what the drawer and the status line say is wrong, as a
// start of ccdrawu does with no screen (tdp E3-31): an item the plugin does not
// have taken out and a value that does not fit its default, the file kept
// aside and written again, each said on stderr with what it was; a key ccdrawu
// does not read stays (tdp E3-38). Run again, there is nothing to set right
func TestDoctorSetsRightWhatDoesNotFit(t *testing.T) {
	dir := t.TempDir()
	loc := config.Locations{Cache: filepath.Join(dir, "cache"), ConfigFile: filepath.Join(dir, "config", "config.yaml"), Claude: filepath.Join(dir, "claude"), Data: filepath.Join(dir, ".ccdrawu")}
	for path, text := range map[string]string{
		loc.ConfigFile: "dock:\n  - pet\n  - account\nbar: wavy\nrefesh: 5\n",
		loc.Catalog():  `{"items": [{"id": "pet"}, {"id": "model"}]}`,
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out, errs bytes.Buffer
	if status := doctor(loc, &out, &errs); status != 0 {
		t.Fatalf("exit %d, stderr %q", status, errs.String())
	}
	want := "ccdrawu: 1 value and 1 entry in config.yaml did not fit: defaults used and entries taken out, yours kept in config.yaml.bak\n" +
		"ccdrawu: config.yaml: dock \"account\" did not fit; taken out\n" +
		"ccdrawu: config.yaml: bar \"wavy\" did not fit; using \"squares\"\n" +
		"ccdrawu: Not read: refesh in " + loc.ConfigFile + "; delete it if you don't need it\n"
	if errs.String() != want || out.Len() != 0 {
		t.Fatalf("stderr:\n%s\nwant:\n%s\nstdout %q", errs.String(), want, out.String())
	}
	if kept, _ := os.ReadFile(loc.ConfigFile + ".bak"); !strings.Contains(string(kept), "account") {
		t.Errorf("not kept aside: %q", kept)
	}
	c, misfits, err := config.Check(loc)
	if err != nil || len(misfits) != 0 || strings.Join(c.Dock, ",") != "pet" {
		t.Fatalf("written again: %+v %v %v", c, misfits, err)
	}
	if text, _ := os.ReadFile(loc.ConfigFile); !strings.Contains(string(text), "refesh") {
		t.Errorf("the key not read is gone:\n%s", text)
	}

	out.Reset()
	errs.Reset()
	if status := doctor(loc, &out, &errs); status != 0 || out.String() != "Nothing to set right in "+loc.ConfigFile+".\n" {
		t.Fatalf("again: exit %d, stdout %q", status, out.String())
	}
	if bak, _ := filepath.Glob(loc.ConfigFile + ".bak.*"); len(bak) != 0 {
		t.Errorf("kept aside again: %v", bak)
	}
}

// tdp E3, can't be read, with no screen: doctor leaves a config or originals
// that can't be read as it is, as Esc does, says which, why and what to do,
// and exits 1
func TestDoctorLeavesAFileThatCantBeRead(t *testing.T) {
	dir := t.TempDir()
	loc := config.Locations{Cache: filepath.Join(dir, "cache"), ConfigFile: filepath.Join(dir, "config", "config.yaml"), Claude: filepath.Join(dir, "claude"), Data: filepath.Join(dir, ".ccdrawu")}
	write := func(path, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []struct{ file, text string }{{loc.ConfigFile, "drawer: [oops\n"}, {loc.Originals(), "theme: [oops\n"}} {
		write(loc.ConfigFile, "bar: wavy\n")
		write(c.file, c.text)
		var out, errs bytes.Buffer
		if status := doctor(loc, &out, &errs); status != 1 || out.Len() != 0 {
			t.Fatalf("%s: exit %d, stdout %q", c.file, status, out.String())
		}
		lines := strings.Split(strings.TrimSuffix(errs.String(), "\n"), "\n")
		if len(lines) != 2 || !strings.HasPrefix(lines[0], "ccdrawu: "+c.file+" can't be read: ") || lines[1] != "ccdrawu: Open `ccdrawu` to move it aside and start fresh, or set it right by hand." {
			t.Fatalf("%s: stderr %q", c.file, errs.String())
		}
		if got, _ := os.ReadFile(c.file); string(got) != c.text {
			t.Errorf("%s changed: %q", c.file, got)
		}
		if bak, _ := filepath.Glob(filepath.Join(dir, "*", "*.bak*")); len(bak) != 0 {
			t.Errorf("%s: moved aside: %v", c.file, bak)
		}
	}
}
