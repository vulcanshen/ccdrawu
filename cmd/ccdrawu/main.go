// Command ccdrawu sets up the ccdrawu mod for Claude Code and picks the items
// its drawer shows.
package main

import (
	"fmt"
	"os"
	"time"

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
		case "statusline":
			statusLine()
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

// statusLine is what Claude Code's statusLine setting runs. What it prints is
// the status line, so it prints nothing; a failure goes to stderr alone.
func statusLine() {
	loc, err := config.UserLocations()
	if err == nil {
		err = statusline.Write(os.Stdin, loc.Data, time.Now())
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ccdrawu statusline:", err)
	}
}

const usage = `ccdrawu — a status drawer for Claude Code

Usage:
  ccdrawu              pick the items the drawer shows
  ccdrawu statusline   for Claude Code's statusLine: hands vim's live mode to the drawer
  ccdrawu version      print the version (also --version, -v)
  ccdrawu help         print this (also --help, -h)

Environment:
  CCDRAWU__CONFIG            the settings folder, instead of ~/.config/ccdrawu
  CCDRAWU__DATA              the data folder, instead of ~/.ccdrawu
  CCDRAWU__ICON_WIDTH        1 or 2: how many cells a Nerd Font icon takes,
                             instead of asking the terminal
  CCDRAWU__POWERLINE_WIDTH   1 or 2: the same for a powerline symbol
  TERMINU__ICON_WIDTH        the same as CCDRAWU__ICON_WIDTH, for every terminu app
  TERMINU__POWERLINE_WIDTH   the same as CCDRAWU__POWERLINE_WIDTH, for every terminu app
`
