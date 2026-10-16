package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewerComparesPartByPart(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"0.1.3", "0.1.2", true},
		{"0.1.10", "0.1.9", true},
		{"0.2", "0.1.9", true},
		{"0.1.2", "0.1.2", false},
		{"0.1.2", "0.1.3", false},
		{"1.0.0", "0.9.9", true},
	} {
		if got := Newer(c.a, c.b); got != c.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestHowTellsTheMethodFromThePath(t *testing.T) {
	for path, want := range map[string]Method{
		"/opt/homebrew/Cellar/ccdrawu/0.1.2/bin/ccdrawu":              Homebrew,
		"/home/linuxbrew/.linuxbrew/Cellar/ccdrawu/0.1.2/bin/ccdrawu": Homebrew,
		`C:\Users\me\scoop\apps\ccdrawu\0.1.2\ccdrawu.exe`:            Scoop,
		`C:\Users\me\AppData\Local\ccdrawu\ccdrawu.exe`:               Self,
		"/home/me/.local/bin/ccdrawu":                                 Self,
		"/usr/local/bin/ccdrawu":                                      Self,
	} {
		if got := How(path); got != want {
			t.Errorf("How(%q) = %v, want %v", path, got, want)
		}
	}
	if got := BrewBin("/opt/homebrew/Cellar/ccdrawu/0.1.0/bin/ccdrawu"); got != "/opt/homebrew/bin/ccdrawu" {
		t.Errorf("BrewBin = %q", got)
	}
}

// release serves version's checksums.txt and archive for goos, the archive
// holding exe; corrupt changes the archive after its checksum is taken
func release(t *testing.T, version, goos string, exe []byte, corrupt bool) Release {
	t.Helper()
	name := Archive(version, goos, "arm64")
	var archive bytes.Buffer
	if goos == "windows" {
		zw := zip.NewWriter(&archive)
		for file, body := range map[string][]byte{"README.md": []byte("readme"), "ccdrawu.exe": exe} {
			w, _ := zw.Create(file)
			_, _ = w.Write(body)
		}
		_ = zw.Close()
	} else {
		gz := gzip.NewWriter(&archive)
		tw := tar.NewWriter(gz)
		for _, f := range []struct {
			name string
			body []byte
		}{{"LICENSE", []byte("license")}, {"ccdrawu", exe}} {
			_ = tw.WriteHeader(&tar.Header{Name: f.name, Mode: 0o755, Size: int64(len(f.body)), Typeflag: tar.TypeReg})
			_, _ = tw.Write(f.body)
		}
		_ = tw.Close()
		_ = gz.Close()
	}
	sum := sha256.Sum256(archive.Bytes())
	sums := hex.EncodeToString(sum[:]) + "  " + name + "\n" + strings.Repeat("0", 64) + "  other.tar.gz\n"
	if corrupt {
		archive.WriteString("x")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/latest":
			_, _ = w.Write([]byte(`{"tag_name": "v` + version + `"}`))
		case "/download/v" + version + "/checksums.txt":
			_, _ = w.Write([]byte(sums))
		case "/download/v" + version + "/" + name:
			_, _ = w.Write(archive.Bytes())
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return Release{Latest: srv.URL + "/latest", Download: srv.URL + "/download", Client: srv.Client()}
}

func TestReplacePutsTheCheckedExecutableInPlace(t *testing.T) {
	r := release(t, "0.1.3", "darwin", []byte("new ccdrawu"), false)
	if v, err := r.LatestVersion(); err != nil || v != "0.1.3" {
		t.Fatalf("LatestVersion = %q, %v", v, err)
	}
	target := filepath.Join(t.TempDir(), "ccdrawu")
	if err := os.WriteFile(target, []byte("old ccdrawu"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := r.Replace(target, "0.1.3", "darwin", "arm64"); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(target); string(got) != "new ccdrawu" {
		t.Fatalf("target holds %q", got)
	}
	if fi, _ := os.Stat(target); runtime.GOOS != "windows" && fi.Mode().Perm() != 0o755 {
		t.Fatalf("target mode %v", fi.Mode())
	}
	// Nothing else left beside it
	if entries, _ := os.ReadDir(filepath.Dir(target)); len(entries) != 1 {
		t.Fatalf("left %v", entries)
	}
}

func TestReplaceOnWindowsMovesTheRunningOneAside(t *testing.T) {
	r := release(t, "0.1.3", "windows", []byte("new ccdrawu.exe"), false)
	dir := t.TempDir()
	target := filepath.Join(dir, "ccdrawu.exe")
	_ = os.WriteFile(target, []byte("old ccdrawu.exe"), 0o755)
	// An .old from the update before goes
	_ = os.WriteFile(target+".old", []byte("older"), 0o755)
	if err := r.Replace(target, "0.1.3", "windows", "arm64"); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(target); string(got) != "new ccdrawu.exe" {
		t.Fatalf("target holds %q", got)
	}
	if got, _ := os.ReadFile(target + ".old"); string(got) != "old ccdrawu.exe" {
		t.Fatalf(".old holds %q", got)
	}
}

func TestReplaceRefusesAnArchiveThatFailsItsChecksum(t *testing.T) {
	r := release(t, "0.1.3", "linux", []byte("tampered"), true)
	target := filepath.Join(t.TempDir(), "ccdrawu")
	_ = os.WriteFile(target, []byte("old ccdrawu"), 0o755)
	err := r.Replace(target, "0.1.3", "linux", "arm64")
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("Replace = %v, want a checksum error", err)
	}
	if got, _ := os.ReadFile(target); string(got) != "old ccdrawu" {
		t.Fatalf("target holds %q", got)
	}
}

func TestReplaceStopsWhereItCannotWrite(t *testing.T) {
	if runtime.GOOS == "windows" || os.Getuid() == 0 {
		t.Skip("folder permissions as a user on unix")
	}
	r := release(t, "0.1.3", "linux", []byte("new ccdrawu"), false)
	dir := t.TempDir()
	target := filepath.Join(dir, "ccdrawu")
	_ = os.WriteFile(target, []byte("old ccdrawu"), 0o755)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	err := r.Replace(target, "0.1.3", "linux", "arm64")
	if err == nil || !strings.Contains(err.Error(), "cannot write in "+dir) || !errors.Is(err, os.ErrPermission) {
		t.Fatalf("Replace = %v, want cannot write in %s", err, dir)
	}
}
