package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func makeFirefoxProfile(t *testing.T, c *Context) string {
	t.Helper()
	root := filepath.Join(c.Paths.Home, "Library", "Application Support", "Firefox")
	path := filepath.Join(root, "Profiles", "test.default")
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "profiles.ini"), []byte("[Profile0]\nIsRelative=1\nPath=Profiles/test.default\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(path, "user.js")
}
func TestAppConfigBackupRestoreAndEdits(t *testing.T) {
	for _, id := range []string{"zed", "firefox"} {
		t.Run(id, func(t *testing.T) {
			c := terminalTestContext(t)
			if id == "firefox" {
				makeFirefoxProfile(t, c)
			}
			s := appConfigStep{IDValue: id}
			path, err := appConfigPath(c.Paths, id)
			if err != nil {
				t.Fatal(err)
			}
			original := []byte("// my existing settings\n")
			if err := writeFileAtomic(path, original, 0640); err != nil {
				t.Fatal(err)
			}
			if err := s.Apply(c); err != nil {
				t.Fatal(err)
			}
			if state, err := s.Check(c); err != nil || state != StateOK {
				t.Fatal(state, err)
			}
			if err := s.Apply(c); err != nil {
				t.Fatal(err)
			}
			data, _ := os.ReadFile(path)
			os.WriteFile(path, append(data, []byte("// user edit")...), 0640)
			if err := s.Apply(c); err == nil {
				t.Fatal("overwrote edits")
			}
			if err := s.Remove(c); err == nil {
				t.Fatal("restored over edits")
			}
			os.WriteFile(path, data, 0640)
			if err := s.Remove(c); err != nil {
				t.Fatal(err)
			}
			restored, _ := os.ReadFile(path)
			if !bytes.Equal(restored, original) {
				t.Fatal("lost original")
			}
			info, _ := os.Stat(path)
			if info.Mode().Perm() != 0640 {
				t.Fatal("lost mode")
			}
			os.Remove(path)
			c.DryRun = true
			if err := s.Apply(c); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatal("dry run wrote file")
			}
			c.DryRun = false
			if err := s.Apply(c); err != nil {
				t.Fatal(err)
			}
			if err := s.Remove(c); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatal("did not restore absence")
			}
		})
	}
}
func TestAppConfigPendingWriteAndSymlink(t *testing.T) {
	c := terminalTestContext(t)
	s := appConfigStep{IDValue: "zed"}
	path, _ := appConfigPath(c.Paths, "zed")
	original := []byte("original")
	writeFileAtomic(path, original, 0600)
	snap := appConfigSnapshot{Path: path, Original: original, Existed: true, Mode: 0600, Applied: original, Pending: []byte(zedConfig)}
	raw, _ := json.Marshal(snap)
	writeFileAtomic(s.backup(c.Paths), raw, 0600)
	writeFileAtomic(path, []byte(zedConfig), 0600)
	if err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove(c); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !bytes.Equal(data, original) {
		t.Fatal("pending recovery lost backup")
	}
	os.Remove(path)
	target := filepath.Join(c.Paths.Home, "target")
	os.WriteFile(target, original, 0600)
	os.Symlink(target, path)
	if err := s.Apply(c); err == nil {
		t.Fatal("followed symlink")
	}
}
func TestFirefoxProfileAmbiguityAndTraversal(t *testing.T) {
	c := terminalTestContext(t)
	path := makeFirefoxProfile(t, c)
	got, err := firefoxConfigPath(c.Paths)
	if err != nil || got != path {
		t.Fatal(got, err)
	}
	ini := filepath.Join(c.Paths.Home, "Library", "Application Support", "Firefox", "profiles.ini")
	for _, data := range []string{
		"[Profile0]\nPath=Profiles/test.default\n[Profile1]\nPath=Profiles/another\n",
		"[Profile0]\nPath=../../outside\n",
		"[Profile0]\nIsRelative=0\nPath=relative\n",
	} {
		os.WriteFile(ini, []byte(data), 0600)
		if _, err := firefoxConfigPath(c.Paths); err == nil {
			t.Fatal("accepted", data)
		}
	}
}
func TestAppConfigMenuManifestExportAndRestorePlan(t *testing.T) {
	c := terminalTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen = macScreenAppConfigs
	for i, row := range m.rows() {
		if row.ID == "config:zed" {
			m.cursor = i
			break
		}
	}
	press(m, "enter")
	manifest := m.selectionManifest()
	if !oneOf("zed", manifest.Mac.AppConfigs) || !m.selected["zed"] {
		t.Fatal(manifest)
	}
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(c.Paths.Config, "test.toml")
	if err := manifest.Save(path); err != nil {
		t.Fatal(err)
	}
	saved, err := LoadManifest(path)
	if err != nil || !oneOf("zed", saved.Mac.AppConfigs) {
		t.Fatal(saved, err)
	}
	saved.Mac.Packages = nil
	steps := macSteps(saved)
	if len(steps) != 2 || steps[0].ID() != "package:zed" || steps[1].ID() != "config:zed" {
		t.Fatalf("wrong dependency order: %+v", steps)
	}
	saved.Mac.AppConfigs = []string{"unknown"}
	if saved.Validate() == nil {
		t.Fatal("accepted unknown")
	}
	saved.Mac.AppConfigs = []string{"zed", "zed"}
	if saved.Validate() == nil {
		t.Fatal("accepted duplicate")
	}
	for _, id := range []string{"zed", "firefox", "raycast"} {
		path, err := exportAppConfig(c.Paths, id)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := exportAppConfig(c.Paths, id); err == nil {
			t.Fatal("overwrote export", path)
		}
	}
	if err := (appConfigStep{IDValue: "zed"}).Apply(c); err != nil {
		t.Fatal(err)
	}
	steps, err = restoreSteps(c)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, step := range steps {
		if step.ID() == "config:zed" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing restore")
	}
	if !strings.Contains(raycastGuide, "Hyper Key") || len(firefoxExtensionRows()) != 5 {
		t.Fatal("missing guided setup")
	}
}

func TestAppConfigRestoreRecoveryAndDependencyReview(t *testing.T) {
	c := terminalTestContext(t)
	s := appConfigStep{IDValue: "zed"}
	path, _ := appConfigPath(c.Paths, "zed")
	original := []byte("// original")
	writeFileAtomic(path, original, 0600)
	if err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	snap, _, err := s.read(c.Paths)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a successful restore followed by failure to remove the backup.
	snap.Restoring = true
	raw, _ := json.Marshal(snap)
	writeFileAtomic(s.backup(c.Paths), raw, 0600)
	writeFileAtomic(path, original, 0600)
	if err := s.Remove(c); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.backup(c.Paths)); !os.IsNotExist(err) {
		t.Fatal("left completed snapshot")
	}
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.selected["config:zed"], m.selected["zed"] = true, true
	m.toggle("zed")
	if !m.selected["zed"] {
		t.Fatal("hid required app")
	}
	m.toggle("config:zed")
	m.toggle("zed")
	if m.selected["zed"] {
		t.Fatal("could not remove app")
	}
}
