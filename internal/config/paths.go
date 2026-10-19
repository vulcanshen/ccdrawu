// Package config is what ccdrawu keeps and where (tdp E3): the catalog the mod
// writes and the config the TUI writes.
package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/vulcanshen/ccdrawu/internal/claude"
)

// Locations are where ccdrawu keeps its files (tdp E3), the same places the mod
// uses (hooks/paths.ts). A variable set but empty counts as unset.
//
//   - ConfigFile is config.yaml in CCDRAWU__CONFIG, else $XDG_CONFIG_HOME/ccdrawu,
//     else ~/.config/ccdrawu: what you pick in ccdrawu.
//   - Cache is CCDRAWU__CACHE, else the system's cache folder's ccdrawu: the
//     catalog and each session's statusline values, which the mod writes anew,
//     so a cache (tdp E3: what the app makes again when it is deleted).
//   - Claude is Claude Code's own folder, where the theme goes.
//   - OldData is where 0.1.4 and before kept the catalog and the sessions:
//     CCDRAWU__DATA, else ~/.ccdrawu. Move takes it to Cache.
type Locations struct {
	ConfigFile, Cache, Claude, OldData string
}

func locate(getenv func(string) string, home, goos string) Locations {
	config := getenv("CCDRAWU__CONFIG")
	if config == "" {
		if xdg := getenv("XDG_CONFIG_HOME"); xdg != "" {
			config = filepath.Join(xdg, "ccdrawu")
		} else {
			config = filepath.Join(home, ".config", "ccdrawu")
		}
	}
	old := getenv("CCDRAWU__DATA")
	if old == "" {
		old = filepath.Join(home, ".ccdrawu")
	}
	return Locations{
		ConfigFile: filepath.Join(config, "config.yaml"),
		Cache:      cacheDir(getenv, home, goos),
		Claude:     claude.Dir(getenv, home),
		OldData:    old,
	}
}

// cacheDir is the system's cache folder's ccdrawu (tdp E3), as os.UserCacheDir
// has it: %LOCALAPPDATA% on Windows, ~/Library/Caches on macOS, else
// $XDG_CACHE_HOME when it is an absolute path, else ~/.cache.
func cacheDir(getenv func(string) string, home, goos string) string {
	if dir := getenv("CCDRAWU__CACHE"); dir != "" {
		return dir
	}
	switch goos {
	case "windows":
		if dir := getenv("LOCALAPPDATA"); dir != "" {
			return filepath.Join(dir, "ccdrawu")
		}
		return filepath.Join(home, "AppData", "Local", "ccdrawu")
	case "darwin":
		return filepath.Join(home, "Library", "Caches", "ccdrawu")
	}
	if dir := getenv("XDG_CACHE_HOME"); filepath.IsAbs(dir) {
		return filepath.Join(dir, "ccdrawu")
	}
	return filepath.Join(home, ".cache", "ccdrawu")
}

func UserLocations() (Locations, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Locations{}, err
	}
	return locate(os.Getenv, home, runtime.GOOS), nil
}

// Catalog is the catalog's file: in Cache, else where 0.1.4 and before kept it
// until it is moved (Move).
func (l Locations) Catalog() string {
	path := filepath.Join(l.Cache, "catalog.json")
	if old := filepath.Join(l.OldData, "catalog.json"); !fileThere(path) && fileThere(old) {
		return old
	}
	return path
}

func fileThere(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
