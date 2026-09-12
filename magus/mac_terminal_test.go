package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
)

func terminalTestContext(t *testing.T) *Context {
	c := macTestContext(t)
	home, err := filepath.EvalSymlinks(c.Paths.Home)
	if err != nil {
		t.Fatal(err)
	}
	c.Paths = newPathsUnder(home)
	return c
}
func TestTerminalMenuSelectionAndPersistence(t *testing.T) {
	c := terminalTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.cursor = 8
	press(m, "enter")
	if m.screen != macScreenTerminal {
		t.Fatal(m.screen)
	}
	press(m, "enter")
	if !m.selected["ghostty"] || !m.selected["font-jetbrains-mono"] || m.selectionManifest().Mac.Terminal != "Catppuccin" {
		t.Fatal(m.selected)
	}
	m.cursor = 1
	press(m, "enter")
	if m.selected[terminalID("Catppuccin")] || m.selectionManifest().Mac.Terminal != "TokyoNight" {
		t.Fatal(m.selected)
	}
	manifest := m.selectionManifest()
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(c.Paths.Config, "terminal-test.toml")
	if err := manifest.Save(path); err != nil {
		t.Fatal(err)
	}
	saved, err := LoadManifest(path)
	if err != nil || saved.Mac.Terminal != "TokyoNight" {
		t.Fatal(saved, err)
	}
	steps := macSteps(saved)
	if steps[len(steps)-1].ID() != terminalID("TokyoNight") {
		t.Fatal("profile must run after packages")
	}
	m.screen = macScreenReview
	found := false
	for _, r := range m.rows() {
		if r.ID == terminalID("TokyoNight") {
			found = true
		}
	}
	if !found {
		t.Fatal("profile missing from review")
	}
	for _, size := range [][2]int{{80, 24}, {72, 20}, {120, 35}} {
		m.width, m.height = size[0], size[1]
		for _, screen := range []string{"menu", "terminal", "review"} {
			m.screen, m.cursor = macScreen(screen), 0
			view := m.viewContent()
			if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
				t.Fatal("overflow", size, screen)
			}
			if screen == "menu" && size[0] == 80 && !strings.Contains(view, "App setups") {
				t.Fatal("app setups menu item hidden")
			}
		}
	}
}
func TestTerminalBackupSwitchRestoreAndEdits(t *testing.T) {
	c := terminalTestContext(t)
	path := terminalPath(c.Paths)
	original := []byte("font-size = 18\n")
	if err := writeFileAtomic(path, original, 0640); err != nil {
		t.Fatal(err)
	}
	s := terminalStep{"Catppuccin"}
	if err := s.write(c); err != nil {
		t.Fatal(err)
	}
	if state, err := s.Check(c); err != nil || state != StateOK {
		t.Fatal(state, err)
	}
	s.Theme = "Rose Pine"
	if err := s.write(c); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove(c); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	info, _ := os.Stat(path)
	if string(b) != string(original) || info.Mode().Perm() != 0640 {
		t.Fatal("original not restored")
	}
	if err := s.write(c); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(path, []byte("font-size = 22\n"), 0600)
	if _, err := s.Check(c); err == nil {
		t.Fatal("external edits accepted")
	}
	if err := s.Remove(c); err == nil {
		t.Fatal("external edits overwritten")
	}
}
func TestTerminalDryRunSymlinkAndConflicts(t *testing.T) {
	c := terminalTestContext(t)
	s := terminalStep{"Catppuccin"}
	c.DryRun = true
	if r := executeMacStep(c, s, false); r.Err != nil {
		t.Fatal(r.Err)
	}
	if _, err := os.Stat(terminalPath(c.Paths)); !os.IsNotExist(err) {
		t.Fatal("preview wrote config")
	}
	if _, err := os.Stat(terminalSnapshotPath(c.Paths)); !os.IsNotExist(err) {
		t.Fatal("preview wrote backup")
	}
	c.DryRun = false
	other := filepath.Join(c.Paths.Home, "external")
	os.WriteFile(other, []byte("font-size = 20\n"), 0600)
	os.MkdirAll(filepath.Dir(terminalPath(c.Paths)), 0755)
	os.Symlink(other, terminalPath(c.Paths))
	if err := s.write(c); err == nil {
		t.Fatal("replaced symlink")
	}
	os.Remove(terminalPath(c.Paths))
	conflict := filepath.Join(c.Paths.Home, "Library", "Application Support", "com.mitchellh.ghostty", "config")
	writeFileAtomic(conflict, []byte("theme = TokyoNight\n"), 0600)
	if _, err := s.Check(c); err == nil {
		t.Fatal("ignored overriding config")
	}
}

func TestTerminalApplyValidatesBeforeWriting(t *testing.T) {
	c := terminalTestContext(t)
	s := terminalStep{"Catppuccin"}
	c.Execute = func(_ context.Context, name string, args ...string) (string, error) {
		if filepath.Base(name) != "ghostty" || args[0] != "+validate-config" {
			t.Fatal(name, args)
		}
		config, err := os.ReadFile(strings.TrimPrefix(args[1], "--config-file="))
		if err != nil || string(config) != terminalConfig(s.Theme) {
			t.Fatal("wrong validator input", err)
		}
		return "", fmt.Errorf("invalid config")
	}
	if err := s.Apply(c); err == nil {
		t.Fatal("ignored validator failure")
	}
	if _, err := os.Stat(terminalPath(c.Paths)); !os.IsNotExist(err) {
		t.Fatal("wrote invalid config")
	}
	c.Execute = func(context.Context, string, ...string) (string, error) { return "", nil }
	if err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	if state, err := s.Check(c); state != StateOK || err != nil {
		t.Fatal(state, err)
	}
	if err := s.Remove(c); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(terminalPath(c.Paths)); !os.IsNotExist(err) {
		t.Fatal("did not restore original absence")
	}
}

func TestTerminalReviewAction(t *testing.T) {
	c := terminalTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.cursor = macScreenTerminal, 2
	press(m, "enter")
	m.cursor = 3
	press(m, "enter")
	if m.screen != macScreenReview || m.selectionManifest().Mac.Terminal != "Rose Pine" {
		t.Fatal(m.screen, m.selected)
	}
}
