package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
}

// The rows of [1]: the drawer beside the transcript and above the prompt,
// then the command's alias
const (
	rowSidebar = "sidebar"
	rowInline  = "inline"
	rowAlias   = "alias"
)

var placements = map[string]config.Placement{rowSidebar: config.Dock, rowInline: config.Inline}

const glyphPencil = "\U000F03EB" // nf-md-pencil, the alias input's title

type catalogTick struct{}

// Run opens the TUI on the files in loc.
func Run(loc config.Locations) error {
	d, err := newDrawer(loc)
	if err != nil {
		return err
	}
	return d.app.Run()
}

func newDrawer(loc config.Locations) (*drawer, error) {
	cfg, err := config.ReadConfig(loc.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", loc.ConfigFile, err)
	}
	d := &drawer{
		loc: loc,
		cfg: cfg,
		pages: map[string]*terminu.Values{
			rowSidebar: {},
			rowInline:  {},
		},
		fail: &terminu.Failure{
			Fact: "Nothing to list until ccdrawu runs in Claude Code",
			Hint: "Run ccdrawu install, then start Claude Code",
		},
	}
	a := terminu.New("ccdrawu")
	a.Splash = splash()
	a.Status = d.status
	a.OnMsg = d.onMsg
	a.Start = func() tea.Cmd { return d.reload() }
	d.app = a
	d.pages[rowAlias] = &terminu.Values{Fields: []*terminu.Field{d.aliasField()}}

	list := &terminu.List{}
	list.SetRows([]terminu.Row{
		{ID: rowSidebar, Cells: []string{rowSidebar}, Group: "Drawer"},
		{ID: rowInline, Cells: []string{rowInline}, Group: "Drawer"},
		{ID: rowAlias, Cells: []string{rowAlias}, Group: "Settings"},
	})
	list.OnMove = func(r terminu.Row) tea.Cmd {
		d.show(r.ID)
		return nil
	}
	d.detail = &terminu.Panel{}
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
	for id, p := range placements {
		d.pages[id].Fields = d.fields(p)
	}
}

func (d *drawer) reload() tea.Cmd {
	d.load()
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg { return catalogTick{} })
}

func (d *drawer) onMsg(msg tea.Msg) tea.Cmd {
	if _, ok := msg.(catalogTick); ok {
		return d.reload()
	}
	return nil
}

// fields are the catalog's groups for p, each a group of checkboxes.
func (d *drawer) fields(p config.Placement) []*terminu.Field {
	var fields []*terminu.Field
	var ids [][]string
	for _, it := range d.cat.Items {
		if len(fields) == 0 || fields[len(fields)-1].Label != it.Group {
			fields = append(fields, &terminu.Field{Label: it.Group, Multi: true})
			ids = append(ids, nil)
		}
		f := fields[len(fields)-1]
		f.Options = append(f.Options, it.Name)
		f.Checked = append(f.Checked, d.cfg.On(p, it))
		ids[len(ids)-1] = append(ids[len(ids)-1], it.ID)
	}
	for i, f := range fields {
		group := ids[i]
		f.OnChange = func() tea.Cmd { return d.toggled(p, f, group) }
	}
	return fields
}

// toggled writes the change of one checkbox in f.
func (d *drawer) toggled(p config.Placement, f *terminu.Field, ids []string) tea.Cmd {
	byID := map[string]config.Item{}
	for _, it := range d.cat.Items {
		byID[it.ID] = it
	}
	for k, id := range ids {
		if on := k < len(f.Checked) && f.Checked[k]; on != d.cfg.On(p, byID[id]) {
			d.cfg.Set(p, d.cat, id, on)
		}
	}
	return d.save()
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
