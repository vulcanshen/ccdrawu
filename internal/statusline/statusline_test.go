package statusline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteKeepsVimsModePerSession(t *testing.T) {
	data := t.TempDir()
	now := time.Now()
	for _, mode := range []string{"INSERT", "NORMAL"} {
		in := `{"session_id": "46daa4a7-76e0", "vim": {"mode": "` + mode + `"}, "model": {"id": "x"}}`
		if err := Write(strings.NewReader(in), data, now); err != nil {
			t.Fatal(err)
		}
		if got, _ := os.ReadFile(filepath.Join(data, "vim", "46daa4a7-76e0")); string(got) != mode {
			t.Fatalf("vim file holds %q, want %q", got, mode)
		}
	}
	// vim off: the mode is empty, and the file says so
	if err := Write(strings.NewReader(`{"session_id": "46daa4a7-76e0"}`), data, now); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(data, "vim", "46daa4a7-76e0")); len(got) != 0 {
		t.Fatalf("vim file holds %q with vim off, want nothing", got)
	}
}

func TestWriteStaysInItsFolder(t *testing.T) {
	data := t.TempDir()
	for _, id := range []string{"", "../escape", "a/b", `a\b`, ".hidden"} {
		in := `{"session_id": "` + strings.ReplaceAll(id, `\`, `\\`) + `", "vim": {"mode": "INSERT"}}`
		if err := Write(strings.NewReader(in), data, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	entries, _ := os.ReadDir(filepath.Join(data, "vim"))
	if len(entries) != 0 {
		t.Fatalf("wrote %d files for ids that are no file name", len(entries))
	}
	if _, err := os.Stat(filepath.Join(data, "escape")); err == nil {
		t.Fatal("a session id reached out of the vim folder")
	}
}

func TestWriteDropsSessionsAWeekOld(t *testing.T) {
	data := t.TempDir()
	dir := filepath.Join(data, "vim")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	old, fresh := filepath.Join(dir, "old"), filepath.Join(dir, "fresh")
	for _, f := range []string{old, fresh} {
		if err := os.WriteFile(f, []byte("NORMAL"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chtimes(old, now.Add(-8*24*time.Hour), now.Add(-8*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := Write(strings.NewReader(`{"session_id": "new", "vim": {"mode": "INSERT"}}`), data, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(old); err == nil {
		t.Fatal("a session untouched for 8 days is still there")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Fatal("a fresh session went:", err)
	}
}

func TestWriteSaysWhenItIsNotJSON(t *testing.T) {
	if err := Write(strings.NewReader("not json"), t.TempDir(), time.Now()); err == nil {
		t.Fatal("no error for input that is not JSON")
	}
}
