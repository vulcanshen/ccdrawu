package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/vulcanshen/terminu-go"

	"github.com/vulcanshen/ccdrawu/internal/claude"
)

// A config.yaml that reads as YAML, a mapping, with values ccdrawu can't use
// is not one it can't read (tdp E3, a value that does not fit): each such
// value reads as its default, a Misfit naming the key, what the file has and
// what is used now. Kinds of value it can't use: a name it does not know (an
// animal, a bar, a format, a color, a theme, an item the catalog does not
// have), a color that is not #rrggbb, a number of seconds below 0 or past 60,
// an alias the mod does not take, a separator with a control character, and a
// value of the wrong kind (a word where a number goes). A key that is not
// ccdrawu's at all is left out, as before.
// TODO(tdp #21): tdp says nothing of a key that was never the app's (one
// mistyped); writing the file again leaves it out, the file as it was kept in
// the .bak when values did not fit.

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// MaxRefresh is the most seconds the statusline's Refresh takes.
const MaxRefresh = 60

// fit reads the config's values one by one, each that does not fit its
// default, as a Misfit.
type fit struct {
	misfits []terminu.Misfit
	known   []string // the catalog's ids; nil while there is no catalog
}

func (f *fit) misfit(key string, raw json.RawMessage, now string) {
	was := string(raw)
	var s string
	if json.Unmarshal(raw, &s) == nil {
		was = s
	}
	f.misfits = append(f.misfits, terminu.Misfit{Key: key, Was: was, Now: now})
}

// object is the mapping at key in m, nil when there is none; one that is not
// a mapping does not fit, all its values their defaults.
func (f *fit) object(m map[string]json.RawMessage, key string) map[string]json.RawMessage {
	raw, ok := m[key]
	if !ok || string(raw) == "null" {
		return nil
	}
	var obj map[string]json.RawMessage
	if json.Unmarshal(raw, &obj) != nil {
		f.misfit(key, raw, "default")
		return nil
	}
	return obj
}

// value reads the value at name in m into into, as key; one of another kind
// does not fit, now its default. It reports whether there was one to read.
func (f *fit) value(m map[string]json.RawMessage, key, name string, into any, now string) bool {
	raw, ok := m[name]
	if !ok || string(raw) == "null" {
		return false
	}
	if json.Unmarshal(raw, into) != nil {
		f.misfit(key, raw, now)
		return false
	}
	return true
}

// oneOf reads a name that is one of names, "" for the first; another does not
// fit, now the first.
func (f *fit) oneOf(m map[string]json.RawMessage, key, name string, into *string, names []string, now string) {
	var s string
	if !f.value(m, key, name, &s, now) {
		return
	}
	if s != "" && !slices.Contains(names, s) {
		f.misfit(key, m[name], now)
		return
	}
	*into = s
}

// flag reads a yes or no.
func (f *fit) flag(m map[string]json.RawMessage, key, name string, into *bool) {
	f.value(m, key, name, into, "false")
}

// ids reads a placement's list of item ids, nil for the catalog's defaults:
// one that is not a list does not fit, now the defaults; an entry that is not
// an id the catalog has is taken out, while there is a catalog to know them by.
func (f *fit) ids(m map[string]json.RawMessage, key, name string, into *[]string) {
	var entries []json.RawMessage
	if !f.value(m, key, name, &entries, "default") {
		return
	}
	list := []string{}
	for _, raw := range entries {
		var id string
		if json.Unmarshal(raw, &id) != nil || f.known != nil && !slices.Contains(f.known, id) {
			f.misfit(key+" "+string(raw), raw, "")
			continue
		}
		list = append(list, id)
	}
	*into = list
}

// readFit reads the config at path, as ReadConfig does, with the values that
// do not fit their defaults, each a Misfit. known are the catalog's ids, nil
// while there is none. An error is a file that can't be read: not YAML, or not
// a mapping (tdp E3, can't be read).
func readFit(path string, known []string) (Config, []terminu.Misfit, error) {
	var c Config
	data, err := configData(path)
	if err != nil || data == nil {
		return c, nil, err
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return c, nil, fmt.Errorf("not a config: %w", err)
	}
	f := &fit{known: known}
	f.ids(top, "dock", "dock", &c.Dock)
	f.ids(top, "inline", "inline", &c.Inline)
	var alias string
	if f.value(top, "alias", "alias", &alias, "none") {
		if _, err := CleanAlias(alias); err != nil {
			f.misfit("alias", top["alias"], "none")
		} else {
			c.Alias = alias
		}
	}
	f.oneOf(f.object(top, "pet"), "pet.species", "species", &c.Pet.Species, PetSpecies, PetSpecies[0])
	f.oneOf(top, "bar", "bar", &c.Bar, BarStyles, BarStyles[0])
	drawer := f.object(top, "drawer")
	f.flag(drawer, "drawer.sidebarOff", "sidebarOff", &c.Drawer.SidebarOff)
	f.flag(drawer, "drawer.inlineOff", "inlineOff", &c.Drawer.InlineOff)
	f.flag(drawer, "drawer.off", "off", &c.Drawer.Off)
	if c.Drawer.Off {
		c.Drawer = Drawer{SidebarOff: true, InlineOff: true}
	}
	theme := f.object(top, "theme")
	f.oneOf(theme, "theme.base", "base", &c.Theme.Base, claude.Themes, "off")
	var background string
	if f.value(theme, "theme.background", "background", &background, "none") {
		if background != "" && !hexColor.MatchString(background) {
			f.misfit("theme.background", theme["background"], "none")
		} else {
			c.Theme.Background = background
		}
	}
	f.value(theme, "theme.before", "before", &c.Theme.Before, "none")
	st := f.object(top, "statusline")
	f.flag(st, "statusline.on", "on", &c.Statusline.On)
	f.oneOf(st, "statusline.when", "when", &c.Statusline.When, []string{WhenNarrow}, "always")
	f.ids(st, "statusline.items", "items", &c.Statusline.Items)
	f.oneOf(st, "statusline.format", "format", &c.Statusline.Format, Formats, Formats[0])
	f.oneOf(st, "statusline.color", "color", &c.Statusline.Color, Colors, Colors[0])
	var sep string
	if f.value(st, "statusline.separator", "separator", &sep, "|") {
		if strings.ContainsFunc(sep, unicode.IsControl) {
			f.misfit("statusline.separator", st["separator"], "|")
		} else {
			c.Statusline.Separator = sep
		}
	}
	var refresh int
	if f.value(st, "statusline.refresh", "refresh", &refresh, strconv.Itoa(DefaultRefresh)) {
		if refresh < 0 || refresh > MaxRefresh {
			f.misfit("statusline.refresh", st["refresh"], strconv.Itoa(DefaultRefresh))
		} else {
			c.Statusline.Refresh = refresh
		}
	}
	f.value(st, "statusline.before", "before", &c.Statusline.Before, "none")
	return c, f.misfits, nil
}

// Settled is what a start found in config.yaml (tdp E3, a value that does not
// fit): the values ccdrawu could not use, each now its default; Copy, where
// the file was kept as it was before it was written again with them; and Err,
// why that could not be done, the file then left as it was and the defaults
// used for this run only.
type Settled struct {
	Misfits []terminu.Misfit
	Copy    string
	Err     error
}

// Settle reads config.yaml as a start does (tdp E3): values that do not fit
// are their defaults, the file copied aside (terminu.KeepAside) and written
// again with them. A theme that does not fit is off, and off gives Claude Code
// its own theme back, as Claude theme off does. The error is a file that
// can't be read, for the app to ask about, or with no screen to say on stderr.
func Settle(loc Locations) (Config, Settled, error) {
	var known []string
	if cat, ok, err := ReadCatalog(loc.Catalog()); err == nil && ok {
		for _, it := range cat.Items {
			known = append(known, it.ID)
		}
	}
	c, misfits, err := readFit(loc.ConfigFile, known)
	// Only config.yaml is settled: the config.json of 0.1.4 and before is
	// Move's to take
	if err != nil || len(misfits) == 0 || !fileThere(loc.ConfigFile) {
		return c, Settled{}, err
	}
	s := Settled{Misfits: misfits}
	copy, err := terminu.KeepAside(loc.ConfigFile)
	if err != nil {
		s.Err = err
		return c, s, nil
	}
	s.Copy = copy
	if err := WriteConfig(loc.ConfigFile, c); err != nil {
		s.Err = fmt.Errorf("%s could not be written: %w; yours is in %s", filepath.Base(loc.ConfigFile), err, filepath.Base(copy))
		return c, s, nil
	}
	// What the defaults are in Claude Code too: the theme off gives it its own
	// back, as Claude theme off does; the refresh is its status line's
	misfit := func(key string) bool {
		return slices.ContainsFunc(misfits, func(m terminu.Misfit) bool { return m.Key == key })
	}
	if misfit("theme.base") {
		if err := themeOff(loc, c); err != nil {
			s.Err = err
		}
	}
	if misfit("statusline.refresh") && c.Statusline.On {
		if _, ours, err := claude.CurrentStatusline(loc.Claude); err == nil && ours {
			if _, _, err := claude.ApplyStatusline(loc.Claude, DefaultRefresh); err != nil {
				s.Err = fmt.Errorf("can't set the statusline's refresh: %w", err)
			}
		}
	}
	return c, s, nil
}

// Lines are what a start with no screen says on stderr of what Settle did, a
// line each: the values kept aside and written again with the defaults
// (terminu.DidNotFitText), and why that could not be done (tdp E3, a value
// that does not fit; several at one start).
func (s Settled) Lines(path string) []string {
	var lines []string
	if text := terminu.DidNotFitText(path, s.Copy, s.Misfits...); text != "" {
		lines = append(lines, text)
	}
	if s.Err != nil {
		lines = append(lines, s.Err.Error())
	}
	return lines
}

// themeOff gives Claude Code its own theme back and forgets it, as Claude
// theme off does in the TUI.
func themeOff(loc Locations, c Config) error {
	orig, err := ReadOriginals(loc.Originals(), c)
	if err != nil {
		return err
	}
	if err := claude.RestoreTheme(loc.Claude, orig.Theme); err != nil {
		return fmt.Errorf("can't put your own theme back: %w", err)
	}
	orig.Theme = ""
	return WriteOriginals(loc.Originals(), orig)
}
