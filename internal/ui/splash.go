package ui

import (
	_ "embed"

	"github.com/vulcanshen/ccdrawu/internal/version"
	terminu "github.com/vulcanshen/terminu-go"
)

// icon is ccdrawu's family mark (tdp S5), drawn by the splash: a copy of
// docs/icon.svg, which go:embed can't reach from here (a test holds the two
// the same).
//
//go:embed icon.svg
var icon []byte

// splash is the family's easter egg, V on a panel (tdp S).
func splash() terminu.Splash {
	s, err := terminu.NewSplash(icon, "ccdrawu", version.Display(), "A status drawer for Claude Code")
	if err != nil {
		panic(err) // the embedded icon is broken: a mistake in the program
	}
	return s
}
