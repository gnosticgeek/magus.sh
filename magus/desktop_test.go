package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The graphical session's PATH has no ~/.local/bin, so a bare command name in
// Exec works from a terminal and silently fails from the launcher (§8).
func TestDesktopEntryExecIsAbsolute(t *testing.T) {
	e := desktopEntry{Name: "magus", Exec: "/home/deck/.local/bin/magus"}
	body := e.render()
	if !strings.Contains(body, "Exec=/home/deck/.local/bin/magus\n") {
		t.Errorf("Exec line wrong:\n%s", body)
	}
	// A TryExec that misses removes the entry from the launcher without any
	// error, turning a repairable problem into an invisible one.
	if strings.Contains(body, "TryExec") {
		t.Error("entries must not carry TryExec")
	}
}

func TestDesktopEntryEscapesUserChosenPaths(t *testing.T) {
	e := desktopEntry{Name: "odd", Exec: `/home/deck/My \Games/run`}
	body := e.render()
	if !strings.Contains(body, `Exec=/home/deck/My \\Games/run`) {
		t.Errorf("backslash not escaped:\n%s", body)
	}
}

func TestWriteDesktopEntryRefusesToClobberForeignEntries(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	path := filepath.Join(c.Paths.Apps, "kitty.desktop")
	theirs := "[Desktop Entry]\nName=Their kitty\nExec=/usr/bin/kitty\n"
	if err := os.WriteFile(path, []byte(theirs), 0o644); err != nil {
		t.Fatal(err)
	}

	err := writeDesktopEntry(c, "kitty.desktop", desktopEntry{Name: "kitty", Exec: "/x"})
	if err == nil {
		t.Fatal("want an error rather than a silent overwrite")
	}
	body, _ := os.ReadFile(path)
	if string(body) != theirs {
		t.Error("someone else's launcher entry was overwritten")
	}
}

// A foreign entry is not drift — the step must not fight the user for the file.
func TestDesktopEntryCurrentIgnoresForeignEntries(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	os.WriteFile(filepath.Join(c.Paths.Apps, "kitty.desktop"),
		[]byte("[Desktop Entry]\nExec=/usr/bin/kitty\n"), 0o644)

	if !desktopEntryCurrent(c, "kitty.desktop", "/anything/else") {
		t.Error("a foreign entry should not be reported as drift")
	}
}

// --- paths ---

func TestPathsFollowTheCanonicalLayout(t *testing.T) {
	for _, v := range []string{"XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CONFIG_HOME"} {
		t.Setenv(v, "")
	}
	p := newPathsUnder("/home/deck")

	for _, tc := range []struct{ got, want string }{
		{p.Bin, "/home/deck/.local/bin"},
		{p.Data, "/home/deck/.local/share/magus"},
		{p.State, "/home/deck/.local/state/magus"},
		{p.Config, "/home/deck/.config/magus"},
		{p.Apps, "/home/deck/.local/share/applications"},
		{p.ManifestPath(), "/home/deck/.config/magus/manifest.toml"},
		{p.AppDir("kitty"), "/home/deck/.local/kitty.app"},
	} {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
}

func TestPathsHonourXDGOverrides(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/var/tmp/state")
	p := newPathsUnder("/home/deck")
	if p.State != "/var/tmp/state/magus" {
		t.Errorf("State = %q, want /var/tmp/state/magus", p.State)
	}
}

// A relative XDG value is invalid per spec; resolving it against the cwd would
// scatter files somewhere unpredictable.
func TestPathsIgnoreRelativeXDGValues(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "relative/path")
	p := newPathsUnder("/home/deck")
	if p.Config != "/home/deck/.config/magus" {
		t.Errorf("Config = %q, want the default when XDG is relative", p.Config)
	}
}

// A crash mid-write leaves a temp behind; without a sweep they accumulate.
// The age floor keeps the sweep from racing a write that is genuinely in flight.
func TestSweepTempsRemovesOnlyStaleTemps(t *testing.T) {
	home := t.TempDir()
	for _, v := range []string{"XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CONFIG_HOME"} {
		t.Setenv(v, "")
	}
	p := newPathsUnder(home)
	if err := p.EnsureDirs(); err != nil {
		t.Fatal(err)
	}

	stale := filepath.Join(p.State, "state.json"+tempMarker+".abc")
	fresh := filepath.Join(p.State, "state.json"+tempMarker+".xyz")
	keep := filepath.Join(p.State, "state.json")
	for _, f := range []string{stale, fresh, keep} {
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-10 * time.Minute)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}

	p.SweepTemps()

	if _, err := os.Stat(stale); err == nil {
		t.Error("stale temp survived the sweep")
	}
	for _, f := range []string{fresh, keep} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("sweep removed %s, which it should not have", filepath.Base(f))
		}
	}
}
