package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestPackageInspectionUsesABoundedWorkerPool(t *testing.T) {
	packages := make([]MacPackage, maxConcurrentPackageInspections+3)
	for i := range packages {
		packages[i] = MacPackage{ID: string(rune('a' + i)), Kind: "formula"}
	}
	var active, peak atomic.Int32
	release := make(chan struct{})
	started := make(chan struct{}, maxConcurrentPackageInspections)
	c := &Context{Brew: "brew", Execute: func(context.Context, string, ...string) (string, error) {
		n := active.Add(1)
		for {
			old := peak.Load()
			if n <= old || peak.CompareAndSwap(old, n) {
				break
			}
		}
		started <- struct{}{}
		<-release
		active.Add(-1)
		return "", nil
	}}
	finished := make(chan map[string]string, 1)
	go func() { finished <- inspectPackageStates(context.Background(), c, packages) }()
	for range maxConcurrentPackageInspections {
		<-started
	}
	if got := peak.Load(); got != maxConcurrentPackageInspections {
		t.Fatalf("peak workers = %d, want %d", got, maxConcurrentPackageInspections)
	}
	close(release)
	states := <-finished
	if len(states) != len(packages) {
		t.Fatalf("states = %d, want %d", len(states), len(packages))
	}
}

func TestAsyncGenerationRejectsSupersededResults(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)

	oldInventory := m.inventoryGeneration.next()
	currentInventory := m.inventoryGeneration.next()
	m.Update(macInventoryResult{inventory: macInventory{states: map[string]string{"git": "installed"}}, generation: oldInventory})
	if m.inventory.states["git"] != "" {
		t.Fatal("superseded inventory changed the active model")
	}
	m.Update(macInventoryResult{inventory: macInventory{states: map[string]string{"git": "installed"}}, generation: currentInventory})
	if m.inventory.states["git"] != "installed" {
		t.Fatal("current inventory result was not accepted")
	}

	oldSession := m.sessionGeneration.next()
	currentSession := m.sessionGeneration.next()
	m.outcomes = []macOutcome{{ID: "original", Name: "Original", Status: "unfinished"}}
	m.Update(macEvent{Kind: macEventPlan, generation: oldSession, Outcomes: []macOutcome{{ID: "stale"}}})
	if m.outcomes[0].ID != "original" {
		t.Fatal("superseded session event changed the active plan")
	}
	m.Update(macEvent{Kind: macEventPlan, generation: currentSession, Outcomes: []macOutcome{{ID: "current"}}})
	if m.outcomes[0].ID != "current" {
		t.Fatal("current session event was not accepted")
	}
}

func TestScreenCapabilitiesCoverInteractiveViews(t *testing.T) {
	for _, screen := range []macScreen{
		macScreenMenu, macScreenDeveloper, macScreenCategories, macScreenBrowse, macScreenReview,
		macScreenPresets, macScreenTerminal, macScreenShell, macScreenAppConfigs,
		macScreenRaycast,
	} {
		if !screen.supportsDetails() {
			t.Fatalf("%q unexpectedly rejects the details pane", screen)
		}
	}
	for _, screen := range []macScreen{macScreenInstall, macScreenSummary, macScreenBootstrap, macScreenRestore, macScreenUpdates, macScreenUpdateConfirm, macScreenTerminalRestore} {
		if screen.supportsDetails() {
			t.Fatalf("%q unexpectedly accepts the details pane", screen)
		}
	}
}
