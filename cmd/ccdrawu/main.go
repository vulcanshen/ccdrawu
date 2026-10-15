// Command ccdrawu sets up the ccdrawu mod for Claude Code and picks the items
// its drawer shows.
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/vulcanshen/ccdrawu/internal/claude"
	"github.com/vulcanshen/ccdrawu/internal/config"
	"github.com/vulcanshen/ccdrawu/internal/statusline"
	"github.com/vulcanshen/ccdrawu/internal/ui"
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
		msg, err := claude.Install(claude.Exec, claude.SourceFrom(os.Getenv))
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
	if !ok {
		fmt.Println("The ccdrawu plugin is not in Claude Code.")
		return nil
	}
	fmt.Println(msg)
	loc, err := config.UserLocations()
	if err != nil {
		return err
	}
	// ccdrawu's theme goes with the drawer: Claude Code has the user's own back
	if cfg, err := config.ReadConfig(loc.ConfigFile); err == nil && cfg.Theme.Base != "" {
		if err := claude.RestoreTheme(loc.Claude, cfg.Theme.Before); err != nil {
			return err
		}
		cfg.Theme.Base, cfg.Theme.Before = "", ""
		if err := config.WriteConfig(loc.ConfigFile, cfg); err != nil {
			return err
		}
		fmt.Println("Claude Code has its own theme back.")
	}
	// So does the statusline: Claude Code has the user's own status line back, or none
	if cfg, err := config.ReadConfig(loc.ConfigFile); err == nil && cfg.Statusline.On {
		if err := claude.RestoreStatusline(loc.Claude, cfg.Statusline.Before); err != nil {
			return err
		}
		cfg.Statusline.On, cfg.Statusline.Before = false, nil
		if err := config.WriteConfig(loc.ConfigFile, cfg); err != nil {
			return err
		}
		fmt.Println("Claude Code has its own status line back.")
	}
	fmt.Printf("Your picks stay in %s, the plugin's data in %s.\n", loc.ConfigFile, loc.Data)
	return nil
}

const usage = `ccdrawu — a status drawer for Claude Code

Usage:
  ccdrawu                     pick the items the drawer shows
  ccdrawu install claude      add the drawer to Claude Code (its ccdrawu plugin)
  ccdrawu uninstall claude    take it out again
  ccdrawu statusline          Claude Code's status line, as set in ccdrawu
                              (Claude Code runs it; the session on stdin)
  ccdrawu version             print the version (also --version, -v)
  ccdrawu help                print this (also --help, -h)

Environment:
  CCDRAWU__CONFIG            the settings folder, instead of ~/.config/ccdrawu
  CCDRAWU__DATA              the data folder, instead of ~/.ccdrawu
  CCDRAWU__MARKETPLACE       where install takes the plugin from, instead of
                             vulcanshen/ccdrawu: owner/repo, a git URL or a folder
  CCDRAWU__ICON_WIDTH        1 or 2: how many cells a Nerd Font icon takes,
                             instead of asking the terminal
  CCDRAWU__POWERLINE_WIDTH   1 or 2: the same for a powerline symbol
  TERMINU__ICON_WIDTH        the same as CCDRAWU__ICON_WIDTH, for every terminu app
  TERMINU__POWERLINE_WIDTH   the same as CCDRAWU__POWERLINE_WIDTH, for every terminu app
`
