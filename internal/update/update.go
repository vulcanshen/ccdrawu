// Package update brings ccdrawu to its latest release: through the package
// manager that installed it, or by putting the release's executable in place
// of its own, checked against the release's checksums.
package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"cmp"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Release is where the latest release is looked up, in GitHub's API, and
// where its files download from: <Download>/v<version>/<name>.
type Release struct {
	Latest   string
	Download string
	Client   *http.Client
}

// GitHub is ccdrawu's releases, the ones the install scripts take.
var GitHub = Release{
	Latest:   "https://api.github.com/repos/vulcanshen/ccdrawu/releases/latest",
	Download: "https://github.com/vulcanshen/ccdrawu/releases/download",
	Client:   &http.Client{Timeout: 2 * time.Minute},
}

// A release's archive is a few MB; anything past this is not one
const maxDownload = 64 << 20

// LatestVersion is the latest release's version, without its v: 0.1.3.
// GitHub's latest is never a pre-release, so a beta is offered to no one.
func (r Release) LatestVersion() (string, error) {
	body, err := r.get(r.Latest)
	if err != nil {
		return "", err
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &release); err != nil || release.TagName == "" {
		return "", fmt.Errorf("%s: no release tag in its answer", r.Latest)
	}
	return strings.TrimPrefix(release.TagName, "v"), nil
}

func (r Release) get(url string) ([]byte, error) {
	resp, err := r.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxDownload))
}

// Newer reports whether version a is newer than b, as semver orders them:
// part by numeric part, 0.1.10 newer than 0.1.9; a release newer than its
// pre-releases, 1.0.1 than 1.0.1-beta.1, and those newer than 1.0.0.
func Newer(a, b string) bool {
	na, pa := split(a)
	nb, pb := split(b)
	if c := compareNumbers(na, nb); c != 0 {
		return c > 0
	}
	if pa == "" || pb == "" {
		return pa == "" && pb != ""
	}
	return comparePre(pa, pb) > 0
}

// split takes a version apart into its numbers and its pre-release, the
// build metadata left out: 1.0.1-beta.1+dirty is 1.0.1 and beta.1.
func split(v string) (numbers, pre string) {
	v, _, _ = strings.Cut(v, "+")
	numbers, pre, _ = strings.Cut(v, "-")
	return numbers, pre
}

// compareNumbers compares dotted numbers part by part, a missing part 0
func compareNumbers(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := range max(len(pa), len(pb)) {
		if c := cmp.Compare(part(pa, i), part(pb, i)); c != 0 {
			return c
		}
	}
	return 0
}

// comparePre compares pre-releases identifier by identifier, as semver does:
// numbers by value and before words, words by their bytes (beta before rc),
// and where all are the same, the one with more is newer.
func comparePre(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := range min(len(pa), len(pb)) {
		x, errX := strconv.Atoi(pa[i])
		y, errY := strconv.Atoi(pb[i])
		c := 0
		switch {
		case errX == nil && errY == nil:
			c = cmp.Compare(x, y)
		case errX == nil:
			c = -1
		case errY == nil:
			c = 1
		default:
			c = strings.Compare(pa[i], pb[i])
		}
		if c != 0 {
			return c
		}
	}
	return cmp.Compare(len(pa), len(pb))
}

func part(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, _ := strconv.Atoi(parts[i])
	return n
}

// Method is how a ccdrawu was installed, and so how it updates.
type Method int

const (
	// Self is the install script's, or a download: it replaces its own executable
	Self Method = iota
	Homebrew
	Scoop
)

// How tells the method from where the executable really is, symlinks
// followed: Homebrew keeps it under Cellar/ccdrawu, Scoop under
// scoop/apps/ccdrawu.
func How(exe string) Method {
	p := strings.ToLower(strings.ReplaceAll(exe, `\`, "/"))
	switch {
	case strings.Contains(p, "/cellar/ccdrawu/"):
		return Homebrew
	case strings.Contains(p, "/scoop/apps/ccdrawu/"):
		return Scoop
	}
	return Self
}

// BrewBin is the ccdrawu Homebrew links into its prefix's bin, for an
// executable under its Cellar: after an upgrade that is the new one, where the
// Cellar path is the old version's.
func BrewBin(exe string) string {
	if i := strings.Index(exe, "/Cellar/"); i >= 0 {
		return exe[:i] + "/bin/ccdrawu"
	}
	return exe
}

// Archive is the name of a release's archive for a platform, as goreleaser
// names it: a zip for Windows, a tar.gz for the rest.
func Archive(version, goos, goarch string) string {
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("ccdrawu_%s_%s_%s.%s", version, goos, goarch, ext)
}

// Replace puts the executable of release version for goos and goarch in place
// of target: its archive downloaded and checked against the release's
// checksums.txt, the executable written beside target, then renamed over it.
// Windows does not overwrite a running executable but renames one, so there
// target moves to target.old first, which the next update removes. A folder it
// cannot write in fails before anything downloads: it never asks for more
// rights.
func (r Release) Replace(target, version, goos, goarch string) error {
	tmp, err := os.CreateTemp(filepath.Dir(target), ".ccdrawu-update-*")
	if err != nil {
		return fmt.Errorf("cannot write in %s: %w", filepath.Dir(target), err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	name := Archive(version, goos, goarch)
	sums, err := r.get(r.file(version, "checksums.txt"))
	if err != nil {
		return err
	}
	archive, err := r.get(r.file(version, name))
	if err != nil {
		return err
	}
	if err := check(archive, name, sums); err != nil {
		return err
	}
	exe, err := extract(archive, goos)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if _, err := tmp.Write(exe); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		return err
	}
	if goos != "windows" {
		return os.Rename(tmp.Name(), target)
	}
	old := target + ".old"
	_ = os.Remove(old)
	if err := os.Rename(target, old); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), target); err != nil {
		_ = os.Rename(old, target)
		return err
	}
	return nil
}

func (r Release) file(version, name string) string {
	return r.Download + "/v" + version + "/" + name
}

// check holds the archive to its line in checksums.txt: "<sha256>  <name>".
func check(archive []byte, name string, sums []byte) error {
	for line := range strings.Lines(string(sums)) {
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[1] != name {
			continue
		}
		sum := sha256.Sum256(archive)
		if hex.EncodeToString(sum[:]) != strings.ToLower(fields[0]) {
			return fmt.Errorf("%s does not match its checksum in checksums.txt: not installed", name)
		}
		return nil
	}
	return fmt.Errorf("checksums.txt has no line for %s: not installed", name)
}

// extract is the ccdrawu executable out of a release archive.
func extract(archive []byte, goos string) ([]byte, error) {
	if goos == "windows" {
		zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if path.Base(f.Name) != "ccdrawu.exe" {
				continue
			}
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(io.LimitReader(rc, maxDownload))
		}
		return nil, errors.New("no ccdrawu.exe in it")
	}
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil, errors.New("no ccdrawu in it")
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag == tar.TypeReg && path.Base(h.Name) == "ccdrawu" {
			return io.ReadAll(io.LimitReader(tr, maxDownload))
		}
	}
}
