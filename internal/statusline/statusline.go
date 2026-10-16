// Package statusline is `ccdrawu statusline`, set as Claude Code's status
// line: the items picked for it in ccdrawu on one row. The mod works out each
// item's value and colors (hooks/statusline.ts) into the session's file,
// <data>/sessions/<session id>.json; this lays them out.
package statusline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/vulcanshen/ccdrawu/internal/config"
)

// Seg is a piece of an item's value in its color, as the mod writes it.
type Seg struct {
	Text  string `json:"text"`
	Color string `json:"color"`
	Bold  bool   `json:"bold,omitempty"`
}

// Input is the part of what Claude Code hands its status line on stdin that
// is read here: the session, to find its file, and the few figures shown
// while there is none.
type Input struct {
	SessionID string `json:"session_id"`
	Model     struct {
		ID string `json:"id"`
	} `json:"model"`
	ContextWindow struct {
		UsedPercentage *float64 `json:"used_percentage"`
	} `json:"context_window"`
	Cost struct {
		TotalCostUSD *float64 `json:"total_cost_usd"`
	} `json:"cost"`
}

// Catppuccin Mocha, as the drawer has them (hooks/items.ts): labels in blue,
// the separator in surface.
const (
	textColor     = "#cdd6f4"
	lavenderColor = "#b4befe"
	labelColor    = "#89b4fa"
	sepColor      = "#45475a"
)

// noValue is what an item on reads while it has no value, as in the drawer
// (hooks/items.ts): it keeps its place, so the items after it do not move when
// the value comes.
const noValue = "-"

// edge is the cells of the terminal's width Claude Code keeps off its status
// line: two on either side (a longer line it cuts itself, in the middle of an
// item).
const edge = 4

// hidden is a status line of no row at all: Claude Code keeps the row while a
// status line is set, blank for an empty line, but draws a line of nothing
// but a reset none (measured in Claude Code 2.1.300).
const hidden = "\x1b[0m"

// dockColumns is the terminal's width from which Claude Code docks a pane the
// plugin opens by itself beside the transcript ($.ui.open). One the person
// opened docks from 110; the session's file says when it is docked.
const dockColumns = 144

// keep is how long a session's file stays after its last write: the mod
// cannot delete one, so a session's file outlives it until then.
const keep = 7 * 24 * time.Hour

// Run reads Claude Code's input from in and writes the status line to out:
// the items on in the statusline, from the session's file, or the few the
// input has while there is none (the plugin not loaded, or not yet). columns
// is the terminal's width, 0 when unknown; the line is cut to fit it. On only
// when narrow, there is none while the drawer is beside the transcript, or the
// terminal is wide enough for it; with nothing to show there is none either,
// and the row goes (hidden). Files of sessions gone quiet for a week are
// removed.
func Run(in io.Reader, out io.Writer, loc config.Locations, columns int, now time.Time) error {
	var input Input
	if err := json.NewDecoder(in).Decode(&input); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	cfg, err := config.ReadConfig(loc.ConfigFile)
	if err != nil {
		return err
	}
	cat, ok, err := config.ReadCatalog(filepath.Join(loc.Data, "catalog.json"))
	if err != nil || !ok {
		cat = fallback
	}
	sessions := filepath.Join(loc.Data, "sessions")
	items, docked := sessionItems(sessions, input.SessionID)
	if items == nil {
		items = fromInput(input)
	}
	wide := columns >= dockColumns
	if columns > 0 {
		columns = max(1, columns-edge)
	}
	line := Line(cfg, cat, items, columns)
	if line == "" || cfg.Statusline.When == config.WhenNarrow && (docked || wide) {
		line = hidden
	}
	_, err = fmt.Fprintln(out, line)
	prune(sessions, now)
	return err
}

// Line lays out the items on in the statusline, in the config's order: each
// one's label as its format says and its value, the separator between them;
// an item with nothing to show reads noValue. Highlighted, every piece has its
// color, as Claude Code dims what has none; dimmed, none has, and Claude Code
// dims it all alike. It is cut to columns cells when that is above 0.
func Line(cfg config.Config, cat config.Catalog, items map[string][]Seg, columns int) string {
	sep := cfg.Statusline.Separator
	if sep == "" {
		sep = "|"
	}
	var parts []string
	for _, it := range cfg.Shown(config.Status, cat) {
		segs := items[it.ID]
		if len(segs) == 0 {
			segs = []Seg{{Text: noValue, Color: textColor}}
		}
		var b strings.Builder
		switch label := it.Label; cfg.Statusline.Format {
		case "value":
		case "short":
			if it.Short != "" {
				label = it.Short
			}
			fallthrough
		default:
			b.WriteString(paint(label+" ", labelColor, false))
		}
		for _, s := range segs {
			b.WriteString(paint(s.Text, s.Color, s.Bold))
		}
		parts = append(parts, b.String())
	}
	line := strings.Join(parts, paint(" "+sep+" ", sepColor, false))
	tail := paint("…", sepColor, false)
	if cfg.Statusline.Color == "dimmed" {
		line, tail = ansi.Strip(line), "…"
	}
	if columns > 0 {
		line = ansi.Truncate(line, columns, tail)
	}
	return line
}

// paint is text in a color "#rrggbb" (truecolor), bold or not; a color it
// cannot read leaves the text as it is.
func paint(text, color string, bold bool) string {
	if text == "" {
		return ""
	}
	var b strings.Builder
	if r, g, bl, ok := rgb(color); ok {
		fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm", r, g, bl)
	}
	if bold {
		b.WriteString("\x1b[1m")
	}
	b.WriteString(text)
	b.WriteString("\x1b[0m")
	return b.String()
}

func rgb(color string) (r, g, b uint64, ok bool) {
	if len(color) != 7 || color[0] != '#' {
		return 0, 0, 0, false
	}
	v, err := strconv.ParseUint(color[1:], 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	return v >> 16, v >> 8 & 0xff, v & 0xff, true
}

// sessionItems is the session's file as the mod wrote it: its items, nil
// while there is none or it cannot be read (caught half written: the next run
// reads it), and whether the drawer is beside the transcript.
func sessionItems(dir, id string) (map[string][]Seg, bool) {
	if id == "" || strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		return nil, false
	}
	data, err := os.ReadFile(filepath.Join(dir, id+".json"))
	if err != nil {
		return nil, false
	}
	var f struct {
		Docked bool             `json:"docked"`
		Items  map[string][]Seg `json:"items"`
	}
	if json.Unmarshal(data, &f) != nil || f.Items == nil {
		return nil, false
	}
	return f.Items, f.Docked
}

// fallback is the catalog while the mod has written none: the items the
// input has.
var fallback = config.Catalog{Items: []config.Item{
	{ID: "model", Label: "Model", Short: "mdl", Statusline: true},
	{ID: "context", Label: "Context", Short: "ctx", Statusline: true},
	{ID: "cost", Label: "Cost", Short: "cost", Statusline: true},
}}

// fromInput is the items the input has, read as the drawer reads them
// (hooks/items.ts): the model by its id, as the drawer has it, not the
// input's display name.
func fromInput(in Input) map[string][]Seg {
	items := map[string][]Seg{}
	if in.Model.ID != "" {
		items["model"] = []Seg{{Text: in.Model.ID, Color: lavenderColor, Bold: true}}
	}
	if p := in.ContextWindow.UsedPercentage; p != nil {
		items["context"] = []Seg{{Text: fmt.Sprintf("%d%%", int(math.Round(*p))), Color: usageColor(*p)}}
	}
	if c := in.Cost.TotalCostUSD; c != nil {
		items["cost"] = []Seg{{Text: fmt.Sprintf("$%.2f", *c), Color: textColor}}
	}
	return items
}

// usageColor is the drawer's (hooks/format.ts): green under 50%, yellow under
// 70%, peach under 90%, red from 90%.
func usageColor(percent float64) string {
	switch {
	case percent < 50:
		return "#a6e3a1"
	case percent < 70:
		return "#f9e2af"
	case percent < 90:
		return "#fab387"
	}
	return "#f38ba8"
}

// prune removes the files of sessions not written for keep.
func prune(dir string, now time.Time) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		if info, err := e.Info(); err == nil && now.Sub(info.ModTime()) > keep {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}
