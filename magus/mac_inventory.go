package main

import (
	"context"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
)

const maxConcurrentPackageInspections = 4

type packageInspection struct {
	id    string
	label string
}

// inspectPackageStates uses a small fixed worker pool. Package checks can
// include filesystem probes, while all Homebrew output has already been
// snapshotted before this point. Keeping the pool bounded avoids turning a
// large catalogue into an unbounded set of filesystem operations.
func inspectPackageStates(ctx context.Context, c *Context, packages []MacPackage) map[string]string {
	states := make(map[string]string, len(packages))
	if len(packages) == 0 {
		return states
	}
	workers := min(maxConcurrentPackageInspections, len(packages))
	jobs := make(chan MacPackage)
	results := make(chan packageInspection, len(packages))
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case p, ok := <-jobs:
					if !ok {
						return
					}
					label := "needs Homebrew"
					if c.brewPath() != "" {
						state, err := (brewStep{p}).Check(c)
						label = "not installed"
						if err != nil {
							label = "inspection failed"
						} else if state == StateOK {
							label = "installed"
						} else if state == StateNotApplicable {
							label = "outside Homebrew"
						}
					}
					select {
					case <-ctx.Done():
						return
					case results <- packageInspection{id: p.ID, label: label}:
					}
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, p := range packages {
			select {
			case <-ctx.Done():
				return
			case jobs <- p:
			}
		}
	}()
	go func() {
		group.Wait()
		close(results)
	}()
	for result := range results {
		states[result.id] = result.label
	}
	return states
}

// inspect schedules a fresh, read-only snapshot of package and preference
// state. Results carry a generation so a superseded probe cannot prune the
// current basket after a later inspection has already completed.
func (m *macModel) inspect() tea.Cmd {
	if m.inventoryCancel != nil {
		m.inventoryCancel()
	}
	generation := m.inventoryGeneration.next()
	ctx, cancel := context.WithCancel(context.Background())
	m.inventoryCancel = cancel
	paths := m.paths
	return func() tea.Msg {
		defer cancel()
		brew := findBrew()
		inv := macInventory{states: map[string]string{}, brew: brew != "", appleSilicon: runtime.GOARCH == "arm64", osVersion: runtime.GOOS + " / " + runtime.GOARCH}
		if runtime.GOOS != "darwin" {
			return macInventoryResult{inv, generation}
		}
		c := &Context{Paths: paths, Brew: brew, Parent: ctx, Timeout: 30 * time.Second, Report: &Reporter{Out: io.Discard}}
		if v, err := c.macCommand("/usr/bin/sw_vers", "-productVersion"); err == nil {
			version := strings.TrimSpace(v)
			inv.osVersion = "macOS " + version + " / " + runtime.GOARCH
			inv.osMajor = macVersionMajor(version)
		}
		if v, err := c.macCommand("/usr/bin/xcodebuild", "-version"); err == nil {
			inv.xcodeMajor = macVersionMajor(strings.TrimPrefix(strings.TrimSpace(strings.SplitN(v, "\n", 2)[0]), "Xcode "))
		}
		// Snapshot each package kind once for browsing. Execution always probes anew.
		outputs := map[string]string{}
		errs := map[string]error{}
		if inv.brew {
			for _, kind := range []string{"cask", "formula"} {
				outputs[kind], errs[kind] = c.macCommand(c.brewPath(), "list", "--"+kind, "-1")
			}
		}
		c.Execute = func(ctx context.Context, name string, args ...string) (string, error) {
			if len(args) == 3 && args[0] == "list" {
				kind := strings.TrimPrefix(args[1], "--")
				return outputs[kind], errs[kind]
			}
			return runMacCommand(ctx, nil, name, args...)
		}
		for id, label := range inspectPackageStates(ctx, c, macPackages) {
			inv.states[id] = label
		}
		for _, s := range macSettings {
			state, err := (preferenceStep{s}).Check(c)
			label := "not applied"
			if err != nil {
				label = "inspection failed"
			} else if state == StateOK {
				label = "already set"
			}
			inv.states[s.ID] = label
		}
		return macInventoryResult{inv, generation}
	}
}

func macVersionMajor(version string) int {
	major, _, _ := strings.Cut(version, ".")
	n, _ := strconv.Atoi(major)
	return n
}
