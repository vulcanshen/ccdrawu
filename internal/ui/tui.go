package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vulcanshen/ccdrawu/internal/claude"
	"github.com/vulcanshen/ccdrawu/internal/config"
	terminu "github.com/vulcanshen/terminu-go"
)

// drawer is the TUI (tdp components/layout/panel, groups and details): [1]
// Preferences lists what can be set, in groups; [2] is the one under its
// cursor and follows it at once. A change is written at once, and every open
// Claude Code session takes it within a second.
type drawer struct {
	app    *terminu.App
	loc    config.Locations
	cat    config.Catalog
	stamp  time.Time // the catalog's mtime when it was read
	cfg    config.Config
	detail *terminu.Panel
	pages  map[string]*terminu.Values // [2]'s body for each row of [1]

	background *terminu.Field            // the drawer background, which x resets
	drawerOn   map[string]*terminu.Field // the drawer's on and off atop sidebar and inline, each its own
	statusOn   *terminu.Field            // the statusline's on and off, set back to off by an uninstall
	run        claude.Run                // what runs `claude plugin ...`
	plugin     string                    // the plugin row's value; "" while it is being found
	pluginID   string                    // the plugin's id in Claude Code, "" when not there
	pluginErr  bool                      // the plugin row's value says why it can't tell
	busy       bool                      // an install or uninstall under way
	// radios are the fields drawn in place, each with its option as d.cfg has
	// it: a change that could not be saved puts them back with d.cfg
	radios map[*terminu.Field]func() int
}

// The rows of [1]: the drawer beside the transcript and above the prompt, the
// statusline (statusline.go), then the preferences, every setting on one page
// (the alias here, the rest in settings.go)
const (
	rowSidebar     = "sidebar"
	rowInline      = "inline"
	rowStatusline  = "statusline"
	rowPreferences = "preferences"
)

var placements = map[string]config.Placement{rowSidebar: config.Dock, rowInline: config.Inline, rowStatusline: config.Status}

const glyphPencil = "\U000F03EB" // nf-md-pencil, the title of the alias input and of the items popup

const petID = "pet" // the pet's id in the catalog (hooks/items.ts)

type catalogTick struct{}

// Run opens the TUI on the files in loc.
func Run(loc config.Locations) error {
	d, err := newDrawer(loc, claude.Exec)
	if err != nil {
		return err
	}
	return d.app.Run()
}

func newDrawer(loc config.Locations, run claude.Run) (*drawer, error) {
	// What 0.1.4 and before kept elsewhere, taken where tdp E3 has it now,
	// said once the screen is up
	moved, moveErr := config.Move(loc)
	// A config that can't be read is not worked on: the screen opens on none
	// and asks to move it aside first (tdp E3, can't be read)
	source := config.Source(loc.ConfigFile)
	cfg, readErr := config.ReadConfig(loc.ConfigFile)
	if readErr != nil {
		cfg = config.Config{}
	}
	// A status line set in Claude Code since, in place of ccdrawu's, turns the
	// statusline off: the mod need not keep values for it
	statusTaken := false
	if _, ours, err := claude.CurrentStatusline(loc.Claude); readErr == nil && err == nil && cfg.Statusline.On && !ours {
		cfg.Statusline.On, cfg.Statusline.Before = false, nil
		if err := config.WriteConfig(loc.ConfigFile, cfg); err != nil {
			return nil, err
		}
		statusTaken = true
	}
	// A page of items before the mod has written the catalog: an empty state
	// (tdp components/layout/screen), not a failure
	page := func() *terminu.Values {
		return &terminu.Values{Header: [2]string{"Setting", "Value"},
			EmptyFact: "Nothing to list until ccdrawu runs in Claude Code", EmptyHint: "Press [Enter] on Plugin in preferences"}
	}
	d := &drawer{
		loc:    loc,
		cfg:    cfg,
		run:    run,
		pages:  map[string]*terminu.Values{rowSidebar: page(), rowInline: page(), rowStatusline: page()},
		radios: map[*terminu.Field]func() int{},
	}
	a := terminu.New("ccdrawu")
	// It opens in a narrow pane beside Claude Code, and one panel and a page of
	// values fit in 40 x 12 (tdp L1; the user's)
	a.MinWidth, a.MinHeight = 40, 12
	a.InputTabWidth = cfg.InputTabWidth
	a.Splash = splash()
	a.Status = d.status
	a.OnMsg = d.onMsg
	a.Start = func() tea.Cmd {
		if statusTaken {
			a.Log(terminu.Info, "Claude Code has a status line of its own now: ccdrawu's statusline is off")
		}
		cmds := []tea.Cmd{d.reload(), d.checkPlugin(), d.sayMoved(moved, moveErr)}
		if readErr != nil {
			cmds = append(cmds, a.MoveAside(source, readErr, func() tea.Cmd {
				d.revert() // none there now: the defaults
				return nil
			}))
		}
		return tea.Batch(cmds...)
	}
	d.app = a
	d.background, d.statusOn = d.backgroundField(), d.statusOnField()
	d.drawerOn = map[string]*terminu.Field{rowSidebar: d.drawerOnField(&d.cfg.Drawer.SidebarOff), rowInline: d.drawerOnField(&d.cfg.Drawer.InlineOff)}
	d.pages[rowPreferences] = &terminu.Values{Header: [2]string{"Key", "Value"}, Fields: []*terminu.Field{
		d.aliasField(), d.themeField(), d.background, d.barField(), d.petField(), d.pluginField(),
	}}

	list := &terminu.List{}
	list.SetRows([]terminu.Row{
		{ID: rowSidebar, Cells: []string{rowSidebar}, Group: "Drawer"},
		{ID: rowInline, Cells: []string{rowInline}, Group: "Drawer"},
		{ID: rowStatusline, Cells: []string{rowStatusline}, Group: "Statusline"},
		{ID: rowPreferences, Cells: []string{rowPreferences}, Group: "Settings"},
	})
	list.OnMove = func(r terminu.Row) tea.Cmd {
		d.show(r.ID)
		return nil
	}
	d.detail = &terminu.Panel{Actions: []terminu.Action{d.editItems(), d.resetBackground()}}
	d.show(rowSidebar)
	// Enter goes into the row's page in [2]: the one change of panels Enter
	// makes (tdp K3, details, since tdp v0.6.3). Every row has a page.
	prefs := &terminu.Panel{Label: "Preferences", Body: list, Details: d.detail, Actions: []terminu.Action{
		{Key: "Enter", Label: "Edit", Desc: "go into it", Scope: terminu.ItemScope, Hint: "edit",
			ToDetails: func() bool { return true }},
	}}
	a.AddScreen(&terminu.Screen{
		Name: "Preferences", Key: "P", Desc: "what the drawer and the statusline show, and the command",
		Panels:      []*terminu.Panel{prefs, d.detail},
		Layout:      sidebar,
		NarrowBelow: 60,
	})
	d.load()
	return d, nil
}

// sayMoved says what moved where, in an Info toast; what could not move, with
// both places, in an error popup, to be moved by hand (tdp E3, a place changed).
func (d *drawer) sayMoved(moved []config.Moved, err error) tea.Cmd {
	if err != nil {
		return d.app.Error("Can't move", err.Error()+"\nccdrawu reads the old place until it is moved; move it there yourself.")
	}
	var notes []string
	for _, m := range moved {
		notes = append(notes, "Moved "+shortHome(m.From)+" to "+shortHome(m.To))
	}
	if len(notes) == 0 {
		return nil
	}
	return d.app.Toast(terminu.Info, strings.Join(notes, "; "))
}

// sidebar puts [1] in a sidebar 24 columns wide, as locku and webu do, and
// [2] beside it.
func sidebar(w, h int) []terminu.Rect {
	left := min(24, w/2)
	return []terminu.Rect{{X: 0, Y: 0, W: left, H: h}, {X: left, Y: 0, W: w - left, H: h}}
}

// show puts the page of row id in [2], named after it.
func (d *drawer) show(id string) {
	d.detail.Label = id
	d.detail.Body = d.pages[id]
}

// load reads the catalog when it is new or changed: Claude Code writes it each
// time the mod starts, so a mod updated (or run for the first time) while the
// TUI is open shows up by itself.
func (d *drawer) load() {
	path := d.loc.Catalog()
	info, err := os.Stat(path)
	if err != nil || info.ModTime().Equal(d.stamp) {
		return
	}
	cat, ok, err := config.ReadCatalog(path)
	// A cache: one that can't be read is thrown away, and the plugin writes it
	// again as a session starts (tdp E3, can't be read)
	if err != nil {
		os.Remove(path)
		d.app.Log(terminu.Warning, shortHome(path)+" can't be read, thrown away for the plugin to write anew: "+err.Error())
		return
	}
	if !ok {
		return
	}
	d.cat, d.stamp = cat, info.ModTime()
	for _, id := range []string{rowSidebar, rowInline} {
		d.pages[id].Fields = append([]*terminu.Field{d.drawerOn[id]}, d.summary(id)...)
	}
	d.pages[rowStatusline].Fields = d.statuslineFields()
}

// drawerOnField turns the drawer on or off in one place, beside the transcript
// or above the prompt, as off says (config.Drawer), drawn in place. Off there,
// open sessions take it out of that place within a second; off in both,
// sessions start without it. /ccdrawu still opens and closes it. A radio of
// off and on, as the statusline's off, on and on when narrow (the user's; tdp
// components/input/radio has a switch either way).
func (d *drawer) drawerOnField(off *bool) *terminu.Field {
	f := &terminu.Field{Label: "Drawer", Options: []string{"off", "on"}}
	f.OnChange = func() tea.Cmd {
		*off = f.Selected == 0
		return d.save()
	}
	return d.radio(f, func() int {
		if *off {
			return 0
		}
		return 1
	})
}

// radio sets f on its option as d.cfg has it, and keeps it so for when a
// change can't be saved.
func (d *drawer) radio(f *terminu.Field, selected func() int) *terminu.Field {
	d.radios[f] = selected
	f.Selected = selected()
	return f
}

func (d *drawer) reload() tea.Cmd {
	d.load()
	// The app's clock, which a test moves on (tdptest)
	return d.app.Tick(2*time.Second, func(time.Time) tea.Msg { return catalogTick{} })
}

func (d *drawer) onMsg(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case catalogTick:
		return d.reload()
	case pluginChecked:
		d.pluginChecked(m)
	case pluginDone:
		return d.pluginDone(m)
	}
	return nil
}

// summary is the page id in [2], its values shown as tdp components/input
// shows a checkbox popup's and a reorder's, joined with ", ": Items, the
// drawer's keys of the items on there, all set in one popup that Enter (or e,
// editItems) opens; and Order, the same in the order the drawer draws them,
// set in a reorder popup (tdp v0.6.6 components/input/reorder) once there are
// two to order.
func (d *drawer) summary(id string) []*terminu.Field {
	p := placements[id]
	return []*terminu.Field{
		{Label: "Items", Empty: "none", Value: func() string { return d.chosen(p) }, Edit: func() tea.Cmd { return d.pick(id) }},
		{Label: "Order", Empty: "none", Value: func() string { return d.order(p) }, Edit: func() tea.Cmd { return d.reorder(id) },
			Disabled: func() bool { return len(d.ordered(p)) < 2 }},
	}
}

// chosen is the items on in p, in the catalog's order, as the popup lists them.
func (d *drawer) chosen(p config.Placement) string {
	var keys []string
	for _, it := range d.choices(p) {
		if d.cfg.On(p, it) {
			keys = append(keys, it.Label)
		}
	}
	return strings.Join(keys, ", ")
}

func (d *drawer) order(p config.Placement) string {
	var keys []string
	for _, it := range d.ordered(p) {
		keys = append(keys, it.Label)
	}
	return strings.Join(keys, ", ")
}

// ordered is the items on in p in the drawer's order, but the pet: the drawer
// draws it on its own, over the rows or left of the columns, wherever it is in
// the list.
func (d *drawer) ordered(p config.Placement) []config.Item {
	return slices.DeleteFunc(d.cfg.Shown(p, d.cat), func(it config.Item) bool { return it.ID == petID })
}

// reorder opens the reorder popup for the page id: the items on there in the
// drawer's order, each as the checkbox popup writes it. D picks one up, j|k
// move it, Enter drops it, and Enter with nothing held writes the order back.
func (d *drawer) reorder(id string) tea.Cmd {
	p, cat := placements[id], d.cat
	items := d.ordered(p)
	return d.app.Reorder(terminu.Reorder{
		Title: "Order", Glyph: glyphPencil, Subject: id, Options: d.rows(items),
		Commit: func(order []int) error {
			return d.commit(func() error {
				ids := make([]string, len(order))
				for k, i := range order {
					ids[k] = items[i].ID
				}
				d.cfg.Reorder(p, cat, ids)
				return nil
			})
		},
	})
}

// editItems is [2]'s panel operation on the sidebar and inline pages: one
// checkbox popup of every item, more than five so a popup (tdp
// components/input/checkbox), where Space checks and Enter applies them all.
// Enter on Items opens it too and is the one in the hint.
func (d *drawer) editItems() terminu.Action {
	return terminu.Action{
		Key: "e", Label: "Edit", Desc: "pick the items it shows", Scope: terminu.PanelScope,
		Hidden: func() bool {
			_, ok := placements[d.detail.Label]
			return !ok
		},
		Disabled: func() bool { return len(d.cat.Items) == 0 },
		Do:       func() tea.Cmd { return d.pick(d.detail.Label) },
	}
}

// pick opens the checkbox popup for the page id: the items in the catalog's
// order, each as its group, its key in the drawer and, where that says more,
// its name.
func (d *drawer) pick(id string) tea.Cmd {
	p, cat := placements[id], d.cat
	items := d.choices(p)
	checked := make([]bool, len(items))
	for i, it := range items {
		checked[i] = d.cfg.On(p, it)
	}
	return d.app.Checks(terminu.Checks{
		Title: "Edit", Glyph: glyphPencil, Subject: id, Options: d.rows(items), Checked: checked,
		Commit: func(on []bool) error {
			return d.commit(func() error {
				for i, it := range items {
					if on[i] != d.cfg.On(p, it) {
						d.cfg.Set(p, cat, it.ID, on[i])
					}
				}
				return nil
			})
		},
	})
}

// choices is what the checkbox popup offers for p: every item, but the pet in
// the statusline, a row of text it has no room in.
func (d *drawer) choices(p config.Placement) []config.Item {
	if p != config.Status {
		return d.cat.Items
	}
	return slices.DeleteFunc(slices.Clone(d.cat.Items), func(it config.Item) bool { return it.ID == petID })
}

// aliasField is the alias row: one more name for /ccdrawu, set in an input
// popup and checked as the mod reads it.
func (d *drawer) aliasField() *terminu.Field {
	return &terminu.Field{Label: "Alias", Value: func() string { return d.cfg.Alias }, Empty: "none",
		// One set by hand in config.yaml that the mod does not take (tdp
		// components/layout/panel: its label is Red)
		Err: func() bool {
			_, err := config.CleanAlias(d.cfg.Alias)
			return err != nil
		},
		Edit: func() tea.Cmd {
			return d.app.Text(terminu.TextInput{
				Title: "Alias", Glyph: glyphPencil, Prompt: "One more name for /ccdrawu, such as ccd; empty for none",
				Verb: "save", Old: d.cfg.Alias, MaxLen: 64,
				Check: func(v string) error {
					_, err := config.CleanAlias(v)
					return err
				},
				Commit: func(v string) error {
					return d.commit(func() error {
						d.cfg.Alias, _ = config.CleanAlias(v)
						return nil
					})
				},
			})
		}}
}

// rows writes each item as its group, its key in the drawer and, where that
// says more, its name, in columns.
func (d *drawer) rows(items []config.Item) []string {
	gw, lw := 0, 0
	for _, it := range items {
		gw, lw = max(gw, terminu.Width(it.Group)), max(lw, terminu.Width(it.Label))
	}
	rows := make([]string, len(items))
	for i, it := range items {
		row := terminu.Pad(it.Group, gw) + "  " + terminu.Pad(it.Label, lw)
		if it.Name != it.Label {
			row += "  " + it.Name
		}
		rows[i] = strings.TrimRight(row, " ")
	}
	return rows
}

// save writes the config, or says why it can't: a change made in place.
func (d *drawer) save() tea.Cmd {
	if err := d.commit(func() error { return nil }); err != nil {
		return d.app.Error("Can't save", err.Error())
	}
	return nil
}

// commit makes a change and writes the config, or puts it all back as saved
// and says why: an input's Commit, which keeps the input open with its value
// under the error (tdp components/layout/popup).
func (d *drawer) commit(change func() error) error {
	err := change()
	if err == nil {
		err = config.WriteConfig(d.loc.ConfigFile, d.cfg)
	}
	if err != nil {
		d.revert()
		return err
	}
	d.app.Log(terminu.Info, "Saved "+shortHome(d.loc.ConfigFile))
	return nil
}

// revert puts d.cfg, and the radios with it, back to the file as it is: a
// change that could not be made is not shown as if it were.
func (d *drawer) revert() {
	if c, err := config.ReadConfig(d.loc.ConfigFile); err == nil {
		d.cfg = c
		for f, selected := range d.radios {
			f.Selected = selected()
		}
	}
}

// status says where the choice goes and that open sessions take it: the
// drawer within a second, the statusline when Claude Code runs it next. It is
// one sentence, not two items side by side (tdp components, spacing).
func (d *drawer) status(int) []terminu.Span {
	grey := terminu.Fg(terminu.Overlay0)
	follow := "; open Claude Code sessions follow within a second"
	if d.detail.Label == rowStatusline {
		follow = fmt.Sprintf("; the statusline follows within %d seconds", refresh(d.cfg.Statusline))
	}
	// The lib cuts it at the width, and draws a control character in the path by name
	return []terminu.Span{{Text: "Saved to ", Style: grey}, {Text: shortHome(d.loc.ConfigFile), Style: terminu.Fg(terminu.Lavender)}, {Text: follow, Style: grey}}
}

func shortHome(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && (path == home || strings.HasPrefix(path, home+string(filepath.Separator))) {
		return "~" + path[len(home):]
	}
	return path
}
