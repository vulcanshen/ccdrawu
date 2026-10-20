package ui

import (
	"cmp"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vulcanshen/ccdrawu/internal/claude"
	"github.com/vulcanshen/ccdrawu/internal/config"
	"github.com/vulcanshen/ccdrawu/internal/version"
	terminu "github.com/vulcanshen/terminu-go"
)

// The preferences past the alias: Claude Code's theme with the drawer's
// background (internal/claude, theme), how the drawer draws its bars, the pet,
// and whether the plugin is in Claude Code.

const glyphPalette = "\U000F03D8" // nf-md-palette, the background picker's title

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// themeField picks the built-in theme ccdrawu's stands on, or off: the user's
// own theme as it was. More than five to pick from, so a select (tdp
// components/input/radio); it takes effect when picked.
func (d *drawer) themeField() *terminu.Field {
	options := append([]string{"off"}, claude.Themes...)
	return &terminu.Field{Label: "Claude theme", Value: func() string { return cmp.Or(d.cfg.Theme.Base, "off") }, Edit: func() tea.Cmd {
		var note string
		return d.app.Select(terminu.Select{
			Title: "Claude theme", Glyph: glyphPalette, Options: options, Selected: max(0, slices.Index(options, d.cfg.Theme.Base)),
			Commit: func(i int) (err error) {
				d.cfg.Theme.Base = ""
				if i > 0 {
					d.cfg.Theme.Base = options[i]
				}
				note, err = d.setTheme()
				return err
			},
			Done: func(int) tea.Cmd { return d.app.Toast(terminu.Success, note) },
		})
	}}
}

// backgroundField is the drawer's background over that theme, in a color
// picker; none keeps the theme's own. Off with the theme. Its value is a
// swatch and the hex (tdp components/input/color-picker).
func (d *drawer) backgroundField() *terminu.Field {
	return &terminu.Field{
		Label: "Drawer background", Empty: "the theme's own",
		Value: func() string {
			if !hexColor.MatchString(d.cfg.Theme.Background) {
				return d.cfg.Theme.Background
			}
			return terminu.HexText(terminu.Hex(d.cfg.Theme.Background))
		},
		// The swatch in its colour, drawn by the lib (tdp components, control characters)
		Styled: func() []terminu.Span {
			if !hexColor.MatchString(d.cfg.Theme.Background) {
				return nil
			}
			return terminu.ColorSpans(terminu.Hex(d.cfg.Theme.Background))
		},
		Disabled: func() bool { return d.cfg.Theme.Base == "" },
		// Not a color: set by hand in config.yaml
		Err: func() bool { return d.cfg.Theme.Background != "" && !hexColor.MatchString(d.cfg.Theme.Background) },
		Edit: func() tea.Cmd {
			var old terminu.Color
			if hexColor.MatchString(d.cfg.Theme.Background) {
				old = terminu.Hex(d.cfg.Theme.Background)
			}
			var note string
			return d.app.PickColor(terminu.ColorPicker{Title: "Background", Glyph: glyphPalette, Value: old,
				Commit: func(c terminu.Color) (err error) {
					d.cfg.Theme.Background = ""
					if c.Set {
						d.cfg.Theme.Background = c.String()
					}
					note, err = d.setTheme()
					return err
				},
				Done: func(terminu.Color) tea.Cmd { return d.app.Toast(terminu.Success, note) },
			})
		},
	}
}

// resetBackground is x on the drawer background: back to the theme's own. The
// color picker has no way to choose none (tdp components/input/color-picker),
// so it is an item operation of its own, there only on that field.
func (d *drawer) resetBackground() terminu.Action {
	return terminu.Action{
		Key: "x", Label: "Reset", Desc: "the drawer background to the theme's own", Scope: terminu.ItemScope, Hint: "reset",
		Hidden: func() bool {
			page := d.pages[rowPreferences]
			return d.detail.Body != page || page.Current() != d.background
		},
		Disabled: func() bool { return d.cfg.Theme.Background == "" },
		Do: func() tea.Cmd {
			d.cfg.Theme.Background = ""
			return d.applyTheme()
		},
	}
}

// setTheme puts the theme as picked into Claude Code, ccdrawu's or the user's
// own as it was before it, and saves: what to say when it is done, or why it
// can't be, with everything put back as saved.
func (d *drawer) setTheme() (note string, err error) {
	err = d.commit(func() error {
		t := &d.cfg.Theme
		if t.Base == "" {
			if err := claude.RestoreTheme(d.loc.Claude, t.Before); err != nil {
				return fmt.Errorf("can't put your own theme back: %w", err)
			}
			t.Before = ""
			note = "Claude Code has its own theme back"
			return nil
		}
		before, switched, err := claude.ApplyTheme(d.loc.Claude, t.Base, t.Background)
		if err != nil {
			return err
		}
		if switched {
			t.Before = before
		}
		note = "Claude Code's theme is now " + t.Base
		if t.Background != "" {
			note += " with " + t.Background
		}
		return nil
	})
	return note, err
}

// applyTheme sets the theme where no input waits on it: a Success toast, or
// an error popup.
func (d *drawer) applyTheme() tea.Cmd {
	note, err := d.setTheme()
	if err != nil {
		return d.app.Error("Can't set the theme", err.Error())
	}
	return d.app.Toast(terminu.Success, note)
}

// barField picks how the drawer draws its bars, drawn in place: Nerd Font
// squares, or a line for a terminal with no Nerd Font. No terminal says which
// fonts it has, so each option shows its own bar, and one this terminal can't
// draw shows as boxes.
func (d *drawer) barField() *terminu.Field {
	options := []string{
		"squares  " + strings.Repeat("\uF445", 4) + strings.Repeat("\uF51D", 4),
		"line     " + strings.Repeat("━", 8),
	}
	f := &terminu.Field{Label: "Bar", Options: options}
	f.OnChange = func() tea.Cmd {
		d.cfg.Bar = config.BarStyles[f.Selected]
		return d.save()
	}
	return d.radio(f, func() int { return max(0, slices.Index(config.BarStyles, d.cfg.Bar)) })
}

// petField picks the animal the drawer's pet is, drawn in place; open sessions
// draw the new one within a second.
func (d *drawer) petField() *terminu.Field {
	options := config.PetSpecies
	f := &terminu.Field{Label: "Pet", Options: options}
	f.OnChange = func() tea.Cmd {
		d.cfg.Pet.Species = options[f.Selected]
		return d.save()
	}
	return d.radio(f, func() int { return max(0, slices.Index(options, d.cfg.Pet.Species)) })
}

// The plugin row: whether Claude Code has the ccdrawu plugin, found with
// `claude plugin list` off the screen's loop; Enter installs or uninstalls
// it, after asking, as `ccdrawu install claude` and `uninstall claude` do.
type pluginChecked struct {
	id      string
	enabled bool
	err     error
}

type pluginDone struct {
	msg       string
	uninstall bool
	err       error
}

func (d *drawer) pluginField() *terminu.Field {
	return &terminu.Field{
		Label: "Plugin", Value: func() string { return d.plugin }, Empty: "checking…",
		Disabled: func() bool { return d.plugin == "" || d.busy },
		Err:      func() bool { return d.pluginErr },
		Edit:     d.pluginAction,
	}
}

func (d *drawer) checkPlugin() tea.Cmd {
	run := d.run
	return func() tea.Msg {
		id, enabled, err := claude.Installed(run)
		return pluginChecked{id, enabled, err}
	}
}

func (d *drawer) pluginChecked(m pluginChecked) {
	d.busy, d.pluginID, d.pluginErr = false, m.id, m.err != nil
	switch {
	case m.err != nil:
		d.plugin = "can't tell: " + m.err.Error()
		d.app.Log(terminu.Warning, "claude plugin list did not say whether Claude Code has the plugin: "+m.err.Error())
	case m.id == "":
		d.plugin = "not installed"
	case !m.enabled:
		d.plugin = "installed, turned off"
	default:
		d.plugin = "installed"
	}
}

func (d *drawer) pluginAction() tea.Cmd {
	run, source := d.run, claude.SourceFrom(os.Getenv, version.Version)
	if d.pluginID == "" {
		return d.app.Confirm(terminu.Ask{
			Title: "Install", Subject: "Claude Code", Question: "Add the drawer to Claude Code?", Verb: "install",
			Details: []string{
				"Installs the ccdrawu plugin for your user, as ccdrawu install claude does.",
				"New sessions have the drawer; open ones take it with /reload-plugins.",
			},
		}, func() tea.Cmd {
			d.busy = true
			return func() tea.Msg {
				msg, err := claude.Install(run, source)
				return pluginDone{msg: msg, err: err}
			}
		})
	}
	return d.app.Confirm(terminu.Ask{
		Title: "Uninstall", Subject: "Claude Code", Question: "Take the drawer out of Claude Code?", Verb: "uninstall",
		Details: []string{
			"Removes the ccdrawu plugin and its marketplace, as ccdrawu uninstall claude does.",
			"ccdrawu's theme and status line go too; your picks here stay.",
		},
	}, func() tea.Cmd {
		d.busy = true
		return func() tea.Msg {
			msg, _, err := claude.Uninstall(run)
			return pluginDone{msg: msg, uninstall: true, err: err}
		}
	})
}

// pluginDone says how it went and looks again; an uninstall puts the user's
// own theme and status line back, as the command line's does.
func (d *drawer) pluginDone(m pluginDone) tea.Cmd {
	if m.err != nil {
		d.busy = false
		// What failed, in the title; no name there, as a title is upper case
		// throughout (tdp components, capitals)
		title := "Install failed"
		if m.uninstall {
			title = "Uninstall failed"
		}
		return tea.Batch(d.app.Error(title, m.err.Error()), d.checkPlugin())
	}
	cmds := []tea.Cmd{d.app.Toast(terminu.Info, m.msg), d.checkPlugin()}
	if m.uninstall && d.cfg.Theme.Base != "" {
		d.cfg.Theme.Base = ""
		cmds = append(cmds, d.applyTheme())
	}
	if m.uninstall && d.cfg.Statusline.On {
		d.statusOn.Selected = 0
		cmds = append(cmds, d.statuslineOff())
	}
	return tea.Batch(cmds...)
}
