package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Paths resolves the canonical userland layout from the brief (§6). Everything
// magus writes lives under $HOME — nothing in /usr or /opt survives a SteamOS
// atomic update, and nothing here needs root.
//
// Every field is derived from the environment at construction time rather than
// read from a global, so tests can point a Paths at a temp dir and exercise the
// real code.
type Paths struct {
	Home   string // $HOME
	Bin    string // ~/.local/bin           — executables and symlinks into app dirs
	Local  string // ~/.local               — parent of the <app>.app dirs
	Data   string // ~/.local/share/magus   — static assets, templates
	State  string // ~/.local/state/magus   — runtime state, temp renders
	Config string // ~/.config/magus        — manifest, user config, themes
	Apps   string // ~/.local/share/applications — .desktop launcher entries
}

// NewPaths resolves the layout, honouring the XDG variables when they are set
// and falling back to the documented defaults when they are not.
func NewPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}
	return newPathsUnder(home), nil
}

// newPathsUnder builds a layout rooted at a given home. XDG overrides are only
// honoured when they are absolute — a relative XDG_* value is invalid per spec
// and silently resolving it against the cwd would scatter files unpredictably.
func newPathsUnder(home string) Paths {
	xdg := func(env, def string) string {
		if v := os.Getenv(env); filepath.IsAbs(v) {
			return v
		}
		return filepath.Join(home, def)
	}
	dataHome := xdg("XDG_DATA_HOME", ".local/share")
	return Paths{
		Home:   home,
		Bin:    filepath.Join(home, ".local", "bin"),
		Local:  filepath.Join(home, ".local"),
		Data:   filepath.Join(dataHome, "magus"),
		State:  filepath.Join(xdg("XDG_STATE_HOME", ".local/state"), "magus"),
		Config: filepath.Join(xdg("XDG_CONFIG_HOME", ".config"), "magus"),
		Apps:   filepath.Join(dataHome, "applications"),
	}
}

// ManifestPath is where the wizard writes and the reconciler reads.
func (p Paths) ManifestPath() string { return filepath.Join(p.Config, "manifest.toml") }

// AppDir is the self-contained install directory for a third-party app,
// e.g. ~/.local/kitty.app.
func (p Paths) AppDir(app string) string { return filepath.Join(p.Local, app+".app") }

// EnsureDirs creates the directories magus owns. Safe to call on every run.
func (p Paths) EnsureDirs() error {
	for _, d := range []string{p.Bin, p.Data, p.State, p.Config, p.Apps} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// SweepTemps deletes orphaned atomic-write temp files older than a minute.
// A crash mid-write leaves one behind; without this they accumulate forever.
// The age floor keeps it from racing a write that is genuinely in flight.
func (p Paths) SweepTemps() {
	for _, dir := range []string{p.State, p.Config} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.Contains(e.Name(), tempMarker) {
				continue
			}
			info, err := e.Info()
			if err != nil || time.Since(info.ModTime()) < time.Minute {
				continue
			}
			os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

// tempMarker identifies our atomic-write temp files so SweepTemps can find them
// without risking anything else in the directory.
const tempMarker = ".magus-tmp"

// writeFileAtomic writes data to path via a uniquely-named temp file in the same
// directory, then renames over the target. The rename is atomic within a
// filesystem, so a reader either sees the whole old file or the whole new one —
// never a half-written manifest.
//
// The temp name must be unique per write, not per process: two goroutines
// writing at once would otherwise collide on the same temp and interleave.
// os.CreateTemp's random suffix gives that for free.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, filepath.Base(path)+tempMarker+".*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp) // no-op once the rename below succeeds

	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	// fsync before rename: a rename can otherwise land in the directory while
	// the file's contents are still only in page cache, so a power loss leaves
	// a correctly-named empty file.
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
