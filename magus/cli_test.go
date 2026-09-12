package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadOnlyLinuxCommandsDoNotPrepareState(t *testing.T) {
	for _, test := range []struct {
		verb string
		dry  bool
	}{
		{"version", false},
		{"help", false},
		{"doctor", false},
		{"unknown", false},
		{"run", true},
		{"reconcile", true},
		{"uninstall", true},
	} {
		t.Run(test.verb, func(t *testing.T) {
			paths := newPathsUnder(filepath.Join(t.TempDir(), "home"))
			if err := prepareLinuxPaths(paths, test.verb, test.dry); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(paths.Config); !os.IsNotExist(err) {
				t.Fatalf("read-only command prepared config directory: %v", err)
			}
		})
	}
}

func TestMutatingLinuxCommandsPrepareAndSweepState(t *testing.T) {
	for _, verb := range []string{"run", "reconcile", "uninstall"} {
		t.Run(verb, func(t *testing.T) {
			paths := newPathsUnder(filepath.Join(t.TempDir(), "home"))
			if err := paths.EnsureDirs(); err != nil {
				t.Fatal(err)
			}
			stale := filepath.Join(paths.State, "state.json"+tempMarker+".stale")
			if err := os.WriteFile(stale, []byte("stale"), 0600); err != nil {
				t.Fatal(err)
			}
			old := time.Now().Add(-2 * time.Minute)
			if err := os.Chtimes(stale, old, old); err != nil {
				t.Fatal(err)
			}
			if err := prepareLinuxPaths(paths, verb, false); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(stale); !os.IsNotExist(err) {
				t.Fatalf("stale temp was not swept: %v", err)
			}
		})
	}
}
