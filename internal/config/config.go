package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var aliasName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*$`)

// CleanAlias is the alias as the mod reads it (hooks/format.ts commandNames):
// spaces around it and leading slashes dropped, "" for none. One the mod
// would not take is an error saying why.
func CleanAlias(s string) (string, error) {
	name := strings.TrimLeft(strings.TrimSpace(s), "/")
	switch {
	case name == "":
		return "", nil
	case name == "ccdrawu":
		return "", errors.New("That is /ccdrawu itself")
	case !aliasName.MatchString(name):
		return "", errors.New("One name of letters, digits and -, such as ccd")
	}
	return name, nil
}

// Item is one row of the catalog the mod writes to <data>/catalog.json
// (hooks/items.ts): what the drawer can show, in the order it shows it.
type Item struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Name        string `json:"name"`
	Group       string `json:"group"`
	Description string `json:"description"`
	Dock        bool   `json:"dock"`
	Inline      bool   `json:"inline"`
}

type Catalog struct {
	Version int    `json:"version"`
	Items   []Item `json:"items"`
}

// ReadCatalog reads the catalog; ok is false while there is none (the mod has
// not run in Claude Code yet).
func ReadCatalog(path string) (c Catalog, ok bool, err error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Catalog{}, false, nil
	}
	if err != nil {
		return Catalog{}, false, err
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return Catalog{}, false, err
	}
	return c, true, nil
}

// Placement is where the drawer sits: beside the transcript, or above the
// prompt on a narrow terminal.
type Placement string

const (
	Dock   Placement = "dock"
	Inline Placement = "inline"
)

// Config is what the TUI writes and every open session reads within a second
// (hooks/layout.ts): the ids on in each placement. A placement left out takes
// the catalog's defaults: it is written as null, while [] is every item off.
// Alias is one more name for /ccdrawu (hooks/commands.ts), "" for none. Pet is
// the drawer's pet, and Bar how it draws its bars. Theme is the Claude Code
// theme ccdrawu selected, if any; the mod never reads it.
type Config struct {
	Dock   []string `json:"dock"`
	Inline []string `json:"inline"`
	Alias  string   `json:"alias,omitempty"`
	Pet    Pet      `json:"pet,omitzero"`
	Bar    string   `json:"bar,omitempty"`
	Theme  Theme    `json:"theme,omitzero"`
}

// BarStyles are the ways the mod draws a bar (bar in hooks/format.ts): Nerd
// Font squares, the first and the one for "", or a line for a terminal with no
// Nerd Font.
var BarStyles = []string{"squares", "line"}

// Pet is the pet as the mod draws it (hooks/pets.ts): Species one of
// PetSpecies ("" for the first, the cat). Its coat is the mod's to pick, a new
// one each session.
type Pet struct {
	Species string `json:"species,omitempty"`
}

// PetSpecies are the animals the mod draws, in the order of its SPECIES
// (hooks/pets.ts); one it does not know is the cat.
var PetSpecies = []string{"cat", "dog", "rabbit", "duck", "penguin"}

// Theme is ccdrawu's Claude Code theme (internal/claude): Base, a built-in
// theme ("" for none, the user's own theme as it was), with Background over
// the drawer ("" for the base's own). Before is the theme selected until
// ccdrawu's was, to go back to ("" when none was set).
type Theme struct {
	Base       string `json:"base,omitempty"`
	Background string `json:"background,omitempty"`
	Before     string `json:"before,omitempty"`
}

func (c *Config) ids(p Placement) *[]string {
	if p == Dock {
		return &c.Dock
	}
	return &c.Inline
}

// ReadConfig reads the config; a file not there is an empty config.
func ReadConfig(path string) (Config, error) {
	var c Config
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	// A config that is not ours is not read and not overwritten blindly: the
	// error says so.
	if err := json.Unmarshal(data, &c); err != nil {
		return c, err
	}
	return c, nil
}

// On reports whether the item is on in p: the config's list when it has one,
// else the catalog's default.
func (c Config) On(p Placement, it Item) bool {
	list := *c.ids(p)
	if list == nil {
		if p == Dock {
			return it.Dock
		}
		return it.Inline
	}
	return slices.Contains(list, it.ID)
}

// Set turns the item on or off in p. The first change in a placement writes
// out the catalog's defaults for it; ids the catalog does not know (from a
// newer mod) are kept.
func (c *Config) Set(p Placement, cat Catalog, id string, on bool) {
	list := c.ids(p)
	if *list == nil {
		next := []string{}
		for _, it := range cat.Items {
			if c.On(p, it) {
				next = append(next, it.ID)
			}
		}
		*list = next
	}
	*list = slices.DeleteFunc(*list, func(s string) bool { return s == id })
	if on {
		*list = append(*list, id)
	}
	// In the catalog's order, then the unknown ones as they were
	order := map[string]int{}
	for i, it := range cat.Items {
		order[it.ID] = i
	}
	slices.SortStableFunc(*list, func(a, b string) int {
		ia, okA := order[a]
		ib, okB := order[b]
		switch {
		case okA && okB:
			return ia - ib
		case okA:
			return -1
		case okB:
			return 1
		}
		return 0
	})
}

// WriteConfig writes the whole file at once (a temporary file renamed over
// it), so a session never reads half of it.
func WriteConfig(path string, c Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
