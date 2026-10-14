// Package config is what ccdrawu keeps and where (tdp E3): the catalog the mod
// writes and the config the TUI writes.
package config

import (
	"os"
	"path/filepath"
)

// Locations are where ccdrawu keeps its files (tdp E3), the same places the mod
// reads (hooks/paths.ts): Data is CCDRAWU__DATA, else ~/.ccdrawu (the catalog);
// ConfigFile is config.json in CCDRAWU__CONFIG, else $XDG_CONFIG_HOME/ccdrawu,
// else ~/.config/ccdrawu. A variable set but empty counts as unset.
type Locations struct {
	Data, ConfigFile string
}

func locate(getenv func(string) string, home string) Locations {
	data := getenv("CCDRAWU__DATA")
	if data == "" {
		data = filepath.Join(home, ".ccdrawu")
	}
	config := getenv("CCDRAWU__CONFIG")
	if config == "" {
		if xdg := getenv("XDG_CONFIG_HOME"); xdg != "" {
			config = filepath.Join(xdg, "ccdrawu")
		} else {
			config = filepath.Join(home, ".config", "ccdrawu")
		}
	}
	return Locations{Data: data, ConfigFile: filepath.Join(config, "config.json")}
}

func UserLocations() (Locations, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Locations{}, err
	}
	return locate(os.Getenv, home), nil
}
