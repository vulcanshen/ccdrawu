// Command ccdrawu sets up the ccdrawu mod for Claude Code and picks the items
// its drawer shows.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/vulcanshen/ccdrawu/internal/claude"
	"github.com/vulcanshen/ccdrawu/internal/config"
	"github.com/vulcanshen/ccdrawu/internal/statusline"
	"github.com/vulcanshen/ccdrawu/internal/ui"
	"github.com/vulcanshen/ccdrawu/internal/update"
	"github.com/vulcanshen/ccdrawu/internal/version"
	"github.com/vulcanshen/terminu-go"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Println("ccdrawu " + version.Display())
			return
		case "help", "--help", "-h":
			fmt.Print(usage)
			return
		case "install", "uninstall":
			if len(os.Args) != 3 || os.Args[2] != "claude" {
				os.Exit(wrongUse(os.Stderr, fmt.Sprintf("Say where: `ccdrawu %s claude`", os.Args[1])))
			}
			failed(integrate(os.Args[1] == "install"))
			return
		case "update":
			failed(selfUpdate(os.Stdout))
			return
		case "doctor":
			loc, err := config.UserLocations()
			failed(err)
			os.Exit(doctor(loc, os.Stdout, os.Stderr))
		case "statusline":
			// Claude Code runs it with the session on stdin and the terminal's width in COLUMNS
			loc, err := config.UserLocations()
			if err == nil {
				columns, _ := strconv.Atoi(os.Getenv("COLUMNS"))
				err = statusline.Run(os.Stdin, os.Stdout, loc, columns, time.Now())
			}
			failed(err)
			return
		default:
			os.Exit(wrongUse(os.Stderr, fmt.Sprintf("Unknown command %q", os.Args[1])))
		}
	}
	loc, err := config.UserLocations()
	if err == nil {
		err = ui.Run(loc)
	}
	failed(err)
}

// failed ends a command that could not be done, saying why on stderr, exit
// status 1 (tdp E1); with no error it does nothing.
func failed(err error) {
	if err == nil {
		return
	}
	for _, line := range strings.Split(err.Error(), "\n") {
		fmt.Fprintln(os.Stderr, "ccdrawu: "+sentence(line))
	}
	os.Exit(1)
}

// wrongUse says on errs what is wrong with how ccdrawu was run, then the
// usage after a blank line; its result is the exit status, 2 (tdp E1).
func wrongUse(errs io.Writer, what string) int {
	fmt.Fprintf(errs, "ccdrawu: %s\n\n%s", what, usage)
	return 2
}

// sentence is a line as a toast's row has it (tdp E1): a capital first, but
// for one that starts with a path, a file's name or a program's, which keeps
// its own case.
func sentence(line string) string {
	first, _, _ := strings.Cut(line, " ")
	first = strings.TrimRight(first, ":,;")
	if line == "" || strings.ContainsAny(first, "/\\.~") || slices.Contains([]string{"ccdrawu", "claude", "brew", "scoop"}, first) {
		return line
	}
	r, size := utf8.DecodeRuneInString(line)
	return string(unicode.ToUpper(r)) + line[size:]
}

// integrate puts the plugin into Claude Code, or takes it out.
func integrate(install bool) error {
	if install {
		msg, err := claude.Install(claude.Exec, claude.SourceFrom(os.Getenv, version.Version))
		if err != nil {
			return err
		}
		fmt.Println(msg)
		fmt.Println("Open Claude Code sessions take the drawer with `/reload-plugins`; new ones have it from the start.")
		fmt.Println("Run `ccdrawu` to pick what it shows.")
		return nil
	}
	loc, err := config.UserLocations()
	if err != nil {
		return err
	}
	return uninstall(loc, claude.Exec, os.Stdout, os.Stderr)
}

// uninstall takes the plugin out of Claude Code and gives Claude Code the
// user's own theme and status line back, saying so on out. A config or
// originals that can't be read stops it before it touches anything: with no
// screen to ask on, it does what the TUI's Esc does (tdp E3, can't be read).
// Values in the config it can't use are their defaults, the file kept aside
// and written again with them, said on errs (tdp E3, a value that does not
// fit).
func uninstall(loc config.Locations, run claude.Run, out, errs io.Writer) error {
	cfg, settled, err := config.Settle(loc)
	if err != nil {
		return errors.New(terminu.CannotReadText(config.Source(loc.ConfigFile), err))
	}
	for _, line := range settled.Lines(loc) {
		fmt.Fprintln(errs, "ccdrawu: "+line)
	}
	if _, err := config.ReadOriginals(loc.Originals(), cfg); err != nil {
		return errors.New(terminu.CannotReadText(loc.Originals(), err))
	}
	msg, ok, err := claude.Uninstall(run)
	if err != nil {
		return err
	}
	if ok {
		fmt.Fprintln(out, msg)
	} else {
		fmt.Fprintln(out, "The ccdrawu plugin is not in Claude Code.")
	}
	// With the plugin gone already too: a status line left set would run a ccdrawu that the
	// uninstall script is about to remove
	if err := putBack(loc, out); err != nil {
		return err
	}
	fmt.Fprintf(out, "Your picks stay in %s, what the plugin wrote for ccdrawu in %s.\n", loc.ConfigFile, loc.Cache)
	return nil
}

// doctor sets right what the drawer and the status line say is wrong with
// ccdrawu's own files, as ccdrawu does when it opens, with no screen to ask on
// (tdp E3): what 0.1.4 and before kept elsewhere moves where it is now; values
// in config.yaml ccdrawu can't use are their defaults, the file copied aside
// and written again with them. What it did goes to errs, a line each, in the
// TUI's words: its toasts' rows, its App Log's WARNs and INFOs, its error
// popup's rows. A config or originals that can't be read it leaves as it is,
// as Esc does: it says which and why, and what to do (tdp E3, can't be read).
// The result is the exit status: 1 where it could not set a file right, else
// 0 (tdp E1). With nothing to set right, it says so on out.
func doctor(loc config.Locations, out, errs io.Writer) int {
	say := func(line string) { fmt.Fprintln(errs, "ccdrawu: "+line) }
	cantRead := func(path string, err error) int {
		say(terminu.CannotReadText(path, err))
		say("Open `ccdrawu` to move it aside and start fresh, or set it right by hand.")
		return 1
	}
	if _, err := config.ReadConfig(loc.ConfigFile); err != nil {
		return cantRead(config.Source(loc.ConfigFile), err)
	}
	if _, err := config.ReadOriginals(loc.Originals(), config.Config{}); err != nil {
		return cantRead(loc.Originals(), err)
	}
	status := 0
	moved, moveErr := config.Move(loc)
	for _, m := range moved {
		what := config.ShortHome(m.From)
		if m.What != "" {
			what = m.What + " in " + what
		}
		say("Moved " + what + " to " + config.ShortHome(m.To))
	}
	if moveErr != nil {
		// Each of errors.Join's says which place ccdrawu reads, and what to do by hand
		errs := []error{moveErr}
		if joined, ok := moveErr.(interface{ Unwrap() []error }); ok {
			errs = joined.Unwrap()
		}
		for _, err := range errs {
			say(err.Error())
		}
		status = 1
	}
	_, settled, err := config.Settle(loc)
	if err != nil {
		return cantRead(config.Source(loc.ConfigFile), err)
	}
	for _, line := range settled.Lines(loc) {
		say(line)
	}
	if settled.Err != nil {
		status = 1
	}
	if len(moved) == 0 && moveErr == nil && len(settled.Misfits) == 0 {
		fmt.Fprintf(out, "Nothing to set right in %s.\n", config.ShortHome(loc.ConfigFile))
	}
	return status
}

// selfUpdate brings ccdrawu to its latest release the way it was installed,
// then, where Claude Code has the plugin, the plugin to the same release: the
// new ccdrawu runs its own install claude, which takes the plugin of its
// version. It never asks for more rights: a folder it cannot write in, it says
// so.
func selfUpdate(out io.Writer) error {
	if version.Version == "dev" {
		return errors.New("This ccdrawu is built from source: update it there (`git pull && make install`), then run `ccdrawu install claude`")
	}
	exe, err := os.Executable()
	if err == nil {
		exe, err = filepath.EvalSymlinks(exe)
	}
	if err != nil {
		return err
	}
	latest, err := update.GitHub.LatestVersion()
	if err != nil {
		return fmt.Errorf("looking up the latest release: %w", err)
	}
	next := exe // the ccdrawu that updates the plugin
	if !update.Newer(latest, version.Version) {
		fmt.Fprintf(out, "ccdrawu %s is the latest release.\n", version.Version)
	} else {
		switch update.How(exe) {
		case update.Homebrew:
			fmt.Fprintf(out, "Updating ccdrawu %s to %s with Homebrew...\n", version.Version, latest)
			brew := exec.Command("brew", "upgrade", "vulcanshen/tap/ccdrawu")
			brew.Stdout, brew.Stderr = out, os.Stderr
			if err := brew.Run(); err != nil {
				return fmt.Errorf("brew upgrade vulcanshen/tap/ccdrawu: %w", err)
			}
			next = update.BrewBin(exe)
		case update.Scoop:
			// Scoop refuses to update an app while it runs, and this is it
			fmt.Fprintf(out, "ccdrawu %s is out. Scoop installed this one: close ccdrawu and run\n\n  scoop update ccdrawu\n  ccdrawu install claude\n", latest)
			return nil
		default:
			fmt.Fprintf(out, "Updating ccdrawu %s to %s...\n", version.Version, latest)
			if err := update.GitHub.Replace(exe, latest, runtime.GOOS, runtime.GOARCH); err != nil {
				if errors.Is(err, os.ErrPermission) {
					return fmt.Errorf("%w\nRun the install script again, or `ccdrawu update` as the folder's owner", err)
				}
				return err
			}
			fmt.Fprintf(out, "ccdrawu %s is in %s.\n", latest, exe)
		}
	}
	// The plugin goes with the command, where Claude Code has it; one taken out stays out
	id, _, err := claude.Installed(claude.Exec)
	if err != nil {
		return err
	}
	if id == "" {
		fmt.Fprintln(out, "The ccdrawu plugin is not in Claude Code; `ccdrawu install claude` adds it.")
		return nil
	}
	install := exec.Command(next, "install", "claude")
	install.Stdout, install.Stderr = out, os.Stderr
	return install.Run()
}

// putBack gives Claude Code the user's own theme and status line back where
// ccdrawu's are set, or none where the user had none, and says so on out.
func putBack(loc config.Locations, out io.Writer) error {
	cfg, err := config.ReadConfig(loc.ConfigFile)
	if err != nil || (cfg.Theme.Base == "" && !cfg.Statusline.On) {
		return nil
	}
	orig, err := config.ReadOriginals(loc.Originals(), cfg)
	if err != nil {
		return err
	}
	// The originals first, as they can't be found again (tdp E3); the
	// config's old place for them is done with once they are written
	save := func() error {
		if err := config.WriteOriginals(loc.Originals(), orig); err != nil {
			return err
		}
		cfg.Theme.Before, cfg.Statusline.Before = "", nil
		return config.WriteConfig(loc.ConfigFile, cfg)
	}
	// ccdrawu's theme goes with the drawer: Claude Code has the user's own back
	if cfg.Theme.Base != "" {
		if err := claude.RestoreTheme(loc.Claude, orig.Theme); err != nil {
			return err
		}
		cfg.Theme.Base, orig.Theme = "", ""
		if err := save(); err != nil {
			return err
		}
		fmt.Fprintln(out, "Claude Code has its own theme back.")
	}
	// So does the statusline: Claude Code has the user's own status line back, or none
	if cfg.Statusline.On {
		if err := claude.RestoreStatusline(loc.Claude, json.RawMessage(orig.Statusline)); err != nil {
			return err
		}
		cfg.Statusline.On, orig.Statusline = false, nil
		if err := save(); err != nil {
			return err
		}
		fmt.Fprintln(out, "Claude Code has its own status line back.")
	}
	return nil
}

const usage = `ccdrawu — a status drawer for Claude Code

Usage:
  ccdrawu                     pick the items the drawer shows
  ccdrawu install claude      add the drawer to Claude Code (its ccdrawu plugin),
                              or update it there
  ccdrawu uninstall claude    take it out again
  ccdrawu update              update ccdrawu to its latest release, the way it
                              was installed, and the plugin in Claude Code with it
  ccdrawu doctor              set right what the drawer or the status line says
                              is wrong with config.yaml, as ccdrawu does when
                              it opens
  ccdrawu statusline          Claude Code's status line, as set in ccdrawu
                              (Claude Code runs it; the session on stdin)
  ccdrawu version             print the version (also --version, -v)
  ccdrawu help                print this (also --help, -h)

Environment:
  CCDRAWU__CONFIG            the settings folder, instead of ~/.config/ccdrawu
  CCDRAWU__CACHE             what the plugin writes for ccdrawu, instead of
                             ccdrawu in the system's cache folder
  CCDRAWU__DATA              the theme and status line ccdrawu replaced in
                             Claude Code, to put back, instead of ~/.ccdrawu
  CCDRAWU__MARKETPLACE       where install takes the plugin from, instead of
                             vulcanshen/ccdrawu: owner/repo, a git URL or a folder
  CCDRAWU__ICON_WIDTH        1 or 2: how many cells a Nerd Font icon takes,
                             instead of asking the terminal
  CCDRAWU__POWERLINE_WIDTH   1 or 2: the same for a powerline symbol
  TERMINU__ICON_WIDTH        the same as CCDRAWU__ICON_WIDTH, for every terminu app
  TERMINU__POWERLINE_WIDTH   the same as CCDRAWU__POWERLINE_WIDTH, for every terminu app
`
