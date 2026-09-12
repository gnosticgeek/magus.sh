package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"howett.net/plist"
)

func macTestContext(t *testing.T) *Context {
	t.Helper()
	// Keep fixtures inside the temporary home even on Linux CI hosts that set
	// absolute XDG paths for the runner account.
	for _, name := range []string{"XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CONFIG_HOME"} {
		t.Setenv(name, "")
	}
	p := newPathsUnder(t.TempDir())
	if err := p.EnsureDirs(); err != nil {
		t.Fatal(err)
	}
	return &Context{Paths: p, Manifest: newMacManifest(), Device: Device{Kind: DeviceMac}, Brew: "fixture-brew", AppRoots: []string{filepath.Join(p.Home, "Applications")}, Report: &Reporter{Out: io.Discard}, Timeout: time.Second}
}
func TestMacManifestIsolation(t *testing.T) {
	m := newMacManifest()
	m.Mac.Packages = []string{"git", "rectangle"}
	m.Mac.Settings = []string{"finder-pathbar"}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, s := range StepsFor(m) {
		switch s.(type) {
		case brewStep, preferenceStep:
		default:
			t.Fatalf("Linux step %T in Mac plan", s)
		}
	}
	if manifestPlatformCheck(m, "linux") == nil {
		t.Fatal("Mac manifest accepted on Linux")
	}
	legacy := DefaultManifest(Device{Kind: DeviceDeck})
	legacy.Magus.Version = "0.3.0"
	if manifestPlatformCheck(legacy, "darwin") == nil {
		t.Fatal("legacy Linux manifest accepted on Mac")
	}
	if !legacy.Migrate() || legacy.Choices.Terminal != "kitty" {
		t.Fatal("legacy migration lost choices")
	}
	m.Mac.Packages = append(m.Mac.Packages, "unknown")
	if m.Validate() == nil {
		t.Fatal("unknown ID accepted")
	}
	future := newMacManifest()
	future.Magus.Version = "99.0"
	if future.Migrate() || future.Validate() == nil {
		t.Fatal("future schema silently accepted")
	}
}
func TestMacBrewInstallVerifyAndRerun(t *testing.T) {
	for _, id := range []string{"git", "rectangle"} {
		t.Run(id, func(t *testing.T) {
			c := macTestContext(t)
			p, _ := macPackage(id)
			installed := false
			installs := 0
			c.Execute = func(ctx context.Context, name string, args ...string) (string, error) {
				if args[0] == "list" {
					if installed {
						return id + "\n", nil
					}
					return "", nil
				}
				if !reflect.DeepEqual(args, []string{"install", "--" + p.Kind, id}) {
					t.Fatalf("unexpected invocation: %v", args)
				}
				installs++
				installed = true
				return "done", nil
			}
			s := brewStep{p}
			r := executeMacStep(c, s, false)
			if r.Err != nil || !r.Changed {
				t.Fatalf("first install: %+v", r)
			}
			r = executeMacStep(c, s, false)
			if r.Err != nil || r.Changed || r.Before != StateOK || installs != 1 {
				t.Fatalf("rerun: %+v / %d", r, installs)
			}
		})
	}
}
func TestMacBrewUnknownExternalAndDryRun(t *testing.T) {
	c := macTestContext(t)
	p, _ := macPackage("rectangle")
	s := brewStep{p}
	calls := 0
	c.Execute = func(_ context.Context, _ string, args ...string) (string, error) {
		calls++
		return "", errors.New("inspection failed")
	}
	r := executeMacStep(c, s, false)
	if r.Err == nil || calls != 1 {
		t.Fatal("applied after failed inspection")
	}
	c.Execute = func(_ context.Context, _ string, args ...string) (string, error) {
		if args[0] != "list" {
			t.Fatal("unexpected install")
		}
		return "", nil
	}
	if err := os.MkdirAll(filepath.Join(c.AppRoots[0], p.AppBundle), 0700); err != nil {
		t.Fatal(err)
	}
	r = executeMacStep(c, s, false)
	if r.Err != nil || r.Before != StateNotApplicable {
		t.Fatalf("external app: %+v", r)
	}
	c.AppRoots = []string{}
	c.DryRun = true
	r = executeMacStep(c, s, false)
	if r.Err != nil || !r.Changed {
		t.Fatalf("dry run: %+v", r)
	}
}
func preferenceFixture(c *Context, values map[string]interface{}) {
	c.Execute = func(_ context.Context, _ string, args ...string) (string, error) {
		switch args[0] {
		case "export":
			b, err := plist.Marshal(values, plist.XMLFormat)
			return string(b), err
		case "write":
			switch args[3] {
			case "-bool":
				values[args[2]] = args[4] == "true"
			case "-string":
				values[args[2]] = args[4]
			default:
				return "", fmt.Errorf("unexpected type")
			}
			return "", nil
		case "delete":
			delete(values, args[2])
			return "", nil
		}
		return "", fmt.Errorf("unexpected command %v", args)
	}
}
func TestMacPreferenceRestoreOriginalAndAbsent(t *testing.T) {
	for _, original := range []interface{}{nil, false, true, "custom"} {
		t.Run(fmt.Sprint(original), func(t *testing.T) {
			c := macTestContext(t)
			s := preferenceStep{macSettings[0]}
			values := map[string]interface{}{}
			if original != nil {
				values[s.Setting.Key] = original
			}
			preferenceFixture(c, values)
			r := executeMacStep(c, s, false)
			if r.Err != nil {
				t.Fatal(r.Err)
			}
			// Applying twice must never replace the pre-Magus snapshot.
			r = executeMacStep(c, s, false)
			if r.Err != nil {
				t.Fatal(r.Err)
			}
			if err := s.Remove(c); err != nil {
				t.Fatal(err)
			}
			got, ok := values[s.Setting.Key]
			if original == nil && ok {
				t.Fatal("original absence not restored")
			}
			if original != nil && !reflect.DeepEqual(got, original) {
				t.Fatalf("got %v, want %v", got, original)
			}
			snap, err := loadSnapshots(c)
			if err != nil || len(snap) != 0 {
				t.Fatalf("snapshot left after restore: %v %v", snap, err)
			}
		})
	}
}
func TestMacPreferenceConflictAndFailedRead(t *testing.T) {
	c := macTestContext(t)
	s := preferenceStep{macSettings[0]}
	values := map[string]interface{}{s.Setting.Key: false}
	preferenceFixture(c, values)
	if err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	values[s.Setting.Key] = "external"
	if err := s.Remove(c); err == nil {
		t.Fatal("external change overwritten")
	}
	if values[s.Setting.Key] != "external" {
		t.Fatal("changed external preference")
	}
	c.Execute = func(context.Context, string, ...string) (string, error) { return "", errors.New("denied") }
	if _, err := s.Check(c); err == nil {
		t.Fatal("failed read treated as absent")
	}
}
func TestMacPreferenceDryRunWritesNothing(t *testing.T) {
	c := macTestContext(t)
	s := preferenceStep{macSettings[0]}
	c.DryRun = true
	c.Execute = func(_ context.Context, _ string, args ...string) (string, error) {
		if args[0] != "export" {
			t.Fatal("dry run wrote preference")
		}
		return "<?xml version=\"1.0\"?><plist version=\"1.0\"><dict/></plist>", nil
	}
	r := executeMacStep(c, s, false)
	if r.Err != nil {
		t.Fatal(r.Err)
	}
	if _, err := os.Stat(snapshotsPath(c)); !os.IsNotExist(err) {
		t.Fatal("dry run wrote snapshot")
	}
}
func TestMacLockAndCancellation(t *testing.T) {
	c := macTestContext(t)
	unlock, err := acquireMacLock(c.Paths)
	if err != nil {
		t.Fatal(err)
	}
	if second, err := acquireMacLock(c.Paths); err == nil {
		second()
		t.Fatal("concurrent operation allowed")
	}
	unlock()
	unlock, err = acquireMacLock(c.Paths)
	if err != nil {
		t.Fatal(err)
	}
	unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err = runMacCommand(ctx, nil, "/bin/sh", "-c", "sleep 30 & wait")
	if err == nil || time.Since(start) > 3*time.Second {
		t.Fatalf("process group cancellation failed: %v", err)
	}
}
func press(m *macModel, k string) {
	var msg tea.KeyPressMsg
	switch k {
	case "space":
		msg = tea.KeyPressMsg{Code: ' ', Text: " "}
	case "enter":
		msg = tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		msg = tea.KeyPressMsg{Code: tea.KeyEsc}
	case "backspace":
		msg = tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "down":
		msg = tea.KeyPressMsg{Code: tea.KeyDown}
	case "up":
		msg = tea.KeyPressMsg{Code: tea.KeyUp}
	default:
		msg = tea.KeyPressMsg{Code: []rune(k)[0], Text: k}
	}
	m.Update(msg)
}

func TestMacBackspaceNavigatesBackButEditsSearch(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category = macScreenBrowse, "tools"
	press(m, "backspace")
	if m.screen != macScreenMenu {
		t.Fatal("backspace did not return to menu")
	}
	press(m, "/")
	press(m, "a")
	press(m, "backspace")
	if !m.searching || m.search.Value() != "" {
		t.Fatal("backspace should edit the active search field")
	}
}

func TestMacMenuOmitsPresetsAndOffersSelfUpdate(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	foundUpdate := false
	for _, row := range m.rows() {
		if row.ID == "presets" {
			t.Fatal("presets menu is still visible")
		}
		foundUpdate = foundUpdate || row.ID == "self-update"
	}
	if !foundUpdate {
		t.Fatal("self-update menu is missing")
	}
	m.screen = macScreenSelfUpdate
	press(m, "enter")
	if m.screen != macScreenMenu || !strings.Contains(m.notice, "not updated") {
		t.Fatal("preview self-update must not mutate the executable")
	}
}

func TestMagusVersionNewer(t *testing.T) {
	for _, test := range []struct {
		latest, current string
		want            bool
	}{
		{"v0.5.0", "v0.4.0", true},
		{"v0.4.1", "v0.4.0", true},
		{"v0.4.0", "v0.4.0", false},
		{"v0.3.9", "v0.4.0", false},
		{"v0.4.0", "dev", true},
		{"invalid", "v0.4.0", false},
	} {
		if got := magusVersionNewer(test.latest, test.current); got != test.want {
			t.Errorf("magusVersionNewer(%q, %q) = %t, want %t", test.latest, test.current, got, test.want)
		}
	}
}

func TestMagusUpdateCheckPublishesOnlyCurrentResult(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	old := m.selfUpdateGeneration.next()
	current := m.selfUpdateGeneration.next()
	m.Update(macSelfUpdateChecked{available: true, latest: "v0.5.0", generation: old})
	if m.magUpdateAvailable {
		t.Fatal("stale self-update check changed the menu")
	}
	m.Update(macSelfUpdateChecked{available: true, latest: "v0.5.0", generation: current})
	if !m.magUpdateAvailable || m.magUpdateLatest != "v0.5.0" {
		t.Fatal("current self-update check did not reach the menu")
	}
}
func TestMacMenuBasketPresetsAndSearch(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, c.Paths.ManifestPath(), newMacManifest(), true, time.Second)
	press(m, "enter")
	press(m, "enter") // Apps -> Browsers -> package list.
	press(m, "space")
	if !m.selected["firefox"] {
		t.Fatal("selection missing")
	}
	press(m, "esc")
	m.screen = macScreenPresets
	press(m, "enter")
	if !m.selected["firefox"] || !m.selected["rectangle"] {
		t.Fatal("preset removed earlier selection")
	}
	m.screen = macScreenMenu
	press(m, "/")
	for _, k := range []string{"r", "q", "?"} {
		press(m, k)
	}
	if m.search.Value() != "rq?" || m.showHelp || !m.selected["firefox"] {
		t.Fatal("search characters activated global shortcut")
	}
	press(m, "esc")
	m.screen = macScreenReview
	if len(m.rows()) != 4 {
		t.Fatalf("basket lost picks: %d", len(m.rows()))
	}
	if err := m.selectionManifest().Validate(); err != nil {
		t.Fatal(err)
	}
}
func TestMacTerminalFitsAndFocusScrolls(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen = macScreenBrowse
	m.category = "tools"
	m.cursor = 14
	for _, size := range [][2]int{{80, 24}, {72, 20}, {120, 35}} {
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		v := m.View().Content
		if lipgloss.Width(v) > size[0] || lipgloss.Height(v) > size[1] {
			t.Fatalf("%v overflows: %dx%d", size, lipgloss.Width(v), lipgloss.Height(v))
		}
		if !strings.Contains(v, "tmux") {
			t.Fatal("focused row scrolled out of view")
		}
	}
}
func TestMacBoundedLogsStripEscapes(t *testing.T) {
	got := cleanLog("\x1b[31mhello\x1b[0m\x00\x07")
	if got != "hello" {
		t.Fatalf("unsafe log: %q", got)
	}
	if len(cleanLog(strings.Repeat("x", 10000))) > 2000 {
		t.Fatal("unbounded log")
	}
}

func TestMacInstallShowsContinuousHonestActivity(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen = macScreenInstall
	m.started = time.Now().Add(-2 * time.Second)
	m.outcomes = []macOutcome{{ID: "package:jq", Name: "jq", Status: "unfinished"}}
	m.active = 0
	m.latestLog = "Downloading jq dependency"
	first := stripTerminal(m.View().Content)
	m.Update(m.spinner.Tick())
	second := stripTerminal(m.View().Content)
	for _, want := range []string{"Item 1 of 1", "0 completed", "Now: jq", "Downloading jq dependency"} {
		if !strings.Contains(first, want) {
			t.Fatalf("install view missing %q:\n%s", want, first)
		}
	}
	if first == second {
		t.Fatal("activity indicator did not move while an item was running")
	}
	if strings.Contains(first, "100%") {
		t.Fatal("active install claimed byte-level percentage")
	}
}

func TestMacCataloguePresetIDs(t *testing.T) {
	if len(macPackages) != 80 || len(macSettings) != 6 {
		t.Fatal("unexpected starter catalogue size")
	}
	for _, preset := range macPresets {
		m := newMacManifest()
		m.Mac.Packages = preset.IDs
		if err := m.Validate(); err != nil {
			t.Fatalf("%s: %v", preset.Name, err)
		}
	}
}
func TestMacEveryScreenFits(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	for _, p := range macPackages {
		m.selected[p.ID] = true
	}
	for _, s := range macSettings {
		m.selected[s.ID] = true
	}
	for _, size := range [][2]int{{80, 24}, {72, 20}, {120, 35}} {
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		for _, screen := range []macScreen{macScreenMenu, macScreenCategories, macScreenBrowse, macScreenPresets, macScreenReview, macScreenRestore, macScreenInstall, macScreenSummary, macScreenBootstrap, macScreenUpdates, macScreenAppConfigs, macScreenRaycast, macScreenTerminal, macScreenTerminalRestore, macScreenShell, macScreenUpdateConfirm} {
			m.screen = screen
			m.category = "settings"
			v := m.View().Content
			if lipgloss.Width(v) > size[0] || lipgloss.Height(v) > size[1] {
				t.Fatalf("%s at %v overflows: %dx%d", screen, size, lipgloss.Width(v), lipgloss.Height(v))
			}
		}
	}
}

func TestMacRealPreferenceFile(t *testing.T) {
	if runtime.GOOS != "darwin" || os.Getenv("MAGUS_TEST_MAC_DEFAULTS") != "1" {
		t.Skip("set MAGUS_TEST_MAC_DEFAULTS=1 outside the sandbox for isolated macOS defaults integration")
	}
	c := macTestContext(t)
	s := macSettings[0]
	s.Domain = filepath.Join(t.TempDir(), "magus-preference-fixture")
	step := preferenceStep{s}
	result := executeMacStep(c, step, false)
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if err := step.Remove(c); err != nil {
		t.Fatal(err)
	}
	after, err := readPreference(c, s)
	if err != nil || after.Kind != "absent" {
		t.Fatalf("original absence not restored: %+v %v", after, err)
	}
}
