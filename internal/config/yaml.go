package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// The config is config.yaml (tdp E3). Its fields keep their JSON names and
// rules (omitempty, a list left unset as null, Kept as a string): a Config goes
// through JSON, and the JSON through YAML, in its own order.

// yamlFromJSON writes a JSON document as block YAML, in its own order.
func yamlFromJSON(data []byte) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	block(&doc)
	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// block drops the flow and the quotes JSON wrote everything with: the encoder
// quotes a string only where it would read as something else.
func block(n *yaml.Node) {
	n.Style = 0
	for _, c := range n.Content {
		block(c)
	}
}

// jsonFromYAML reads a YAML document as JSON, for a Config to take.
func jsonFromYAML(data []byte) ([]byte, error) {
	var v any
	if err := yaml.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	if v == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(v)
}

// Source is the file ReadConfig reads for path: config.yaml, or until there
// is one, the config.json beside it.
func Source(path string) string {
	if old := oldConfig(path); !fileThere(path) && fileThere(old) {
		return old
	}
	return path
}

// oldConfig is the config.json 0.1.4 and before wrote where config.yaml is now.
func oldConfig(path string) string { return filepath.Join(filepath.Dir(path), "config.json") }

// A Moved is what Move took to its new place: What names the items taken out
// of From, "" where all of it went.
type Moved struct{ From, To, What string }

// Move takes what 0.1.4 and before kept elsewhere to where tdp E3 has it now,
// once, where the new place does not have it yet: config.json, written again
// as config.yaml beside it; the originals it kept, into originals.yaml; and
// what else is in ~/.ccdrawu, the catalog and the statusline's values the mod
// writes, into the cache folder. An error names both places of what could not
// move and says which one ccdrawu reads: the old place, until it moves
// (ReadConfig, ReadOriginals, Locations.Catalog), or the config.yaml a
// config.json is left beside, for the user to merge.
func Move(loc Locations) ([]Moved, error) {
	var moved []Moved
	var errs []error
	if old := oldConfig(loc.ConfigFile); fileThere(old) {
		if fileThere(loc.ConfigFile) {
			// Both there: config.json stays where it is, for the user to
			// merge, and says so each start until then (tdp E3)
			errs = append(errs, fmt.Errorf("%s is there beside %s. ccdrawu reads config.yaml: put what you want of config.json in it yourself, then remove config.json.", old, loc.ConfigFile))
		} else {
			c, err := ReadConfig(loc.ConfigFile)
			if err == nil {
				err = WriteConfig(loc.ConfigFile, c)
			}
			if err != nil {
				errs = append(errs, cantMove(old, loc.ConfigFile, err))
			} else {
				moved = append(moved, Moved{From: old, To: loc.ConfigFile})
			}
		}
	}
	// Out of a config.yaml there, readable: one that can't be read is moved
	// aside first (tdp E3)
	if c, err := ReadConfig(loc.ConfigFile); err == nil && fileThere(loc.ConfigFile) {
		m, err := moveOriginals(loc, c)
		if err != nil {
			errs = append(errs, err)
		}
		if m != nil {
			moved = append(moved, *m)
		}
	}
	// Each thing in it on its own, the originals aside: the mod writes the
	// catalog to the cache folder as a session starts, often before ccdrawu
	// first runs
	if entries, err := os.ReadDir(loc.Data); err == nil {
		var failed error
		took := false
		for _, e := range entries {
			if e.Name() == originalsFile {
				continue
			}
			from, to := filepath.Join(loc.Data, e.Name()), filepath.Join(loc.Cache, e.Name())
			// The mod wrote it anew there already: the old one is a stale cache
			if fileThere(to) {
				_ = os.RemoveAll(from)
				continue
			}
			err := os.MkdirAll(loc.Cache, 0o755)
			if err == nil {
				err = os.Rename(from, to)
			}
			if err != nil {
				failed = err
				continue
			}
			took = true
		}
		if failed != nil {
			errs = append(errs, cantMove(loc.Data, loc.Cache, failed))
		} else if took {
			moved = append(moved, Moved{From: loc.Data, To: loc.Cache})
		}
		// Gone once all of it moved, unless it keeps the originals
		_ = os.Remove(loc.Data)
	}
	return moved, paragraph(errs)
}

// cantMove is what could not move from from to to, and why: ccdrawu reads it
// where it is until it moves (tdp E3).
func cantMove(from, to string, err error) error {
	return fmt.Errorf("%s to %s: %w. ccdrawu reads it where it is until it moves: move it there yourself.", from, to, err)
}

// paragraph is errs in one paragraph, sentence after sentence, as an error
// popup is one (tdp components/dialog/note): nil for none.
func paragraph(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	msgs := make([]string, len(errs))
	for i, err := range errs {
		msgs[i] = err.Error()
	}
	return errors.New(strings.Join(msgs, " "))
}
