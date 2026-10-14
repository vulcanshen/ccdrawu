// Package statusline is `ccdrawu statusline`, the command Claude Code's
// statusLine setting runs. Claude Code runs it again on each new message, vim
// mode switch and permission mode change, handing it JSON on stdin. It draws
// nothing: it only writes vim's mode where the mod reads it, since a mod
// cannot read vim's live state.
package statusline

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// keep is how long a session's file stays untouched before it goes.
const keep = 7 * 24 * time.Hour

// A session id is a file name: nothing that could reach out of the folder.
var sessionID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// Write reads the status line's JSON from r and writes vim's mode ("" with
// vim off) to <data>/vim/<session id>, the file the mod reads; files of
// sessions untouched for a week go.
func Write(r io.Reader, data string, now time.Time) error {
	var in struct {
		SessionID string `json:"session_id"`
		Vim       struct {
			Mode string `json:"mode"`
		} `json:"vim"`
	}
	if err := json.NewDecoder(r).Decode(&in); err != nil {
		return err
	}
	dir := filepath.Join(data, "vim")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if sessionID.MatchString(in.SessionID) {
		// A temporary file renamed over it, so the mod never reads half of one
		target := filepath.Join(dir, in.SessionID)
		if err := os.WriteFile(target+".tmp", []byte(in.Vim.Mode), 0o644); err != nil {
			return err
		}
		if err := os.Rename(target+".tmp", target); err != nil {
			return err
		}
	}
	prune(dir, now.Add(-keep))
	return nil
}

// prune removes the files in dir last written before cutoff.
func prune(dir string, cutoff time.Time) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if info, err := e.Info(); err == nil && info.Mode().IsRegular() && info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}
