package claude

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// ThemeSetting is settings.json's theme while ccdrawu's is selected: a custom
// theme, themes/ccdrawu.json, standing on a built-in one with the drawer's
// background. The background is composerSidebarBackground, the token the
// docked pane's frame takes its color from; only a custom theme sets it.
const ThemeSetting = "custom:ccdrawu"

// Themes are the built-in themes the ccdrawu theme can stand on, as Claude
// Code 2.1.299 names them.
var Themes = []string{"dark", "light", "dark-daltonized", "light-daltonized", "dark-ansi", "light-ansi"}

// ApplyTheme writes the ccdrawu theme, base with background ("" for the base's
// own), and selects it. When that changes the selection, switched is true and
// before is the theme selected until now ("" for none), to go back to.
func ApplyTheme(dir, base, background string) (before string, switched bool, err error) {
	theme := map[string]any{"name": "ccdrawu", "base": base}
	if background != "" {
		theme["overrides"] = map[string]string{"composerSidebarBackground": background}
	}
	data, _ := json.MarshalIndent(theme, "", "  ")
	if err := writeFile(filepath.Join(dir, "themes", "ccdrawu.json"), append(data, '\n')); err != nil {
		return "", false, err
	}
	settings, err := readSettings(dir)
	if err != nil {
		return "", false, err
	}
	before = stringAt(settings, "theme")
	if before == ThemeSetting {
		return "", false, nil
	}
	if settings, err = setString(settings, "theme", ThemeSetting); err != nil {
		return "", false, err
	}
	return before, true, writeFile(filepath.Join(dir, "settings.json"), settings)
}

// RestoreTheme selects before again ("" takes the theme setting out) and
// removes the ccdrawu theme. A theme chosen since, in Claude Code's /theme, is
// left as it is.
func RestoreTheme(dir, before string) error {
	settings, err := readSettings(dir)
	if err != nil {
		return err
	}
	if stringAt(settings, "theme") == ThemeSetting {
		if before == "" {
			settings, err = removeKey(settings, "theme")
		} else {
			settings, err = setString(settings, "theme", before)
		}
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(dir, "settings.json"), settings); err != nil {
			return err
		}
	}
	if err := os.Remove(filepath.Join(dir, "themes", "ccdrawu.json")); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// readSettings reads settings.json; a file not there is an empty object.
func readSettings(dir string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return []byte("{}\n"), nil
	}
	return data, err
}
