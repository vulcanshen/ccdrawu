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

// drawer is the TUI: one screen, a panel per placement, each a group of
// checkboxes per catalog group. A toggle is written at once, and every open
// Claude Code session takes it within a second.
type drawer struct {
	app    *terminu.App
	loc    config.Locations
	cat    config.Catalog
	stamp  time.Time // the catalog's mtime when it was read
	cfg    config.Config
	panels map[config.Placement]*terminu.Panel
	values map[config.Placement]*terminu.Values
}

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
		loc:    loc,
		cfg:    cfg,
		panels: map[config.Placement]*terminu.Panel{},
		values: map[config.Placement]*terminu.Values{},
	}
	a := terminu.New("ccdrawu")
	a.Splash = splash()
	a.Status = d.status
	a.OnMsg = d.onMsg
	a.Start = func() tea.Cmd { return d.reload() }
	d.app = a

	for _, p := range []config.Placement{config.Dock, config.Inline} {
		v := &terminu.Values{}
		d.values[p] = v
		d.panels[p] = &terminu.Panel{Label: placementLabel(p), Body: v, Fail: &terminu.Failure{
			Fact: "Nothing to list until ccdrawu runs in Claude Code",
			Hint: "Run ccdrawu install, then start Claude Code",
		}}
	}
	a.AddScreen(&terminu.Screen{
		Name: "Items", Key: "I", Desc: "which items the drawer shows",
		Panels:      []*terminu.Panel{d.panels[config.Dock], d.panels[config.Inline]},
		NarrowBelow: 60,
	})
	d.load()
	return d, nil
}

func placementLabel(p config.Placement) string {
	if p == config.Dock {
		return "Beside the transcript"
	}
	return "Above the prompt"
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
		for _, p := range d.panels {
			p.Fail = &terminu.Failure{Fact: "The catalog can't be read: " + err.Error(), Hint: "Start Claude Code again to write it anew"}
		}
		return
	}
	if !ok {
		return
	}
	d.cat, d.stamp = cat, info.ModTime()
	for p, v := range d.values {
		v.Fields = d.fields(p)
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
