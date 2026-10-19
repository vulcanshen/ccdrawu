package ui

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	terminu "github.com/vulcanshen/terminu-go"
)

// run plays out cmd as press plays out a key: what it sends is sent to the
// app, and a command still waiting after 100ms, a timer, is dropped.
func run(d *drawer, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	got := make(chan tea.Msg, 1)
	go func() { got <- cmd() }()
	select {
	case m := <-got:
		if batch, ok := m.(tea.BatchMsg); ok {
			for _, c := range batch {
				run(d, c)
			}
		} else if m != nil {
			press(d, m)
		}
	case <-time.After(100 * time.Millisecond):
	}
}

// tdp L4, tdp E5 and tdp E6: at every size, with a Nerd Font icon and a powerline
// symbol one or two cells wide, every row is exactly the terminal's width and
// there are as many rows as it is tall, with each popup ccdrawu opens on top;
// opened, then the terminal made narrower, lower, or both, than the popup.
func TestL4EveryRowFullWidth(t *testing.T) {
	t.Cleanup(func() { terminu.SetIconWidth(1, 1) })
	enter, j := tea.KeyMsg{Type: tea.KeyEnter}, key("j")
	toPrefs := []tea.Msg{j, j, j, enter}
	opens := map[string][]tea.Msg{
		"none":   nil,
		"space":  {space},
		"global": {space, key("G"), enter},
		"keys":   {key("?")},
		"items":  {enter, j, j, enter},
		"drag":   {enter, j, j, j, enter, key("D")},
		"alias":  append(toPrefs, enter),
		"theme":  append(toPrefs, j, enter),
		// The theme picked first, as the drawer background is off without one
		"color":   append(toPrefs, j, enter, j, enter, j, enter),
		"channel": append(toPrefs, j, enter, j, enter, j, enter, enter),
		"refresh": {j, j, enter, key("G"), enter},
		"confirm": append(toPrefs, key("G"), enter),
	}
	others := map[string]func(d *drawer) tea.Cmd{
		"error": func(d *drawer) tea.Cmd {
			return d.app.Error("Can't save", "a reason long enough to wrap on a narrow terminal, and then some more words")
		},
		"toast": func(d *drawer) tea.Cmd {
			return d.app.Toast(terminu.Info, "a toast long enough to wrap on a narrow terminal, and then some more words")
		},
	}
	// Each popup opens wide and at the minimum size, 40x12 (tdp L1); the one
	// opened wide is then drawn at each size after, the last three smaller
	// than it one way or both
	opened := [][2]int{{200, 50}, {40, 12}}
	sizes := [][2]int{{120, 40}, {100, 30}, {80, 24}, {60, 20}, {30, 12}, {24, 40}, {120, 7}, {24, 7}}
	check := func(d *drawer, label string, w, h int) {
		t.Helper()
		lines := d.app.Frame()
		if len(lines) != h {
			t.Errorf("%s at %dx%d: %d rows", label, w, h, len(lines))
		}
		for i, l := range lines {
			if got := terminu.Width(l); got != w {
				t.Errorf("%s at %dx%d: row %d is %d cells wide", label, w, h, i, got)
				return
			}
		}
	}
	for _, cells := range [][2]int{{1, 1}, {2, 1}, {2, 2}} {
		terminu.SetIconWidth(cells[0], cells[1])
		for name := range opens {
			others[name] = nil
		}
		for name, open := range others {
			label := fmt.Sprintf("icon %d, powerline %d, %s", cells[0], cells[1], name)
			for i, at := range opened {
				d, _ := newTestDrawer(t)
				d.pluginChecked(pluginChecked{})
				press(d, tea.WindowSizeMsg{Width: at[0], Height: at[1]})
				for _, k := range opens[name] {
					press(d, k)
				}
				if open != nil {
					run(d, open(d))
				}
				check(d, label, at[0], at[1])
				if i > 0 {
					continue
				}
				for _, sz := range sizes {
					d.app.Update(tea.WindowSizeMsg{Width: sz[0], Height: sz[1]})
					check(d, label+" opened at 200x50", sz[0], sz[1])
				}
			}
		}
	}
}
