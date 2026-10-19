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
	"strconv"
	"time"

	"github.com/vulcanshen/ccdrawu/internal/claude"
	"github.com/vulcanshen/ccdrawu/internal/config"
	"github.com/vulcanshen/ccdrawu/internal/statusline"
	"github.com/vulcanshen/ccdrawu/internal/ui"
	"github.com/vulcanshen/ccdrawu/internal/update"
	"github.com/vulcanshen/ccdrawu/internal/version"
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
				fmt.Fprintf(os.Stderr, "ccdrawu: say where: ccdrawu %s claude\n", os.Args[1])
				os.Exit(2)
			}
			if err := integrate(os.Args[1] == "install"); err != nil {
				fmt.Fprintln(os.Stderr, "ccdrawu:", err)
				os.Exit(1)
			}
			return
		case "update":
			if err := selfUpdate(os.Stdout); err != nil {
				fmt.Fprintln(os.Stderr, "ccdrawu:", err)
				os.Exit(1)
			}
			return
		case "statusline":
			// Claude Code runs it with the session on stdin and the terminal's width in COLUMNS
			loc, err := config.UserLocations()
			if err == nil {
				columns, _ := strconv.Atoi(os.Getenv("COLUMNS"))
				err = statusline.Run(os.Stdin, os.Stdout, loc, columns, time.Now())
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, "ccdrawu:", err)
				os.Exit(1)
			}
			return
		default:
			fmt.Fprintf(os.Stderr, "ccdrawu: unknown command %q\n\n%s", os.Args[1], usage)
			os.Exit(2)
		}
	}
	loc, err := config.UserLocations()
	if err == nil {
		err = ui.Run(loc)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ccdrawu:", err)
		os.Exit(1)
	}
}

// integrate puts the plugin into Claude Code, or takes it out.
func integrate(install bool) error {
	if install {
		msg, err := claude.Install(claude.Exec, claude.SourceFrom(os.Getenv, version.Version))
		if err != nil {
			return err
		}
		fmt.Println(msg)
		fmt.Println("Open Claude Code sessions take the drawer with /reload-plugins; new ones have it from the start.")
		fmt.Println("Run ccdrawu to pick what it shows.")
		return nil
	}
	msg, ok, err := claude.Uninstall(claude.Exec)
	if err != nil {
		return err
	}
	if ok {
		fmt.Println(msg)
	} else {
		fmt.Println("The ccdrawu plugin is not in Claude Code.")
	}
	loc, err := config.UserLocations()
	if err != nil {
		return err
	}
	// With the plugin gone already too: a status line left set would run a ccdrawu that the
	// uninstall script is about to remove
	if err := putBack(loc, os.Stdout); err != nil {
		return err
	}
	fmt.Printf("Your picks stay in %s, what the plugin wrote for ccdrawu in %s.\n", loc.ConfigFile, loc.Cache)
	return nil
}

// selfUpdate brings ccdrawu to its latest release the way it was installed,
// then, where Claude Code has the plugin, the plugin to the same release: the
// new ccdrawu runs its own install claude, which takes the plugin of its
// version. It never asks for more rights: a folder it cannot write in, it says
// so.
func selfUpdate(out io.Writer) error {
	if version.Version == "dev" {
		return errors.New("this ccdrawu is built from source: update it there (git pull && make install), then run ccdrawu install claude")
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
					return fmt.Errorf("%w\nRun the install script again, or ccdrawu update as the folder's owner", err)
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
		fmt.Fprintln(out, "The ccdrawu plugin is not in Claude Code; ccdrawu install claude adds it.")
		return nil
	}
	install := exec.Command(next, "install", "claude")
	install.Stdout, install.Stderr = out, os.Stderr
	return install.Run()
}

// putBack gives Claude Code the user's own theme and status line back where
// ccdrawu's are set, or none where the user had none, and says so on out.
func putBack(loc config.Locations, out io.Writer) error {
	// ccdrawu's theme goes with the drawer: Claude Code has the user's own back
	if cfg, err := config.ReadConfig(loc.ConfigFile); err == nil && cfg.Theme.Base != "" {
		if err := claude.RestoreTheme(loc.Claude, cfg.Theme.Before); err != nil {
			return err
		}
		cfg.Theme.Base, cfg.Theme.Before = "", ""
		if err := config.WriteConfig(loc.ConfigFile, cfg); err != nil {
			return err
		}
		fmt.Fprintln(out, "Claude Code has its own theme back.")
	}
	// So does the statusline: Claude Code has the user's own status line back, or none
	if cfg, err := config.ReadConfig(loc.ConfigFile); err == nil && cfg.Statusline.On {
		if err := claude.RestoreStatusline(loc.Claude, json.RawMessage(cfg.Statusline.Before)); err != nil {
			return err
		}
		cfg.Statusline.On, cfg.Statusline.Before = false, nil
		if err := config.WriteConfig(loc.ConfigFile, cfg); err != nil {
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
  ccdrawu statusline          Claude Code's status line, as set in ccdrawu
                              (Claude Code runs it; the session on stdin)
  ccdrawu version             print the version (also --version, -v)
  ccdrawu help                print this (also --help, -h)

Environment:
  CCDRAWU__CONFIG            the settings folder, instead of ~/.config/ccdrawu
  CCDRAWU__CACHE             what the plugin writes for ccdrawu, instead of
                             ccdrawu in the system's cache folder
  CCDRAWU__MARKETPLACE       where install takes the plugin from, instead of
                             vulcanshen/ccdrawu: owner/repo, a git URL or a folder
  CCDRAWU__ICON_WIDTH        1 or 2: how many cells a Nerd Font icon takes,
                             instead of asking the terminal
  CCDRAWU__POWERLINE_WIDTH   1 or 2: the same for a powerline symbol
  TERMINU__ICON_WIDTH        the same as CCDRAWU__ICON_WIDTH, for every terminu app
  TERMINU__POWERLINE_WIDTH   the same as CCDRAWU__POWERLINE_WIDTH, for every terminu app
`
