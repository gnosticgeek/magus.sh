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

//go:embed dotfiles/ghostty/config.ghostty
var ghosttyTemplate string

var terminalThemes = []string{"Catppuccin", "TokyoNight", "Rose Pine"}

func terminalConfig(theme string) string {
	pair := map[string]string{"Catppuccin": "dark:Catppuccin Mocha,light:Catppuccin Latte", "TokyoNight": "dark:TokyoNight,light:TokyoNight Day", "Rose Pine": "dark:Rose Pine,light:Rose Pine Dawn"}[theme]
	return strings.Replace(ghosttyTemplate, "dark:Catppuccin Mocha,light:Catppuccin Latte", pair, 1)
}
func terminalID(theme string) string { return "terminal:" + theme }
func terminalPath(p Paths) string {
	return filepath.Join(filepath.Dir(p.Config), "ghostty", "config.ghostty")
}
func terminalSnapshotPath(p Paths) string { return filepath.Join(p.State, "ghostty-backup.json") }

type terminalSnapshot struct {
	Path     string
	Original []byte
	Existed  bool
	Mode     os.FileMode
	Applied  []byte
}

func readTerminalSnapshot(p Paths) (terminalSnapshot, error) {
	var snap terminalSnapshot
	b, err := os.ReadFile(terminalSnapshotPath(p))
	if err != nil {
		return snap, err
	}
	err = json.Unmarshal(b, &snap)
	return snap, err
}

// Reject symlinks, including parent directories: externally managed dotfiles
// must not be replaced or modified indirectly.
func regularTerminalPath(path string) error {
	for p := path; p != filepath.Dir(p); p = filepath.Dir(p) {
		info, err := os.Lstat(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s is a symlink; use the exported dotfile instead", p)
		}
		if p == path && !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", p)
		}
	}
	return nil
}

type terminalStep struct{ Theme string }

func (s terminalStep) ID() string       { return terminalID(s.Theme) }
func (s terminalStep) Describe() string { return "Configure Ghostty / " + s.Theme }
func (s terminalStep) Check(c *Context) (State, error) {
	if !oneOf(s.Theme, terminalThemes) {
		return StateMissing, fmt.Errorf("unknown terminal theme %q", s.Theme)
	}
	path := terminalPath(c.Paths)
	if err := regularTerminalPath(path); err != nil {
		return StateMissing, err
	}
	// These files load after the generated file and can silently override it.
	for _, other := range []string{filepath.Join(filepath.Dir(path), "config"), filepath.Join(c.Paths.Home, "Library", "Application Support", "com.mitchellh.ghostty", "config"), filepath.Join(c.Paths.Home, "Library", "Application Support", "com.mitchellh.ghostty", "config.ghostty")} {
		b, err := os.ReadFile(other)
		if err != nil && !os.IsNotExist(err) {
			return StateMissing, err
		}
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				return StateMissing, fmt.Errorf("existing settings in %s may override this profile; export the dotfile or consolidate your settings first", other)
			}
		}
	}
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return StateMissing, err
	}
	snap, snapErr := readTerminalSnapshot(c.Paths)
	if snapErr != nil && !os.IsNotExist(snapErr) {
		return StateMissing, snapErr
	}
	if snapErr == nil && (snap.Path != path || !bytes.Equal(b, snap.Applied)) {
		return StateMissing, fmt.Errorf("Ghostty settings changed since Magus last wrote them; export the profile to preserve your edits")
	}
	if bytes.Equal(b, []byte(terminalConfig(s.Theme))) {
		return StateOK, nil
	}
	return StateMissing, nil
}
func (s terminalStep) Apply(c *Context) error {
	if c.DryRun {
		return nil
	}
	if _, err := s.Check(c); err != nil {
		return err
	}
	// Check the application even if its install was skipped after an error.
	binary := filepath.Join("/Applications", "Ghostty.app", "Contents", "MacOS", "ghostty")
	if _, err := os.Stat(binary); err != nil {
		binary = filepath.Join(c.Paths.Home, "Applications", "Ghostty.app", "Contents", "MacOS", "ghostty")
	}
	if err := os.MkdirAll(c.Paths.State, 0700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(c.Paths.State, "ghostty-validate-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	config := []byte(terminalConfig(s.Theme))
	_, err = temp.Write(config)
	closeErr := temp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	output, err := c.macCommand(binary, "+validate-config", "--config-file="+temp.Name())
	if err != nil {
		return fmt.Errorf("Ghostty validation failed: %w", err)
	}
	if strings.TrimSpace(output) != "" {
		return fmt.Errorf("Ghostty validation: %s", output)
	}
	return s.write(c)
}
func (s terminalStep) write(c *Context) error {
	if c.DryRun {
		return nil
	}
	if _, err := s.Check(c); err != nil {
		return err
	}
	path := terminalPath(c.Paths)
	snap, err := readTerminalSnapshot(c.Paths)
	if os.IsNotExist(err) {
		snap = terminalSnapshot{Path: path, Mode: 0600}
		if info, statErr := os.Stat(path); statErr == nil {
			snap.Existed, snap.Mode = true, info.Mode().Perm()
			snap.Original, err = os.ReadFile(path)
			if err != nil {
				return err
			}
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
	} else if err != nil {
		return err
	}
	config := []byte(terminalConfig(s.Theme))
	// Persist a backup before replacing the file. On failure it is retained.
	current, _ := os.ReadFile(path)
	snap.Applied = current
	b, _ := json.MarshalIndent(snap, "", "  ")
	if err := writeFileAtomic(terminalSnapshotPath(c.Paths), b, 0600); err != nil {
		return err
	}
	if err := writeFileAtomic(path, config, snap.Mode); err != nil {
		return err
	}
	snap.Applied = config
	b, _ = json.MarshalIndent(snap, "", "  ")
	if err := writeFileAtomic(terminalSnapshotPath(c.Paths), b, 0600); err != nil {
		return err
	}
	return nil
}
func (s terminalStep) Remove(c *Context) error {
	snap, err := readTerminalSnapshot(c.Paths)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if snap.Path != terminalPath(c.Paths) {
		return fmt.Errorf("Ghostty configuration path has changed")
	}
	if err := regularTerminalPath(snap.Path); err != nil {
		return err
	}
	b, err := os.ReadFile(snap.Path)
	if err != nil {
		return err
	}
	if !bytes.Equal(b, snap.Applied) {
		return fmt.Errorf("Ghostty settings were edited; leaving them unchanged")
	}
	if c.DryRun {
		return nil
	}
	if snap.Existed {
		err = writeFileAtomic(snap.Path, snap.Original, snap.Mode)
	} else {
		err = os.Remove(snap.Path)
	}
	if err != nil {
		return err
	}
	return os.Remove(terminalSnapshotPath(c.Paths))
}

func (m *macModel) terminalRows() []macRow {
	rows := []macRow{}
	for _, theme := range terminalThemes {
		rows = append(rows, macRow{ID: terminalID(theme), Name: "Ghostty / " + theme, Summary: "Install Ghostty and JetBrains Mono. Apply a light/dark theme, 14pt text, comfortable padding and clearly visible blurred transparency.", Source: terminalPath(m.paths), Note: "Enter adds this setup to Review & install. Existing config.ghostty is backed up before replacement; edited Magus files and conflicting configs are left alone.\n\n" + "```\n" + terminalConfig(theme) + "```"})
	}
	rows = append(rows, macRow{ID: "terminal-review", Name: "Review & apply setup", Summary: "Review your basket, then press Enter to install and apply the selected Ghostty theme."})
	rows = append(rows, macRow{ID: "terminal-export", Name: "Export selected Ghostty dotfile", Summary: "Save the selected theme (Catppuccin by default) under the Magus config directory without changing Ghostty."})
	rows = append(rows, macRow{ID: "terminal-restore", Name: "Restore previous Ghostty settings", Summary: "Restore the saved configuration without removing Ghostty or fonts. Enter opens confirmation."})
	return append(rows, macRow{ID: "shell", Name: "Modern commands", Summary: "Choose which modern tools to use in your Zsh terminal, or undo the setup."})
}
