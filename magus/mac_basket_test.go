package main

import (
	"testing"
	"time"
)

func TestInstalledSelectionsPrunedAndCannotBeReadded(t *testing.T) {
	c := macTestContext(t)
	manifest := newMacManifest()
	manifest.Mac.Packages = []string{"brave-browser", "ghostty", "zed", "jq"}
	m := newMacModel(c.Paths, "", manifest, true, time.Second)
	m.Update(macInventory{states: map[string]string{"brave-browser": "installed", "ghostty": "installed", "zed": "installed", "git": "outside Homebrew", "jq": "not installed"}})
	if len(m.selected) != 1 || !m.selected["jq"] {
		t.Fatal(m.selected)
	}
	m.toggle("brave-browser")
	if m.selected["brave-browser"] {
		t.Fatal("installed app selected")
	}
	m.screen, m.category = "browse", "tools"
	m.selectAllResults()
	if m.selected["git"] {
		t.Fatal("external app bulk selected")
	}
	m.screen = "presets"
	m.cursor = 1
	press(m, "enter")
	if m.selected["ghostty"] || m.selected["git"] {
		t.Fatal("preset selected existing package")
	}
	m.screen = "review"
	for _, r := range m.rows() {
		if !m.needsSelection(r.ID) {
			t.Fatal("installed app in review")
		}
	}
}

func TestEnterSelectsAppAndTabOpensDetails(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category, m.appGroup = "browse", "apps", "ai"
	for i, row := range m.rows() {
		if row.ID == "lm-studio" {
			m.cursor = i
		}
	}
	press(m, "enter")
	if !m.selected["lm-studio"] || m.details || m.screen != "browse" {
		t.Fatal("Enter should select LM Studio without opening details or installing")
	}
	if m.notice != "Selected LM Studio" {
		t.Fatal(m.notice)
	}
	press(m, "tab")
	if !m.details {
		t.Fatal("Tab should open details")
	}
	press(m, "esc")
	press(m, "space")
	if m.selected["lm-studio"] {
		t.Fatal("Space should remove selection")
	}
	m.inventory.states["lm-studio"] = "installed"
	press(m, "enter")
	if m.selected["lm-studio"] {
		t.Fatal("Enter should not select installed apps")
	}
}
