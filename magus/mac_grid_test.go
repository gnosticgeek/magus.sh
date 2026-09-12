package main

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func TestGridNavigationAndSizes(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category = macScreenBrowse, "tools"
	for _, size := range [][2]int{{72, 20}, {80, 24}, {120, 35}, {180, 35}} {
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		view := m.View().Content
		if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
			t.Fatal("grid overflow", size)
		}
		cols, _, _ := m.browserLayout()
		if size[0] >= 80 && cols < 2 {
			t.Fatal("missing columns")
		}
	}
	m.width = 80
	m.cursor = 0
	m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if m.cursor != 1 {
		t.Fatal("right did not move")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.cursor != 3 {
		t.Fatal("down did not retain column")
	}
}

func TestSelectAllResultsPreservesOtherPicks(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category = macScreenBrowse, "tools"
	m.selected["firefox"] = true
	m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if !m.selected["mole"] || !m.selected["git"] {
		t.Fatal("tools not selected")
	}
	m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if m.selected["git"] || !m.selected["firefox"] {
		t.Fatal("wrong deselection scope")
	}
	m.searching = true
	m.search.SetValue("mole")
	m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if !m.selected["mole"] || m.selected["git"] || m.search.Value() != "mole" {
		t.Fatal("search select-all scope")
	}
	m.searching = false
	m.category = "settings"
	m.Update(tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	if m.selected["restore"] {
		t.Fatal("restore selected by bulk action")
	}
}
