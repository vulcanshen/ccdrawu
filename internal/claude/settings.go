package claude

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Dir is Claude Code's own folder, settings.json and themes/ in it:
// CLAUDE_CONFIG_DIR, as Claude Code reads it, else ~/.claude.
func Dir(getenv func(string) string, home string) string {
	if d := getenv("CLAUDE_CONFIG_DIR"); d != "" {
		return d
	}
	return filepath.Join(home, ".claude")
}

// field is where one key of the top-level object sits in the file: from the
// start of its line's indent (or just past the { or , before it, on a line it
// shares) to the end of its value.
type field struct {
	key                   string
	lineStart, valueStart int
	valueEnd              int
}

// fields lists the top-level keys of a JSON object, with where each sits.
func fields(data []byte) ([]field, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil, errors.New("settings.json is not a JSON object")
	}
	var out []field
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, _ := t.(string)
		keyEnd := int(dec.InputOffset())
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		end := int(dec.InputOffset())
		keyStart := bytes.LastIndexByte(data[:keyEnd-1], '"')
		lineStart := max(bytes.LastIndexByte(data[:keyStart], '\n'), bytes.LastIndexAny(data[:keyStart], "{,")) + 1
		out = append(out, field{key: key, lineStart: lineStart, valueStart: end - len(raw), valueEnd: end})
	}
	return out, nil
}

// setString sets the top-level key to a string value, leaving every other byte
// as it was; a key not there goes first, indented as the next key is.
func setString(data []byte, key, value string) ([]byte, error) {
	quoted, _ := json.Marshal(value)
	return setValue(data, key, quoted)
}

// setValue is setString for any JSON value, written as given.
func setValue(data []byte, key string, quoted []byte) ([]byte, error) {
	fs, err := fields(data)
	if err != nil {
		return nil, err
	}
	for _, f := range fs {
		if f.key == key {
			return join(data[:f.valueStart], quoted, data[f.valueEnd:]), nil
		}
	}
	open := bytes.IndexByte(data, '{') + 1
	entry := fmt.Sprintf("%q: %s", key, quoted)
	if len(fs) == 0 {
		return join(data[:open], []byte("\n  "+entry+"\n"), data[bytes.LastIndexByte(data, '}'):]), nil
	}
	if bytes.IndexByte(data[open:fs[0].lineStart], '\n') < 0 {
		// One line: the next key sits on the brace's line
		return join(data[:open], []byte(entry+", "), data[open:]), nil
	}
	indent := data[fs[0].lineStart : bytes.IndexByte(data[fs[0].lineStart:], '"')+fs[0].lineStart]
	return join(data[:open], []byte("\n"+string(indent)+entry+","), data[open:]), nil
}

// removeKey takes the top-level key out, with its line and its comma.
func removeKey(data []byte, key string) ([]byte, error) {
	fs, err := fields(data)
	if err != nil {
		return nil, err
	}
	for i, f := range fs {
		if f.key != key {
			continue
		}
		end := f.valueEnd
		if i < len(fs)-1 {
			// Up to the next key's line: the comma goes with this one
			end = fs[i+1].lineStart
			return join(data[:f.lineStart], data[end:]), nil
		}
		if i == 0 {
			// The only key: leave the object empty
			return join(data[:bytes.IndexByte(data, '{')+1], data[bytes.LastIndexByte(data, '}'):]), nil
		}
		// The last key: the comma before it goes
		prev := fs[i-1].valueEnd
		return join(data[:prev], data[end:]), nil
	}
	return data, nil
}

// stringAt reads the top-level key's string value, "" when it is not there.
func stringAt(data []byte, key string) string {
	var m map[string]json.RawMessage
	if json.Unmarshal(data, &m) != nil {
		return ""
	}
	var s string
	_ = json.Unmarshal(m[key], &s)
	return s
}

func join(parts ...[]byte) []byte { return bytes.Join(parts, nil) }

// writeFile replaces path at once (a temporary file renamed over it), keeping
// its permissions.
func writeFile(path string, data []byte) error {
	mode := os.FileMode(0o644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".ccdrawu.tmp"
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
