package main

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func TestSearchShowsScopeCountsHighlightsAndRestoresLocation(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category, m.appGroup = macScreenBrowse, "apps", "development"
	m.catalogueFilter = macCatalogueAvailable
	m.cursor = 2
	origin := m.location()

	press(m, "/")
	for _, r := range "vsc" {
		press(m, string(r))
	}
	view := m.View().Content
	plain := ansi.Strip(view)
	for _, want := range []string{"Developer tools", "matches ·", "Visual Studio Code"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("search view missing %q:\n%s", want, plain)
		}
	}
	if highlighted := highlightFuzzyMatch("Visual Studio Code", "vsc"); ansi.Strip(highlighted) != "Visual Studio Code" || !strings.Contains(highlighted, "\x1b[") {
		t.Fatalf("fuzzy highlighting lost text or styling: %q", highlighted)
	}

	press(m, "esc")
	if m.location() != origin {
		t.Fatalf("search restored %+v, want %+v", m.location(), origin)
	}
}

func TestSearchEmptyStateOffersRecovery(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category = macScreenBrowse, "tools"
	press(m, "/")
	for _, r := range "no-such-magus-item" {
		press(m, string(r))
	}
	plain := ansi.Strip(m.View().Content)
	if !strings.Contains(plain, "No matches") || !strings.Contains(plain, "press Esc") {
		t.Fatalf("empty search is not actionable:\n%s", plain)
	}
}

func TestBasketSummaryDrawerAndReviewBoundary(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	for _, id := range []string{"firefox", "jq", "font-inter", "finder-hidden", "config:zed"} {
		m.selected[id] = true
	}
	if got := m.basketSummary(); got != "Basket 5 · 1 app · 1 tool · 1 font · 1 setting · 1 setup" {
		t.Fatal(got)
	}
	m.screen, m.category = macScreenBrowse, "tools"
	press(m, "v")
	if m.screen != macScreenBasket || !strings.Contains(ansi.Strip(m.View().Content), m.basketSummary()) {
		t.Fatal("basket drawer did not preserve the visible summary")
	}
	for i, row := range m.rows() {
		if row.ID == "firefox" {
			m.cursor = i
		}
	}
	press(m, "space")
	if m.selected["firefox"] {
		t.Fatal("basket drawer did not remove the focused item")
	}
	press(m, "enter")
	if m.screen != macScreenReview {
		t.Fatal("basket bypassed the separate review screen")
	}
}

func TestActionPaletteIsNavigableAndRunsSafeActions(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category = macScreenBrowse, "tools"
	for i, row := range m.rows() {
		if row.ID == "jq" {
			m.cursor = i
			break
		}
	}
	press(m, "?")
	plain := ansi.Strip(m.View().Content)
	if !m.showHelp || !strings.Contains(plain, "Actions for") || !strings.Contains(plain, "View basket") {
		t.Fatalf("action palette missing context:\n%s", plain)
	}
	row := m.rows()[m.cursor]
	press(m, "enter")
	if m.showHelp || !m.selected[row.ID] || m.screen != macScreenBrowse {
		t.Fatal("palette action did not route through normal safe selection")
	}
}

func TestTypedHistoryRestoresBreadcrumbLocation(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category, m.appGroup, m.cursor = macScreenBrowse, "apps", "ai", 3
	origin := m.location()
	m.openScreen(macScreenBasket)
	m.openScreen(macScreenReview)
	if !m.back() || m.screen != macScreenBasket || !m.back() || m.location() != origin {
		t.Fatalf("typed history did not restore origin: %+v", m.location())
	}
}

func TestDynamicPreviewRejectsObsoleteTargetAndGeneration(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category, m.appGroup = macScreenBrowse, "apps", "development"
	for i, row := range m.rows() {
		if row.ID == "zed" {
			m.cursor = i
		}
	}
	old := m.beginDynamicPreview("firefox")
	current := m.beginDynamicPreview("zed")
	m.Update(macPreviewResult{target: "firefox", content: "stale target", generation: old})
	m.Update(macPreviewResult{target: "firefox", content: "wrong target", generation: current})
	if m.previewContent != "" {
		t.Fatal("obsolete preview content was published")
	}
	m.Update(macPreviewResult{target: "zed", content: "current", generation: current})
	if m.previewContent != "current" {
		t.Fatal("current preview content was rejected")
	}
}

func TestSemanticStatusHierarchy(t *testing.T) {
	for _, status := range []string{"installed", "skipped", "failed", "unfinished"} {
		styled := semanticStatus(status)
		if ansi.Strip(styled) != status || !strings.Contains(styled, "\x1b[") {
			t.Fatalf("status %q lost semantic styling: %q", status, styled)
		}
	}
}
