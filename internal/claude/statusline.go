package claude

import (
	"bytes"
	"encoding/json"
	"path/filepath"
)

// StatuslineCommand is what settings.json's statusLine runs while ccdrawu's
// is set: the ccdrawu command where the user's shell finds it, so it keeps
// working when a package manager moves the binary.
const StatuslineCommand = "ccdrawu statusline"

type statusLine struct {
	Type            string `json:"type"`
	Command         string `json:"command"`
	RefreshInterval int    `json:"refreshInterval,omitempty"`
}

// CurrentStatusline is settings.json's statusLine as it stands: the command
// it runs ("" when there is none, or it is not a command), and whether that
// is ccdrawu's.
func CurrentStatusline(dir string) (command string, ours bool, err error) {
	settings, err := readSettings(dir)
	if err != nil {
		return "", false, err
	}
	raw := rawAt(settings, "statusLine")
	var s statusLine
	if raw != nil && json.Unmarshal(raw, &s) == nil {
		return s.Command, s.Command == StatuslineCommand, nil
	}
	return "", false, nil
}

// ApplyStatusline sets settings.json's statusLine to ccdrawu's, run again
// every refresh seconds while nothing happens. When it replaces another,
// switched is true and before is that one as it was ("null" for none), to put
// back; ccdrawu's already set only takes the refresh.
func ApplyStatusline(dir string, refresh int) (before json.RawMessage, switched bool, err error) {
	settings, err := readSettings(dir)
	if err != nil {
		return nil, false, err
	}
	_, ours, err := CurrentStatusline(dir)
	if err != nil {
		return nil, false, err
	}
	if !ours {
		before = rawAt(settings, "statusLine")
		if before == nil {
			before = json.RawMessage("null")
		}
	}
	value, _ := json.Marshal(statusLine{Type: "command", Command: StatuslineCommand, RefreshInterval: refresh})
	if settings, err = setValue(settings, "statusLine", value); err != nil {
		return nil, false, err
	}
	return before, !ours, writeFile(filepath.Join(dir, "settings.json"), settings)
}

// RestoreStatusline puts before back ("null" or nothing takes statusLine
// out) while ccdrawu's is the one set; one set since in its place is left as
// it is.
func RestoreStatusline(dir string, before json.RawMessage) error {
	if _, ours, err := CurrentStatusline(dir); err != nil || !ours {
		return err
	}
	settings, err := readSettings(dir)
	if err != nil {
		return err
	}
	if b := bytes.TrimSpace(before); len(b) == 0 || string(b) == "null" {
		settings, err = removeKey(settings, "statusLine")
	} else {
		settings, err = setValue(settings, "statusLine", b)
	}
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(dir, "settings.json"), settings)
}

// rawAt is the top-level key's value as it is written, nil when it is not
// there.
func rawAt(data []byte, key string) json.RawMessage {
	var m map[string]json.RawMessage
	if json.Unmarshal(data, &m) != nil {
		return nil
	}
	return m[key]
}
