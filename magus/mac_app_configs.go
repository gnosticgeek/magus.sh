package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed dotfiles/zed/settings.json
var zedConfig string

//go:embed dotfiles/firefox/user.js
var firefoxConfig string

//go:embed dotfiles/raycast/README.md
var raycastGuide string

type appConfigPreset struct{ ID, Name, File, Body, Note string }

var appConfigPresets = []appConfigPreset{
	{"zed", "Zed", "settings.json", zedConfig, "Close Zed before applying. Replaces the settings file after backing it up. Formatting on save, 14pt text, system theme, telemetry off and build-folder exclusions. Export instead to merge with your own settings."},
	{"firefox", "Firefox", "user.js", firefoxConfig, "Launch Firefox once to create a profile, then quit it before applying. Only a single registered profile is selected automatically; export for multiple profiles. Restart to load the preset. Restoring user.js does NOT reset values already copied into prefs.js: reset the listed preferences in about:config if needed. Restores previous tabs on startup."},
}

func appConfig(id string) (appConfigPreset, bool) {
	for _, p := range appConfigPresets {
		if p.ID == id {
			return p, true
		}
	}
	return appConfigPreset{}, false
}

// Resolve only registered profiles. Never guess between profiles or follow a
// relative path outside Firefox's directory.
func firefoxConfigPath(p Paths) (string, error) {
	root := filepath.Join(p.Home, "Library", "Application Support", "Firefox")
	data, err := os.ReadFile(filepath.Join(root, "profiles.ini"))
	if err != nil {
		return "", fmt.Errorf("launch Firefox once, then quit it: %w", err)
	}
	var profiles []map[string]string
	var current map[string]string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			current = nil
			if strings.HasPrefix(line, "[Profile") {
				current = map[string]string{}
				profiles = append(profiles, current)
			}
		} else if current != nil && !strings.HasPrefix(line, ";") && !strings.HasPrefix(line, "#") {
			if k, v, ok := strings.Cut(line, "="); ok {
				current[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
	}
	var paths []string
	for _, profile := range profiles {
		path := profile["Path"]
		if path == "" {
			continue
		}
		if profile["IsRelative"] != "0" {
			clean := filepath.Clean(path)
			if filepath.IsAbs(path) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
				return "", fmt.Errorf("invalid Firefox profile path")
			}
			path = filepath.Join(root, clean)
		} else if !filepath.IsAbs(path) {
			return "", fmt.Errorf("Firefox absolute profile path is not absolute")
		}
		if !oneOf(path, paths) {
			paths = append(paths, path)
		}
	}
	if len(paths) != 1 {
		return "", fmt.Errorf("found %d Firefox profiles; export user.js and choose the profile through about:profiles", len(paths))
	}
	info, err := os.Stat(paths[0])
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("Firefox profile is not a directory")
	}
	return filepath.Join(paths[0], "user.js"), nil
}
func appConfigPath(p Paths, id string) (string, error) {
	switch id {
	case "zed":
		return filepath.Join(filepath.Dir(p.Config), "zed", "settings.json"), nil
	case "firefox":
		return firefoxConfigPath(p)
	default:
		return "", fmt.Errorf("unknown app configuration %q", id)
	}
}

type appConfigSnapshot struct {
	Path                       string
	Original, Applied, Pending []byte
	Restoring                  bool
	Existed                    bool
	Mode                       os.FileMode
}
type appConfigStep struct{ IDValue string }

func (s appConfigStep) ID() string { return "config:" + s.IDValue }
func (s appConfigStep) Describe() string {
	p, _ := appConfig(s.IDValue)
	return "Configure " + p.Name + " (back up existing settings)"
}
func (s appConfigStep) backup(p Paths) string {
	return filepath.Join(p.State, "app-config-"+s.IDValue+".json")
}
func (s appConfigStep) read(p Paths) (appConfigSnapshot, []byte, error) {
	path, err := appConfigPath(p, s.IDValue)
	if err != nil {
		return appConfigSnapshot{}, nil, err
	}
	if err = regularTerminalPath(path); err != nil {
		return appConfigSnapshot{}, nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return appConfigSnapshot{}, nil, err
	}
	current := appConfigSnapshot{Path: path, Original: data, Existed: err == nil, Mode: 0600}
	if info, e := os.Stat(path); e == nil {
		current.Mode = info.Mode().Perm()
	}
	if err = regularTerminalPath(s.backup(p)); err != nil {
		return current, nil, err
	}
	raw, err := os.ReadFile(s.backup(p))
	if os.IsNotExist(err) {
		return current, data, nil
	}
	if err != nil {
		return current, nil, err
	}
	// Decode into fresh storage. Decoding into current can reuse Original's
	// backing array, which aliases data and corrupts the comparison below.
	var snap appConfigSnapshot
	if err = json.Unmarshal(raw, &snap); err != nil {
		return snap, nil, err
	}
	if snap.Path != path {
		return snap, nil, fmt.Errorf("configuration path changed; restore the original profile first")
	}
	if !bytes.Equal(data, snap.Applied) && (snap.Pending == nil || !bytes.Equal(data, snap.Pending)) && !(snap.Restoring && bytes.Equal(data, snap.Original)) {
		return snap, nil, fmt.Errorf("%s settings changed outside Magus; export to preserve your edits", s.IDValue)
	}
	return snap, data, nil
}
func (s appConfigStep) Check(c *Context) (State, error) {
	p, ok := appConfig(s.IDValue)
	if !ok {
		return StateMissing, fmt.Errorf("unknown app configuration")
	}
	_, data, err := s.read(c.Paths)
	if err != nil {
		return StateMissing, err
	}
	if bytes.Equal(data, []byte(p.Body)) {
		return StateOK, nil
	}
	return StateMissing, nil
}
func (s appConfigStep) Apply(c *Context) error {
	p, ok := appConfig(s.IDValue)
	if !ok {
		return fmt.Errorf("unknown app configuration")
	}
	snap, data, err := s.read(c.Paths)
	if err != nil {
		return err
	}
	if c.DryRun {
		return nil
	}
	snap.Restoring = false
	snap.Applied = data
	snap.Pending = []byte(p.Body)
	raw, _ := json.MarshalIndent(snap, "", "  ")
	if err = writeFileAtomic(s.backup(c.Paths), raw, 0600); err != nil {
		return err
	}
	if err = writeFileAtomic(snap.Path, snap.Pending, snap.Mode); err != nil {
		return err
	}
	snap.Applied = snap.Pending
	snap.Pending = nil
	raw, _ = json.MarshalIndent(snap, "", "  ")
	return writeFileAtomic(s.backup(c.Paths), raw, 0600)
}
func (s appConfigStep) Remove(c *Context) error {
	if _, ok := appConfig(s.IDValue); !ok {
		return fmt.Errorf("unknown app configuration")
	}
	if _, err := os.Stat(s.backup(c.Paths)); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	snap, _, err := s.read(c.Paths)
	if err != nil {
		return err
	}
	if c.DryRun {
		return nil
	}
	snap.Restoring = true
	raw, _ := json.MarshalIndent(snap, "", "  ")
	if err = writeFileAtomic(s.backup(c.Paths), raw, 0600); err != nil {
		return err
	}
	if snap.Existed {
		err = writeFileAtomic(snap.Path, snap.Original, snap.Mode)
	} else {
		err = os.Remove(snap.Path)
		if os.IsNotExist(err) {
			err = nil
		}
	}
	if err != nil {
		return err
	}
	return os.Remove(s.backup(c.Paths))
}
func (m *macModel) appConfigRows() []macRow {
	rows := []macRow{
		{ID: "terminal", Name: "Ghostty", Summary: "Themes, readability, rendering and shell integration. Open terminal setup."},
		{ID: "shell", Name: "Modern CLI", Summary: "Choose shell integrations and aliases. Existing shell settings are preserved."},
	}
	for _, p := range appConfigPresets {
		check := "[ ] "
		if m.selected["config:"+p.ID] {
			check = "[x] "
		}
		if m.screen == "review" {
			check = ""
		}
		rows = append(rows, macRow{ID: "config:" + p.ID, Name: check + p.Name + " preset", Summary: p.Note, Note: "Enter toggles this preset in the review basket.\n\n```\n" + p.Body + "```"}, macRow{ID: "export:" + p.ID, Name: "Export " + p.Name + " preset", Summary: "Save a portable copy without changing the app.", Note: p.Note})
	}
	rows = append(rows, firefoxExtensionRows()...)
	return append(rows, macRow{ID: "raycast", Name: "Raycast guided setup", Summary: "Six settings recommendations and six extension links; configure them inside Raycast."}, macRow{ID: "review", Name: "Review & apply", Summary: "Review selected apps and configurations before making changes."}, macRow{ID: "restore", Name: "Restore Magus settings", Summary: "Restore recorded Mac, shell, Zed and Firefox files. Firefox preferences already loaded must be reset separately."})
}
func exportAppConfig(p Paths, id string) (string, error) {
	file, body := "README.md", raycastGuide
	if id != "raycast" {
		preset, ok := appConfig(id)
		if !ok {
			return "", fmt.Errorf("unknown preset")
		}
		file, body = preset.File, preset.Body
	}
	path := filepath.Join(p.Config, "exports", id, file)
	if err := regularTerminalPath(path); err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("export already exists: %s; move it before exporting again", path)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return path, writeFileAtomic(path, []byte(body), 0600)
}

func withoutAppConfig(ids []string, stepID string) []string {
	var out []string
	for _, id := range ids {
		if "config:"+id != stepID {
			out = append(out, id)
		}
	}
	return out
}
