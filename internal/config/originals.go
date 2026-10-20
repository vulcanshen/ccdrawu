package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const originalsFile = "originals.yaml"

// Originals are what ccdrawu replaced in Claude Code, kept to put back as they
// were: Theme is the theme selected until ccdrawu's was ("" when none was
// set), Statusline the statusLine Claude Code had until ccdrawu's (JSON null
// when it had none, nil while ccdrawu's is not set). Gone, they can't be found
// again, so they are data (tdp E3, CD55): originals.yaml in the data folder.
// The config kept them until then, as theme.before and statusline.before
// (Theme.Before, Statusline.Before), and Move takes them from there.
type Originals struct {
	Theme      string `json:"theme,omitempty"`
	Statusline Kept   `json:"statusline,omitempty"`
}

// Equal says whether o and p keep the same.
func (o Originals) Equal(p Originals) bool {
	return o.Theme == p.Theme && bytes.Equal(o.Statusline, p.Statusline)
}

// ReadOriginals reads originals.yaml at path, none while there is none; one
// that c still keeps in its old place is read from there, as it is until
// Move takes it (tdp E3).
func ReadOriginals(path string, c Config) (Originals, error) {
	var o Originals
	data, err := os.ReadFile(path)
	if err == nil {
		if data, err = jsonFromYAML(data); err == nil {
			err = json.Unmarshal(data, &o)
		}
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Originals{}, err
	}
	if c.Theme.Before != "" {
		o.Theme = c.Theme.Before
	}
	if c.Statusline.Before != nil {
		o.Statusline = c.Statusline.Before
	}
	return o, nil
}

// WriteOriginals writes o to path; with nothing kept the file goes, and the
// data folder with it once that leaves it empty.
func WriteOriginals(path string, o Originals) error {
	if o.Equal(Originals{}) {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		_ = os.Remove(filepath.Dir(path))
		return nil
	}
	data, err := json.Marshal(o)
	if err != nil {
		return err
	}
	if data, err = yamlFromJSON(data); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// moveOriginals takes the originals c keeps, each on its own, to
// originals.yaml, and writes the config without them. One originals.yaml has
// already, the same, only leaves the config; one it has otherwise stays in
// the config, read from there, and the error names both places (tdp E3).
func moveOriginals(loc Locations, c Config) (*Moved, error) {
	if c.Theme.Before == "" && c.Statusline.Before == nil {
		return nil, nil
	}
	o, err := ReadOriginals(loc.Originals(), Config{})
	if err != nil {
		return nil, fmt.Errorf("%s to %s: %w", loc.ConfigFile, loc.Originals(), err)
	}
	var took, kept []string
	if b := c.Theme.Before; b != "" {
		switch {
		case o.Theme == "":
			o.Theme, c.Theme.Before = b, ""
			took = append(took, "theme.before")
		case o.Theme == b:
			c.Theme.Before = ""
		default:
			kept = append(kept, "theme.before")
		}
	}
	if b := c.Statusline.Before; b != nil {
		switch {
		case o.Statusline == nil:
			o.Statusline, c.Statusline.Before = b, nil
			took = append(took, "statusline.before")
		case bytes.Equal(o.Statusline, b):
			c.Statusline.Before = nil
		default:
			kept = append(kept, "statusline.before")
		}
	}
	// The originals first: written twice is better than gone
	if len(took) > 0 {
		if err := WriteOriginals(loc.Originals(), o); err != nil {
			return nil, fmt.Errorf("%s to %s: %w", loc.ConfigFile, loc.Originals(), err)
		}
	}
	if err := WriteConfig(loc.ConfigFile, c); err != nil {
		return nil, fmt.Errorf("%s to %s: %w", loc.ConfigFile, loc.Originals(), err)
	}
	if len(kept) > 0 {
		err = fmt.Errorf("%s in %s to %s: it has one already", strings.Join(kept, " and "), loc.ConfigFile, loc.Originals())
	}
	if len(took) == 0 {
		return nil, err
	}
	return &Moved{From: loc.ConfigFile, To: loc.Originals(), What: strings.Join(took, " and ")}, err
}
