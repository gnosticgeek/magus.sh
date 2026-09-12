package main

import (
	"strings"
	"testing"
	"time"
)

func TestFontsMenuSelectionAndReview(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	found := false
	for i, row := range m.rows() {
		if row.ID == "fonts" {
			m.cursor = i
			found = true
			break
		}
	}
	if !found {
		t.Fatal("Fonts missing from main menu")
	}
	press(m, "enter")
	if m.screen != macScreenBrowse || m.category != "fonts" || len(m.rows()) != 6 {
		t.Fatal("Fonts did not open six choices")
	}
	for _, row := range m.rows() {
		p, ok := macPackage(row.ID)
		if !ok || p.Kind != "cask" || p.AppBundle != "" || !strings.HasPrefix(p.ID, "font-") {
			t.Fatalf("invalid font: %s", row.ID)
		}
	}
	press(m, "ctrl+s")
	if len(m.selected) != 6 {
		t.Fatal("select all did not select six fonts")
	}
	press(m, "esc")
	if m.screen != macScreenMenu {
		t.Fatal("Fonts did not return to menu")
	}
	m.screen = macScreenReview
	if len(m.rows()) != 6 {
		t.Fatal("fonts missing from review")
	}
	if err := m.selectionManifest().Validate(); err != nil {
		t.Fatal(err)
	}
	m.screen, m.category, m.appGroup = macScreenBrowse, "apps", ""
	for _, row := range m.rows() {
		if strings.HasPrefix(row.ID, "font-") {
			t.Fatal("font leaked into apps")
		}
	}
}
