package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestMacProfilesSaveLoadAndList(t *testing.T) {
	paths := newPathsUnder(filepath.Join(t.TempDir(), "home"))
	manifest := newMacManifest()
	manifest.Mac.Packages = []string{"ghostty", "ripgrep"}
	manifest.Mac.Settings = []string{"finder-extensions"}

	path, err := saveMacProfile(paths, "work-dev", manifest)
	if err != nil {
		t.Fatal(err)
	}
	if want, _ := macProfilePath(paths, "work-dev"); path != want {
		t.Fatalf("profile path = %q, want %q", path, want)
	}
	loaded, loadedPath, err := loadMacProfile(paths, "work-dev")
	if err != nil {
		t.Fatal(err)
	}
	if loadedPath != path || !reflect.DeepEqual(loaded, manifest) {
		t.Fatalf("loaded profile = %#v at %q, want %#v at %q", loaded, loadedPath, manifest, path)
	}
	if _, err := saveMacProfile(paths, "work-dev", manifest); err == nil {
		t.Fatal("saving an existing profile succeeded")
	}

	second := newMacManifest()
	if _, err := saveMacProfile(paths, "base", second); err != nil {
		t.Fatal(err)
	}
	profiles, err := listMacProfiles(paths)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"base", "work-dev"}; !reflect.DeepEqual(profiles, want) {
		t.Fatalf("profiles = %v, want %v", profiles, want)
	}
}

func TestMacProfileNamesAreConstrained(t *testing.T) {
	paths := newPathsUnder(filepath.Join(t.TempDir(), "home"))
	for _, name := range []string{"", "Work", "work_dev", "../work", "work.toml"} {
		if _, err := macProfilePath(paths, name); err == nil {
			t.Errorf("macProfilePath(%q) succeeded", name)
		}
	}
	if _, err := macProfilePath(paths, "personal-2026"); err != nil {
		t.Fatalf("valid profile name rejected: %v", err)
	}
}

func TestMergeMacProfilePreservesBasketAndOverlaysPreferences(t *testing.T) {
	current := newMacManifest()
	current.Mac.Packages = []string{"ghostty", "ripgrep"}
	current.Mac.Settings = []string{"finder-extensions"}
	current.Mac.Terminal = "TokyoNight"

	profile := newMacManifest()
	profile.Mac.Packages = []string{"ripgrep", "fd", "git"}
	profile.Mac.Settings = []string{"finder-pathbar"}
	profile.Mac.Terminal = "Catppuccin"

	merged, err := mergeMacProfile(current, profile)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"ghostty", "ripgrep", "fd", "git"}; !reflect.DeepEqual(merged.Mac.Packages, want) {
		t.Fatalf("packages = %v, want %v", merged.Mac.Packages, want)
	}
	if want := []string{"finder-extensions", "finder-pathbar"}; !reflect.DeepEqual(merged.Mac.Settings, want) {
		t.Fatalf("settings = %v, want %v", merged.Mac.Settings, want)
	}
	if merged.Mac.Terminal != "Catppuccin" {
		t.Fatalf("terminal = %q", merged.Mac.Terminal)
	}
}
