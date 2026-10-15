package ui

import (
	"fmt"
	"os"
	"path/filepath"
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
	fail   *terminu.Failure           // what an empty page of items says

	theme      *terminu.Field // the theme's radio, set back to off by an uninstall
	background *terminu.Field // the drawer background, which x resets
	run        claude.Run     // what runs `claude plugin ...`
	plugin     string         // the plugin row's value; "" while it is being found
	pluginID   string         // the plugin's id in Claude Code, "" when not there
	busy       bool           // an install or uninstall under way
}

// The rows of [1]: the drawer beside the transcript and above the prompt, then
// the preferences, every setting on one page (the alias here, the rest in
// settings.go)
const (
	rowSidebar     = "sidebar"
	rowInline      = "inline"
	rowPreferences = "preferences"
)

var placements = map[string]config.Placement{rowSidebar: config.Dock, rowInline: config.Inline}

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
	cfg, err := config.ReadConfig(loc.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", loc.ConfigFile, err)
	}
	d := &drawer{
		loc: loc,
		cfg: cfg,
		run: run,
		pages: map[string]*terminu.Values{
			rowSidebar: {},
			rowInline:  {},
		},
		fail: &terminu.Failure{
			Fact: "Nothing to list until ccdrawu runs in Claude Code",
			Hint: "Run ccdrawu install claude, then start Claude Code",
		},
	}
	a := terminu.New("ccdrawu")
	a.Splash = splash()
	a.Status = d.status
	a.OnMsg = d.onMsg
	a.Start = func() tea.Cmd { return tea.Batch(d.reload(), d.checkPlugin()) }
	d.app = a
	d.theme, d.background = d.themeField(), d.backgroundField()
	d.pages[rowPreferences] = &terminu.Values{Fields: []*terminu.Field{
		d.aliasField(), d.theme, d.background, d.barField(), d.petField(), d.pluginField(),
	}}

	list := &terminu.List{}
	list.SetRows([]terminu.Row{
		{ID: rowSidebar, Cells: []string{rowSidebar}, Group: "Drawer"},
		{ID: rowInline, Cells: []string{rowInline}, Group: "Drawer"},
		{ID: rowPreferences, Cells: []string{rowPreferences}, Group: "Settings"},
	})
	list.OnMove = func(r terminu.Row) tea.Cmd {
		d.show(r.ID)
		return nil
	}
	d.detail = &terminu.Panel{Actions: []terminu.Action{d.editItems(), d.resetBackground()}}
	d.show(rowSidebar)
	// Enter goes into the row's page in [2]: the one change of panels Enter
	// makes (tdp v0.6.3 K3, details). Every row has a page.
	prefs := &terminu.Panel{Label: "Preferences", Body: list, Details: d.detail, Actions: []terminu.Action{
		{Key: "Enter", Label: "Edit", Desc: "go into it", Scope: terminu.ItemScope, Hint: "edit",
			ToDetails: func() bool { return true }},
	}}
	a.AddScreen(&terminu.Screen{
		Name: "Preferences", Key: "P", Desc: "what the drawer shows, and its command",
		Panels:      []*terminu.Panel{prefs, d.detail},
		Layout:      sidebar,
		NarrowBelow: 60,
	})
	d.load()
	return d, nil
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
	d.detail.Fail = d.fail
}

// load reads the catalog when it is new or changed: Claude Code writes it each
// time the mod starts, so a mod updated (or run for the first time) while the
// TUI is open shows up by itself.
func (d *drawer) load() {
	path := filepath.Join(d.loc.Data, "catalog.json")
	info, err := os.Stat(path)
	if err != nil || info.ModTime().Equal(d.stamp) {
		return
	}
	cat, ok, err := config.ReadCatalog(path)
	if err != nil {
		d.fail = &terminu.Failure{Fact: "The catalog can't be read: " + err.Error(), Hint: "Start Claude Code again to write it anew"}
		d.detail.Fail = d.fail
		return
	}
	if !ok {
		return
	}
	d.cat, d.stamp = cat, info.ModTime()
	for id := range placements {
		d.pages[id].Fields = d.summary(id)
	}
}

func (d *drawer) reload() tea.Cmd {
	d.load()
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg { return catalogTick{} })
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
// editItems) opens; and Order, the same in the order the drawer draws them.
// The order is set in a reorder popup once terminu-go has one (CD9, tdp
// v0.6.6 components/input/reorder); until then the row only shows it.
func (d *drawer) summary(id string) []*terminu.Field {
	p := placements[id]
	return []*terminu.Field{
		{Label: "Items", Empty: "none", Value: func() string { return d.chosen(p) }, Edit: func() tea.Cmd { return d.pick(id) }},
		{Label: "Order", Empty: "none", Value: func() string { return d.order(p) }, Disabled: func() bool { return true }},
	}
}

// chosen is the items on in p, in the catalog's order, as the popup lists them.
func (d *drawer) chosen(p config.Placement) string {
	var keys []string
	for _, it := range d.cat.Items {
		if d.cfg.On(p, it) {
			keys = append(keys, it.Label)
		}
	}
	return strings.Join(keys, ", ")
}

// order leaves out the pet: the drawer draws it on its own, over the rows or
// left of the columns, wherever it is in the list.
func (d *drawer) order(p config.Placement) string {
	var keys []string
	for _, it := range d.cfg.Shown(p, d.cat) {
		if it.ID != petID {
			keys = append(keys, it.Label)
		}
	}
	return strings.Join(keys, ", ")
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
	gw, lw := 0, 0
	for _, it := range cat.Items {
		gw, lw = max(gw, terminu.Width(it.Group)), max(lw, terminu.Width(it.Label))
	}
	options := make([]string, len(cat.Items))
	checked := make([]bool, len(cat.Items))
	for i, it := range cat.Items {
		row := terminu.Pad(it.Group, gw) + "  " + terminu.Pad(it.Label, lw)
		if it.Name != it.Label {
			row += "  " + it.Name
		}
		options[i] = strings.TrimRight(row, " ")
		checked[i] = d.cfg.On(p, it)
	}
	return d.app.Checks(terminu.Checks{
		Title: "Edit", Glyph: glyphPencil, Subject: id, Options: options, Checked: checked,
		Done: func(on []bool) tea.Cmd {
			for i, it := range cat.Items {
				if on[i] != d.cfg.On(p, it) {
					d.cfg.Set(p, cat, it.ID, on[i])
				}
			}
			return d.save()
		},
	})
}

// aliasField is the alias row: one more name for /ccdrawu, set in an input
// popup and checked as the mod reads it.
func (d *drawer) aliasField() *terminu.Field {
	return &terminu.Field{Label: "Alias", Value: func() string { return d.cfg.Alias }, Empty: "none", Edit: func() tea.Cmd {
		return d.app.Text(terminu.TextInput{
			Title: "Alias", Glyph: glyphPencil, Prompt: "One more name for /ccdrawu, such as ccd; empty for none",
			Verb: "save", Old: d.cfg.Alias, MaxLen: 64,
			Check: func(v string) error {
				_, err := config.CleanAlias(v)
				return err
			},
			Done: func(v string) tea.Cmd {
				d.cfg.Alias, _ = config.CleanAlias(v)
				return d.save()
			},
		})
	}}
}

func (d *drawer) save() tea.Cmd {
	if err := config.WriteConfig(d.loc.ConfigFile, d.cfg); err != nil {
		return d.app.Error("Can't save", err.Error())
	}
	return nil
}

// status says where the choice goes and that open sessions take it.
func (d *drawer) status(w int) string {
	grey := terminu.Fg(terminu.Overlay0)
	text := grey.Render("Saved to ") + terminu.Fg(terminu.Lavender).Render(shortHome(d.loc.ConfigFile)) +
		grey.Render(" · open Claude Code sessions follow within a second")
	return terminu.Truncate(text, w)
}

func shortHome(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && (path == home || strings.HasPrefix(path, home+string(filepath.Separator))) {
		return "~" + path[len(home):]
	}
	return path
}
