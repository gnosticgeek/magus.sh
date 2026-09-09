package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestUpdateCommandSequenceAndFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "refresh failure"}[fail], func(t *testing.T) {
			dir := t.TempDir()
			script := filepath.Join(dir, "brew")
			log := filepath.Join(dir, "calls")
			t.Setenv("MAGUS_TEST_CALLS", log)
			content := "#!/bin/sh\necho \"$1:$HOMEBREW_NO_AUTO_UPDATE:$HOMEBREW_NO_INSTALL_CLEANUP\" >> \"$MAGUS_TEST_CALLS\"\n"
			if fail {
				content += "exit 1\n"
			}
			if err := os.WriteFile(script, []byte(content), 0700); err != nil {
				t.Fatal(err)
			}
			cmd := &macUpdateCommand{brew: script, timeout: time.Second, out: io.Discard, errOut: io.Discard, refresh: true}
			err := cmd.Run()
			if (err != nil) != fail {
				t.Fatalf("unexpected error: %v", err)
			}
			got, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			want := "update:1:1\n"
			if string(got) != want {
				t.Fatalf("commands: %q", got)
			}
		})
	}
}

func TestUpdatePreviewDoesNotStartCommand(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen = "updates"
	if cmd := m.updateAll(); cmd != nil {
		t.Fatal("preview started updates")
	}
	if !strings.Contains(m.notice, "no updates") {
		t.Fatal(m.notice)
	}
	if _, err := os.Stat(filepath.Join(c.Paths.State, "install.lock")); !os.IsNotExist(err) {
		t.Fatal("preview created state")
	}
}

func TestInstalledBadgeIsIndependentOfBasket(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category, m.appGroup = "browse", "apps", "browsers"
	m.inventory.states["firefox"] = "installed"
	view := stripTerminal(m.View().Content)
	if strings.Contains(view, "[ ] Firefox") || !strings.Contains(view, "installed") {
		t.Fatal(view)
	}
	m.selected["firefox"] = true
	view = stripTerminal(m.View().Content)
	if strings.Contains(view, "[x] Firefox") || !strings.Contains(view, "installed") {
		t.Fatal(view)
	}
	for _, st := range []string{"not installed", "inspection failed", "needs Homebrew"} {
		if installedBadge(st) != "" {
			t.Fatalf("false installed badge for %s", st)
		}
	}
	if !strings.Contains(stripTerminal(installedBadge("outside Homebrew")), "external") {
		t.Fatal("unmanaged status missing")
	}
}
