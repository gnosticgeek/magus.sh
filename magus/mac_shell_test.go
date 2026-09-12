package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
)

func shellTestContext(t *testing.T) *Context {
	t.Helper()
	t.Setenv("ZDOTDIR", "")
	return terminalTestContext(t)
}
func TestModernShellLifecycle(t *testing.T) {
	c := shellTestContext(t)
	path, _ := shellPath(c.Paths)
	original := "# my settings without trailing newline"
	if err := os.WriteFile(path, []byte(original), 0640); err != nil {
		t.Fatal(err)
	}
	s := shellStep{modernShell{true, []string{"ls", "cd", "fzf", "atuin", "delta"}}}
	c.DryRun = true
	if err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(shellBackup(c.Paths)); !os.IsNotExist(err) {
		t.Fatal("dry run wrote backup")
	}
	data, _ := os.ReadFile(path)
	if string(data) != original {
		t.Fatal("dry run changed shell")
	}
	c.DryRun = false
	for i := 0; i < 2; i++ {
		if err := s.Apply(c); err != nil {
			t.Fatal(err)
		}
	}
	data, _ = os.ReadFile(path)
	if strings.Count(string(data), shellStart) != 1 {
		t.Fatal("duplicate block")
	}
	// Unrelated edits survive switching options and undo.
	extra := "\n# my new setting\n"
	os.WriteFile(path, append(data, []byte(extra)...), 0640)
	s.Settings.Commands = []string{"tree", "top"}
	if err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	if state, err := s.Check(c); state != StateOK || err != nil {
		t.Fatal(state, err)
	}
	s.Settings.Enabled = false
	if err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	info, _ := os.Stat(path)
	if string(data) != original+extra || info.Mode().Perm() != 0640 {
		t.Fatal("lost user content or permissions", string(data))
	}
	if state, err := s.Check(c); state != StateOK || err != nil {
		t.Fatal(state, err)
	}
	if err := s.Remove(c); err != nil {
		t.Fatal("undo not idempotent", err)
	}
}
func TestModernShellProtectsEditsAndSymlinks(t *testing.T) {
	c := shellTestContext(t)
	s := shellStep{modernShell{true, []string{"ls"}}}
	path, _ := shellPath(c.Paths)
	if err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	edited := strings.Replace(string(data), "alias ls='eza --group-directories-first'", "alias ls='eza -l'", 1)
	os.WriteFile(path, []byte(edited), 0600)
	if err := s.Apply(c); err == nil {
		t.Fatal("overwrote edited block")
	}
	if err := s.Remove(c); err == nil {
		t.Fatal("removed edited block")
	}
	os.WriteFile(path, data, 0600)
	if err := s.Remove(c); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("did not restore original absence")
	}
	other := filepath.Join(c.Paths.Home, "other")
	os.WriteFile(other, []byte("untouched"), 0600)
	os.Symlink(other, path)
	if err := s.Apply(c); err == nil {
		t.Fatal("followed symlink")
	}
}
func TestModernShellPendingWriteRecovery(t *testing.T) {
	c := shellTestContext(t)
	first := shellStep{modernShell{true, []string{"ls"}}}
	if err := first.Apply(c); err != nil {
		t.Fatal(err)
	}
	path, data, snap, _, err := readShell(c.Paths)
	if err != nil {
		t.Fatal(err)
	}
	next := shellStep{modernShell{true, []string{"top"}}}
	snap.Pending = next.Settings.block()
	raw, _ := json.Marshal(snap)
	if err := os.WriteFile(shellBackup(c.Paths), raw, 0600); err != nil {
		t.Fatal(err)
	}
	// Simulate crash after shell write but before final backup update.
	os.WriteFile(path, []byte(strings.Replace(string(data), snap.Applied, snap.Pending, 1)), 0600)
	if err := next.Remove(c); err != nil {
		t.Fatal(err)
	}
}
func TestModernShellZshIsolation(t *testing.T) {
	c := shellTestContext(t)
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available")
	}
	bin := filepath.Join(c.Paths.Home, "bin")
	os.MkdirAll(bin, 0700)
	for _, name := range []string{"eza", "btop", "delta", "zoxide", "fzf", "atuin"} {
		script := "#!/bin/sh\nexit 0\n"
		if name == "zoxide" {
			script = "#!/bin/sh\nprintf '%s\\n' 'cd() { builtin cd \"$@\"; }'\n"
		}
		os.WriteFile(filepath.Join(bin, name), []byte(script), 0700)
	}
	settings := modernShell{Enabled: true}
	for _, o := range shellOptions {
		settings.Commands = append(settings.Commands, o.ID)
	}
	file := filepath.Join(c.Paths.Home, "generated.zsh")
	os.WriteFile(file, []byte(settings.block()), 0600)
	syntax := exec.Command(zsh, "-n", file)
	if output, err := syntax.CombinedOutput(); err != nil {
		t.Fatal(string(output), err)
	}
	for _, interactive := range []bool{false, true} {
		flags := "-fc"
		expected := "no"
		if interactive {
			flags = "-fic"
			expected = "yes"
		}
		cmd := exec.Command(zsh, flags, `source "$1"; if (( $+aliases[ls] )); then print yes; else print no; fi; print "${GIT_PAGER:-none}"`, "test", file)
		cmd.Env = []string{"HOME=" + c.Paths.Home, "PATH=" + bin + ":/usr/bin:/bin", "TERM=dumb"}
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(string(output), err)
		}
		pager := "none"
		if interactive {
			pager = "delta"
		}
		if strings.TrimSpace(string(output)) != expected+"\n"+pager {
			t.Fatal("wrong shell isolation", string(output))
		}
	}
}
func TestModernShellMenuManifestAndPlan(t *testing.T) {
	c := shellTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	if m.selectionManifest().Mac.ModernShell != nil {
		t.Fatal("enabled by default")
	}
	m.screen = macScreenShell
	press(m, "enter")
	if !m.shellSettings.Enabled || !m.selected["eza"] || !m.selected["shell:configure"] {
		t.Fatal(m.selected)
	}
	manifest := m.selectionManifest()
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(c.Paths.Config, "shell.toml")
	if err := manifest.Save(path); err != nil {
		t.Fatal(err)
	}
	saved, err := LoadManifest(path)
	if err != nil || saved.Mac.ModernShell == nil || !saved.Mac.ModernShell.Enabled {
		t.Fatal(saved, err)
	}
	// Headless manifests receive the same required packages.
	saved.Mac.Packages = nil
	steps := macSteps(saved)
	if len(steps) < 2 || steps[len(steps)-1].ID() != "shell:configure" {
		t.Fatal("missing ordered steps")
	}
	for _, size := range [][2]int{{72, 20}, {80, 24}, {120, 35}} {
		m.width, m.height = size[0], size[1]
		for _, screen := range []string{"shell", "review"} {
			m.screen = macScreen(screen)
			view := m.viewContent()
			if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
				t.Fatal("overflow", size, screen)
			}
		}
	}
	m.screen = macScreenShell
	m.cursor = len(m.shellRows()) - 1
	press(m, "enter")
	if m.screen != macScreenReview || m.selectionManifest().Mac.ModernShell.Enabled {
		t.Fatal("undo not queued")
	}
	saved.Mac.ModernShell.Commands = []string{"grep"}
	if saved.Validate() == nil {
		t.Fatal("unsafe option accepted")
	}
}

func TestModernShellDependencyBasketAndRestore(t *testing.T) {
	c := shellTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.selected["fzf"] = true // A manual pick must survive disabling the shell.
	m.screen = macScreenShell
	press(m, "enter")
	m.screen = macScreenReview
	m.toggle("eza")
	if !m.selected["eza"] {
		t.Fatal("required dependency removed")
	}
	m.screen = macScreenShell
	m.cursor = 0
	press(m, "enter")
	if m.selected["eza"] || !m.selected["fzf"] {
		t.Fatal("wrong dependency cleanup", m.selected)
	}
	s := shellStep{modernShell{true, []string{"ls"}}}
	if err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	steps, err := restoreSteps(c)
	if err != nil || len(steps) != 1 || steps[0].ID() != s.ID() {
		t.Fatal(steps, err)
	}
	if result := executeMacStep(c, steps[0], true); result.Err != nil {
		t.Fatal(result.Err)
	}
}
func TestModernShellCustomDirectoryAndMissingTools(t *testing.T) {
	c := shellTestContext(t)
	dir := filepath.Join(c.Paths.Home, "custom shell")
	t.Setenv("ZDOTDIR", dir)
	s := shellStep{modernShell{true, []string{"ls"}}}
	if err := s.Apply(c); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".zshrc")); err != nil {
		t.Fatal(err)
	}
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh unavailable")
	}
	cmd := exec.Command(zsh, "-fic", `source "$1"; (( $+aliases[ls] == 0 ))`, "test", filepath.Join(dir, ".zshrc"))
	cmd.Env = []string{"PATH=/nonexistent", "HOME=" + c.Paths.Home, "TERM=dumb"}
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatal("missing tools broke shell", string(output), err)
	}
	t.Setenv("ZDOTDIR", "relative")
	if err := s.Remove(c); err == nil {
		t.Fatal("accepted relative ZDOTDIR")
	}
}
