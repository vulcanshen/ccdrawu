package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vulcanshen/ccdrawu/internal/claude"
	"github.com/vulcanshen/ccdrawu/internal/config"
	terminu "github.com/vulcanshen/terminu-go"
)

// The statusline page: whether `ccdrawu statusline` is Claude Code's status
// line (internal/claude, statusline), the items it shows and their order as
// the drawer's pages have them, how each reads, what stands between them, and
// how often Claude Code runs it again while nothing happens.

// statuslineFields is the page, its items' fields as the drawer's.
func (d *drawer) statuslineFields() []*terminu.Field {
	fields := []*terminu.Field{d.statusOn}
	fields = append(fields, d.summary(rowStatusline)...)
	return append(fields, d.formatField(), d.colorField(), d.separatorField(), d.refreshField())
}

// statusOnField turns the statusline off, on, or on when narrow: only while
// the terminal is too narrow for the sidebar. It is drawn in place. Turned on over a status
// line of the user's own, it asks first and keeps that one to put back.
func (d *drawer) statusOnField() *terminu.Field {
	f := &terminu.Field{Label: "Statusline", Options: []string{"off", "on", "on when narrow"}}
	d.radio(f, d.statusSelected)
	f.OnChange = func() tea.Cmd {
		if f.Selected == 0 {
			return d.statuslineOff()
		}
		picked, when := f.Selected, ""
		if picked == 2 {
			when = config.WhenNarrow
		}
		// From on to on when narrow, or back: ccdrawu's is the status line already
		if d.cfg.Statusline.On {
			d.cfg.Statusline.When = when
			return d.save()
		}
		command, ours, err := claude.CurrentStatusline(d.loc.Claude)
		if err != nil {
			f.Selected = 0
			// No name in the title, upper case throughout (tdp components, capitals)
			return d.app.Error("Can't read the settings", err.Error())
		}
		if command == "" || ours {
			return d.statuslineOn(when)
		}
		// On only once the user says so: a no leaves it off
		f.Selected = 0
		return d.app.Confirm(terminu.Ask{
			Title: "Replace", Subject: "status line", Question: "Replace it with ccdrawu's?", Verb: "replace",
			Details: []string{
				"Claude Code runs this as its status line now:",
				command,
				"Turning the statusline off, or ccdrawu uninstall claude, puts it back.",
			},
			Warn: []int{1},
		}, func() tea.Cmd {
			f.Selected = picked
			return d.statuslineOn(when)
		})
	}
	return f
}

// statusSelected is the statusline's option as the config has it: off, on, or
// on when narrow.
func (d *drawer) statusSelected() int {
	switch s := d.cfg.Statusline; {
	case !s.On:
		return 0
	case s.When == config.WhenNarrow:
		return 2
	}
	return 1
}

// statuslineOn sets ccdrawu's status line in Claude Code, keeping the one it
// replaces; the mod starts keeping each session's values for it. when is when
// it shows (config.Statusline).
func (d *drawer) statuslineOn(when string) tea.Cmd {
	s := &d.cfg.Statusline
	before, switched, err := claude.ApplyStatusline(d.loc.Claude, refresh(*s))
	if err != nil {
		d.statusOn.Selected = 0
		return d.app.Error("Can't set the status line", err.Error())
	}
	if switched {
		s.Before = config.Kept(before)
	}
	s.On, s.When = true, when
	return tea.Batch(d.save(), d.app.Toast(terminu.Success, "Claude Code's status line is now ccdrawu's"))
}

// statuslineOff puts back the status line ccdrawu's replaced, or none.
func (d *drawer) statuslineOff() tea.Cmd {
	s := &d.cfg.Statusline
	if err := claude.RestoreStatusline(d.loc.Claude, json.RawMessage(s.Before)); err != nil {
		d.statusOn.Selected = d.statusSelected()
		return d.app.Error("Can't put the status line back", err.Error())
	}
	s.On, s.When, s.Before = false, "", nil
	return tea.Batch(d.save(), d.app.Toast(terminu.Success, "Claude Code has its own status line back"))
}

// formatField picks how each item reads, drawn in place, each option showing
// itself.
func (d *drawer) formatField() *terminu.Field {
	options := []string{"label  Context 42%", "short  ctx 42%", "value  42%"}
	f := &terminu.Field{Label: "Format", Options: options}
	f.OnChange = func() tea.Cmd {
		d.cfg.Statusline.Format = config.Formats[f.Selected]
		return d.save()
	}
	return d.radio(f, func() int { return max(0, slices.Index(config.Formats, d.cfg.Statusline.Format)) })
}

// colorField picks how the statusline is colored, drawn in place: in the
// drawer's colors, or all in Claude Code's dim color.
func (d *drawer) colorField() *terminu.Field {
	f := &terminu.Field{Label: "Color", Options: config.Colors}
	f.OnChange = func() tea.Cmd {
		d.cfg.Statusline.Color = config.Colors[f.Selected]
		return d.save()
	}
	return d.radio(f, func() int { return max(0, slices.Index(config.Colors, d.cfg.Statusline.Color)) })
}

// separatorField is what stands between the items, | when none is set.
func (d *drawer) separatorField() *terminu.Field {
	return &terminu.Field{Label: "Separator", Value: func() string { return d.cfg.Statusline.Separator }, Empty: "|",
		// Control characters set by hand in config.yaml
		Err: func() bool { return strings.ContainsFunc(d.cfg.Statusline.Separator, unicode.IsControl) },
		Edit: func() tea.Cmd {
			return d.app.Text(terminu.TextInput{
				Title: "Separator", Glyph: glyphPencil, Prompt: "What stands between the items, such as | or ·; empty for |",
				Verb: "save", Old: d.cfg.Statusline.Separator, MaxLen: 3,
				Check: func(v string) error {
					if strings.ContainsFunc(v, unicode.IsControl) {
						return errors.New("No control characters")
					}
					return nil
				},
				Commit: func(v string) error {
					return d.commit(func() error {
						d.cfg.Statusline.Separator = strings.TrimSpace(v)
						return nil
					})
				},
			})
		}}
}

// refreshField is how many seconds Claude Code waits to run the statusline
// again while nothing happens, a number (tdp components/input/number) written
// into its settings at once while the statusline is on.
func (d *drawer) refreshField() *terminu.Field {
	// The value and its unit, a whole word (tdp components/input/README)
	return &terminu.Field{Label: "Refresh", Value: func() string { return strconv.Itoa(refresh(d.cfg.Statusline)) + " seconds" },
		// Past 60, set by hand in config.yaml
		Err: func() bool { return refresh(d.cfg.Statusline) > 60 },
		Edit: func() tea.Cmd {
			// The unit on a row of its own under the value (tdp components/input/number)
			return d.app.Number(terminu.NumberInput{
				Title: "Refresh", Glyph: glyphPencil, Prompt: "Seconds between runs while nothing happens, 1-60",
				Verb: "save", Old: strconv.Itoa(refresh(d.cfg.Statusline)), Min: 1, Max: 60, Unit: "seconds",
				Commit: func(v string) error {
					return d.commit(func() error {
						d.cfg.Statusline.Refresh, _ = strconv.Atoi(strings.TrimSpace(v))
						// Into Claude Code's settings while ccdrawu's is the status line set there
						if _, ours, _ := claude.CurrentStatusline(d.loc.Claude); ours && d.cfg.Statusline.On {
							if _, _, err := claude.ApplyStatusline(d.loc.Claude, d.cfg.Statusline.Refresh); err != nil {
								return fmt.Errorf("can't set the status line: %w", err)
							}
						}
						return nil
					})
				},
			})
		}}
}

// refresh is the statusline's refresh in seconds, the default when none is
// set.
func refresh(s config.Statusline) int {
	if s.Refresh > 0 {
		return s.Refresh
	}
	return config.DefaultRefresh
}
