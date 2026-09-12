package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const outdatedFixture = `{"formulae":[{"name":"jq","installed_versions":["1.6"],"current_version":"1.7","pinned":false},{"name":"node","installed_versions":["20"],"current_version":"22","pinned":true}],"casks":[{"name":"iina","installed_versions":["1.3"],"current_version":"1.4"}]}`

func TestOutdatedReviewRejectsBadDataAndSkipsPinned(t *testing.T) {
	items, err := parseMacOutdated(outdatedFixture)
	if err != nil || len(items) != 2 {
		t.Fatalf("%v %v", items, err)
	}
	if items[0].ID != "iina" || items[1].ID != "jq" || items[1].Available != "1.7" {
		t.Fatal(items)
	}
	for _, bad := range []string{`{}`, `not json`, `{"formulae":[],"casks":[{"name":"--force","installed_versions":["1"],"current_version":"2"}]}`} {
		if _, err := parseMacOutdated(bad); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
	empty, err := parseMacOutdated(`{"formulae":[],"casks":[]}`)
	if err != nil || len(empty) != 0 {
		t.Fatal("empty list must be valid")
	}
}

func TestReviewedUpgradeUsesExactTargetsAndRejectsChanges(t *testing.T) {
	for _, changed := range []bool{false, true} {
		t.Run(map[bool]string{false: "reviewed targets", true: "changed version"}[changed], func(t *testing.T) {
			dir := t.TempDir()
			script := filepath.Join(dir, "brew")
			log := filepath.Join(dir, "calls")
			t.Setenv("MAGUS_TEST_CALLS", log)
			fixture := outdatedFixture
			if changed {
				fixture = strings.ReplaceAll(fixture, "1.7", "1.8")
			}
			t.Setenv("MAGUS_TEST_OUTDATED", fixture)
			content := "#!/bin/sh\nif [ \"$1\" = outdated ]; then printf '%s' \"$MAGUS_TEST_OUTDATED\"; exit 0; fi\nprintf '%s\\n' \"$*\" >> \"$MAGUS_TEST_CALLS\"\n"
			if err := os.WriteFile(script, []byte(content), 0700); err != nil {
				t.Fatal(err)
			}
			items, _ := parseMacOutdated(outdatedFixture)
			cmd := macUpdateCommand{brew: script, items: items, timeout: time.Second, out: io.Discard, errOut: io.Discard}
			err := cmd.Run()
			if changed {
				if err == nil || !strings.Contains(err.Error(), "changed") {
					t.Fatal(err)
				}
				if _, err := os.Stat(log); !os.IsNotExist(err) {
					t.Fatal("stale review ran upgrade")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(log)
				if err != nil || string(data) != "upgrade --cask iina\nupgrade --formula jq\n" {
					t.Fatalf("%s %v", data, err)
				}
			}
		})
	}
}

func TestBrowserFuzzyPagingAndDetails(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category = macScreenBrowse, "tools"
	m.Update(tea.WindowSizeMsg{Width: 72, Height: 20})
	m.View()
	m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	if m.cursor < 2 {
		t.Fatal("page down did not move")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	if m.rows()[m.cursor].ID != "mole" {
		t.Fatal("end failed")
	}
	press(m, "space")
	press(m, "/")
	for _, r := range "rgp" {
		press(m, string(r))
	}
	found := false
	for _, r := range m.rows() {
		if r.ID == "ripgrep" {
			found = true
		}
	}
	if !found || !m.selected["mole"] {
		t.Fatal("fuzzy search lost match or basket")
	}
	press(m, "esc")
	m.details = true
	m.detailsView(strings.Repeat("Long setup notes\n", 40), 60, 8)
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if m.previewPane.YOffset() == 0 {
		t.Fatal("details did not scroll")
	}
	press(m, "tab")
	if m.details {
		t.Fatal("Tab did not return focus")
	}
}

func TestCatalogueFiltersKeepTheBasketIntact(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen, m.category = macScreenBrowse, "tools"
	m.inventory.states["git"] = "installed"
	m.selected["ripgrep"] = true

	m.catalogueFilter = macCatalogueAvailable
	for _, row := range m.rows() {
		if row.ID == "git" {
			t.Fatal("not-installed filter included an installed tool")
		}
	}
	m.cycleCatalogueFilter()
	if m.catalogueFilterLabel() != "Selected only" || len(m.rows()) != 1 || m.rows()[0].ID != "ripgrep" {
		t.Fatalf("selected filter = %q, rows = %+v", m.catalogueFilterLabel(), m.rows())
	}
	if !m.selected["ripgrep"] {
		t.Fatal("changing the catalogue filter changed the basket")
	}
}

func TestSummaryAndFailureDiagnosticAreActionable(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.inventory.osVersion = "macOS test / arm64"
	m.outcomes = []macOutcome{
		{ID: "package:jq", Name: "jq", Status: "installed"},
		{ID: "package:git", Name: "Git", Status: "already present"},
		{ID: "package:bad", Name: "Bad package", Status: "failed", Detail: "brew exited 1"},
	}
	m.active = 2
	counts := m.outcomeCounts()
	if counts["installed"] != 1 || counts["already present"] != 1 || counts["failed"] != 1 {
		t.Fatal(counts)
	}
	if diagnostic := m.failureDiagnostic(); !strings.Contains(diagnostic, "package:bad") || !strings.Contains(diagnostic, "Retry with r") {
		t.Fatal(diagnostic)
	}
	if next := m.summaryNextAction(); !strings.Contains(next, "inspect logs") {
		t.Fatal(next)
	}
}

func TestLogsKeepScrollPositionWhenOutputArrives(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.showLogs = true
	m.setLogs(strings.Repeat("line\n", 40))
	m.logs.GotoTop()
	m.setLogs(m.logText + "new output\n")
	if m.logs.YOffset() != 0 {
		t.Fatal("new log forced reader to bottom")
	}
	m.logs.GotoBottom()
	m.setLogs(m.logText + "new output\n")
	if !m.logs.AtBottom() {
		t.Fatal("live following stopped")
	}
}

func TestUpdateReviewFitsAndRequiresConfirmation(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen = macScreenUpdates
	items, _ := parseMacOutdated(outdatedFixture)
	m.acceptUpdateReview(macUpdatesChecked{items: items})
	for _, size := range [][2]int{{72, 20}, {80, 24}, {120, 35}} {
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		view := m.View().Content
		if lipgloss.Width(view) > size[0] || lipgloss.Height(view) > size[1] {
			t.Fatal("update review overflows")
		}
		if !strings.Contains(view, "Installed") || !strings.Contains(view, "1.4") {
			t.Fatal(view)
		}
	}
	press(m, "enter")
	if m.screen != macScreenUpdateConfirm {
		t.Fatal("confirmation missing")
	}
	press(m, "enter")
	if !strings.Contains(m.notice, "no updates") {
		t.Fatal("preview protection missing")
	}
}

func TestNoticeExpiryPreservesNewerMessages(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.flash("Selected jq")
	old := macNoticeExpired{m.noticeGeneration, "Selected jq"}
	m.notice = "Inspection failed"
	m.Update(old)
	if m.notice != "Inspection failed" {
		t.Fatal("selection timer erased error")
	}
	m.flash("Selected Git")
	m.Update(old)
	if m.notice != "Selected Git" {
		t.Fatal("old timer erased newer selection")
	}
	m.Update(macNoticeExpired{m.noticeGeneration, m.notice})
	if m.notice != "" {
		t.Fatal("selection notice did not clear")
	}
}

func TestStaleUpdateInspectionCannotReplaceNewReview(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen = macScreenUpdates
	m.updateReview.generation = 2
	m.Update(macUpdatesChecked{generation: 1})
	if m.updateReview.ready {
		t.Fatal("accepted old inspection")
	}
	m.screen = macScreenMenu
	m.Update(macUpdatesChecked{generation: 2})
	if m.updateReview.ready {
		t.Fatal("inspection reopened exited review")
	}
}

func TestUpgradeFailureStopsRemainingTargets(t *testing.T) {
	dir := t.TempDir()
	script, log := filepath.Join(dir, "brew"), filepath.Join(dir, "calls")
	t.Setenv("MAGUS_TEST_CALLS", log)
	t.Setenv("MAGUS_TEST_OUTDATED", outdatedFixture)
	content := "#!/bin/sh\nif [ \"$1\" = outdated ]; then printf '%s' \"$MAGUS_TEST_OUTDATED\"; exit 0; fi\nprintf '%s\\n' \"$*\" >> \"$MAGUS_TEST_CALLS\"\nexit 1\n"
	if err := os.WriteFile(script, []byte(content), 0700); err != nil {
		t.Fatal(err)
	}
	items, _ := parseMacOutdated(outdatedFixture)
	cmd := macUpdateCommand{brew: script, items: items, timeout: time.Second, out: io.Discard, errOut: io.Discard}
	if err := cmd.Run(); err == nil {
		t.Fatal("upgrade failure hidden")
	}
	data, err := os.ReadFile(log)
	if err != nil || string(data) != "upgrade --cask iina\n" {
		t.Fatalf("continued after failure: %s %v", data, err)
	}
}
