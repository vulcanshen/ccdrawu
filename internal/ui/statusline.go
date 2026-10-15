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

// statusOnField turns the statusline on or off, drawn in place. Turned on over
// a status line of the user's own, it asks first and keeps that one to put
// back.
func (d *drawer) statusOnField() *terminu.Field {
	f := &terminu.Field{Label: "Statusline", Options: []string{"off", "on"}}
	if d.cfg.Statusline.On {
		f.Selected = 1
	}
	f.OnChange = func() tea.Cmd {
		if f.Selected == 0 {
			return d.statuslineOff()
		}
		command, ours, err := claude.CurrentStatusline(d.loc.Claude)
		if err != nil {
			f.Selected = 0
			return d.app.Error("Can't read Claude Code's settings", err.Error())
		}
		if command == "" || ours {
			return d.statuslineOn()
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
			f.Selected = 1
			return d.statuslineOn()
		})
	}
	return f
}

// statuslineOn sets ccdrawu's status line in Claude Code, keeping the one it
// replaces; the mod starts keeping each session's values for it.
func (d *drawer) statuslineOn() tea.Cmd {
	s := &d.cfg.Statusline
	before, switched, err := claude.ApplyStatusline(d.loc.Claude, refresh(*s))
	if err != nil {
		d.statusOn.Selected = 0
		return d.app.Error("Can't set the status line", err.Error())
	}
	if switched {
		s.Before = config.Kept(before)
	}
	s.On = true
	return tea.Batch(d.save(), d.app.Toast(terminu.Info, "Claude Code's status line is now ccdrawu's"))
}

// statuslineOff puts back the status line ccdrawu's replaced, or none.
func (d *drawer) statuslineOff() tea.Cmd {
	s := &d.cfg.Statusline
	if err := claude.RestoreStatusline(d.loc.Claude, json.RawMessage(s.Before)); err != nil {
		d.statusOn.Selected = 1
		return d.app.Error("Can't put the status line back", err.Error())
	}
	s.On, s.Before = false, nil
	return tea.Batch(d.save(), d.app.Toast(terminu.Info, "Claude Code has its own status line back"))
}

// formatField picks how each item reads, drawn in place, each option showing
// itself.
func (d *drawer) formatField() *terminu.Field {
	options := []string{"label  Context 42%", "short  ctx 42%", "value  42%"}
	f := &terminu.Field{Label: "Format", Options: options, Selected: max(0, slices.Index(config.Formats, d.cfg.Statusline.Format))}
	f.OnChange = func() tea.Cmd {
		d.cfg.Statusline.Format = config.Formats[f.Selected]
		return d.save()
	}
	return f
}

// colorField picks how the statusline is colored, drawn in place: in the
// drawer's colors, or all in Claude Code's dim color.
func (d *drawer) colorField() *terminu.Field {
	f := &terminu.Field{Label: "Color", Options: config.Colors, Selected: max(0, slices.Index(config.Colors, d.cfg.Statusline.Color))}
	f.OnChange = func() tea.Cmd {
		d.cfg.Statusline.Color = config.Colors[f.Selected]
		return d.save()
	}
	return f
}

// separatorField is what stands between the items, | when none is set.
func (d *drawer) separatorField() *terminu.Field {
	return &terminu.Field{Label: "Separator", Value: func() string { return d.cfg.Statusline.Separator }, Empty: "|", Edit: func() tea.Cmd {
		return d.app.Text(terminu.TextInput{
			Title: "Separator", Glyph: glyphPencil, Prompt: "What stands between the items, such as | or ·; empty for |",
			Verb: "save", Old: d.cfg.Statusline.Separator, MaxLen: 3,
			Check: func(v string) error {
				if strings.ContainsFunc(v, unicode.IsControl) {
					return errors.New("No control characters")
				}
				return nil
			},
			Done: func(v string) tea.Cmd {
				d.cfg.Statusline.Separator = strings.TrimSpace(v)
				return d.save()
			},
		})
	}}
}

// refreshField is how many seconds Claude Code waits to run the statusline
// again while nothing happens, written into its settings at once while the
// statusline is on.
func (d *drawer) refreshField() *terminu.Field {
	return &terminu.Field{Label: "Refresh", Value: func() string { return fmt.Sprintf("every %d seconds", refresh(d.cfg.Statusline)) }, Edit: func() tea.Cmd {
		return d.app.Text(terminu.TextInput{
			Title: "Refresh", Glyph: glyphPencil, Prompt: "Seconds between runs while nothing happens, 1 to 60",
			Verb: "save", Old: strconv.Itoa(refresh(d.cfg.Statusline)), MaxLen: 2,
			Check: func(v string) error {
				if n, err := strconv.Atoi(strings.TrimSpace(v)); err != nil || n < 1 || n > 60 {
					return errors.New("A whole number of seconds, 1 to 60")
				}
				return nil
			},
			Done: func(v string) tea.Cmd {
				d.cfg.Statusline.Refresh, _ = strconv.Atoi(strings.TrimSpace(v))
				// Into Claude Code's settings while ccdrawu's is the status line set there
				if _, ours, _ := claude.CurrentStatusline(d.loc.Claude); ours && d.cfg.Statusline.On {
					if _, _, err := claude.ApplyStatusline(d.loc.Claude, d.cfg.Statusline.Refresh); err != nil {
						return d.app.Error("Can't set the status line", err.Error())
					}
				}
				return d.save()
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
