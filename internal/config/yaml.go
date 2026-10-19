package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

// oldConfig is the config.json 0.1.4 and before wrote where config.yaml is now.
func oldConfig(path string) string { return filepath.Join(filepath.Dir(path), "config.json") }

// A Moved is what Move took to its new place.
type Moved struct{ From, To string }

// Move takes what 0.1.4 and before kept elsewhere to where tdp E3 has it now,
// once, where the new place does not have it yet: config.json, written again
// as config.yaml beside it; and what is in ~/.ccdrawu, the catalog and the
// statusline's values the mod writes, into the cache folder. An error names
// both places of what could not move, and ccdrawu goes on reading the old
// place (ReadConfig, Locations.Catalog).
func Move(loc Locations) ([]Moved, error) {
	var moved []Moved
	var errs []error
	if old := oldConfig(loc.ConfigFile); fileThere(old) && !fileThere(loc.ConfigFile) {
		c, err := ReadConfig(loc.ConfigFile)
		if err == nil {
			err = WriteConfig(loc.ConfigFile, c)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("%s to %s: %w", old, loc.ConfigFile, err))
		} else {
			moved = append(moved, Moved{old, loc.ConfigFile})
		}
	}
	// Each thing in it on its own: the mod writes the catalog to the cache
	// folder as a session starts, often before ccdrawu first runs
	if entries, err := os.ReadDir(loc.OldData); err == nil {
		var failed error
		took := false
		for _, e := range entries {
			from, to := filepath.Join(loc.OldData, e.Name()), filepath.Join(loc.Cache, e.Name())
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
			errs = append(errs, fmt.Errorf("%s to %s: %w", loc.OldData, loc.Cache, failed))
		} else if took {
			moved = append(moved, Moved{loc.OldData, loc.Cache})
		}
		// Gone once all of it moved
		_ = os.Remove(loc.OldData)
	}
	return moved, errors.Join(errs...)
}
