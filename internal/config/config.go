package config

import (
	"bytes"
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
	Short       string `json:"short"`
	Name        string `json:"name"`
	Group       string `json:"group"`
	Description string `json:"description"`
	Dock        bool   `json:"dock"`
	Inline      bool   `json:"inline"`
	Statusline  bool   `json:"statusline"`
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

// Placement is where the items show: the drawer beside the transcript, or
// above the prompt on a narrow terminal, or the statusline, Claude Code's
// status line under the prompt.
type Placement string

const (
	Dock   Placement = "dock"
	Inline Placement = "inline"
	Status Placement = "statusline"
)

// Config is what the TUI writes and every open session reads within a second
// (hooks/layout.ts): the ids on in each placement. A placement left out takes
// the catalog's defaults: it is written as null, while [] is every item off.
// Alias is one more name for /ccdrawu (hooks/commands.ts), "" for none. Pet is
// the drawer's pet, and Bar how it draws its bars. Drawer is whether the
// drawer opens by itself. Theme is the Claude Code theme ccdrawu selected, if
// any; the mod never reads it. Statusline is `ccdrawu statusline`; the mod
// reads only whether it is on.
type Config struct {
	Dock       []string   `json:"dock"`
	Inline     []string   `json:"inline"`
	Alias      string     `json:"alias,omitempty"`
	Pet        Pet        `json:"pet,omitzero"`
	Bar        string     `json:"bar,omitempty"`
	Drawer     Drawer     `json:"drawer,omitzero"`
	Theme      Theme      `json:"theme,omitzero"`
	Statusline Statusline `json:"statusline,omitzero"`
}

// Drawer is the drawer itself, on unless Off: off, sessions start without it
// and open ones close it, while /ccdrawu still opens and closes it.
type Drawer struct {
	Off bool `json:"off,omitempty"`
}

// Statusline is `ccdrawu statusline` as Claude Code's status line. On says
// whether it is: the mod keeps each session's values for it while it is.
// When is when it shows while on: always (""), or only while the drawer is not
// beside the transcript (WhenInline). Items are the ids it shows, in order; nil takes the catalog's defaults, as
// the drawer's placements do. Format is how each item reads (one of Formats,
// "" for the first); Color how it is colored (one of Colors, "" for the
// first); Separator what stands between items ("" for "|"); Refresh how many
// seconds Claude Code waits to run it again while nothing happens (0 for
// DefaultRefresh). Before is the statusLine Claude Code had until ccdrawu's,
// to put back as it was written: JSON null when it had none.
type Statusline struct {
	On        bool     `json:"on"`
	When      string   `json:"when,omitempty"`
	Items     []string `json:"items"`
	Format    string   `json:"format,omitempty"`
	Color     string   `json:"color,omitempty"`
	Separator string   `json:"separator,omitempty"`
	Refresh   int      `json:"refresh,omitempty"`
	Before    Kept     `json:"before,omitempty"`
}

// Kept is a JSON value kept byte for byte as it was written elsewhere.
// config.json holds it as a string: held as JSON, writing the config would
// lay it out anew. One held as JSON, as ccdrawu 0.1.1 wrote it, reads too,
// compacted, its own spacing gone already.
type Kept json.RawMessage

func (k Kept) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(k))
}

func (k *Kept) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*k = Kept(s)
		return nil
	}
	var b bytes.Buffer
	if err := json.Compact(&b, data); err != nil {
		return err
	}
	*k = Kept(b.Bytes())
	return nil
}

// WhenInline is a statusline on only while the drawer is not beside the
// transcript: above the prompt, waiting for a wider terminal, or closed.
const WhenInline = "inline"

// Formats are how an item reads in the statusline: its label and value
// ("Context 42%"), its short label and value ("ctx 42%"), or the value alone.
var Formats = []string{"label", "short", "value"}

// Colors are how the statusline is colored: each piece in the drawer's
// colors, or none, which Claude Code draws all in its dim color.
var Colors = []string{"highlighted", "dimmed"}

// DefaultRefresh is the statusline's refresh in seconds when none is set.
const DefaultRefresh = 10

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
	switch p {
	case Dock:
		return &c.Dock
	case Status:
		return &c.Statusline.Items
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
		switch p {
		case Dock:
			return it.Dock
		case Status:
			return it.Statusline
		}
		return it.Inline
	}
	return slices.Contains(list, it.ID)
}

// Set turns the item on or off in p. The first change in a placement writes
// out the catalog's defaults for it. The list keeps its order, the order the
// drawer draws it in, and the ids the catalog does not know (from a newer
// mod): an item turned on goes after the nearest one before it in the catalog
// that is on, so a list never reordered stays in the catalog's order.
func (c *Config) Set(p Placement, cat Catalog, id string, on bool) {
	list := c.written(p, cat)
	*list = slices.DeleteFunc(*list, func(s string) bool { return s == id })
	if !on {
		return
	}
	at := 0
	i := slices.IndexFunc(cat.Items, func(it Item) bool { return it.ID == id })
	for k := i - 1; k >= 0; k-- {
		if j := slices.Index(*list, cat.Items[k].ID); j >= 0 {
			at = j + 1
			break
		}
	}
	*list = slices.Insert(*list, at, id)
}

// Reorder puts ids, items on in p, in that order, in the places they had in
// the list; the other ids (the pet, ones the catalog does not know) keep
// theirs.
func (c *Config) Reorder(p Placement, cat Catalog, ids []string) {
	list := c.written(p, cat)
	next := slices.Clone(*list)
	k := 0
	for i, id := range next {
		if k < len(ids) && slices.Contains(ids, id) {
			next[i] = ids[k]
			k++
		}
	}
	*list = next
}

// written is p's list, the catalog's defaults written out when it has none.
func (c *Config) written(p Placement, cat Catalog) *[]string {
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
	return list
}

// Shown is the items on in p in the order the drawer draws them: the config's
// list when it has one, else the catalog's defaults in its order. Ids the
// catalog does not know are left out.
func (c Config) Shown(p Placement, cat Catalog) []Item {
	list := *c.ids(p)
	if list == nil {
		return slices.DeleteFunc(slices.Clone(cat.Items), func(it Item) bool { return !c.On(p, it) })
	}
	var items []Item
	for _, id := range list {
		if i := slices.IndexFunc(cat.Items, func(it Item) bool { return it.ID == id }); i >= 0 {
			items = append(items, cat.Items[i])
		}
	}
	return items
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
