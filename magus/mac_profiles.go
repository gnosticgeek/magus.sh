package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/muesli/termenv"
)

var profileNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,47}$`)

func (p Paths) ProfileDir() string { return filepath.Join(p.Config, "profiles") }

func macProfilePath(paths Paths, name string) (string, error) {
	if !profileNamePattern.MatchString(name) {
		return "", fmt.Errorf("profile name %q must use lowercase letters, numbers and hyphens", name)
	}
	return filepath.Join(paths.ProfileDir(), name+".toml"), nil
}

func saveMacProfile(paths Paths, name string, manifest Manifest) (string, error) {
	path, err := macProfilePath(paths, name)
	if err != nil {
		return "", err
	}
	if err := manifestPlatformCheck(manifest, "darwin"); err != nil {
		return "", err
	}
	if err := manifest.Validate(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(paths.ProfileDir(), 0o700); err != nil {
		return "", err
	}
	temporary, err := os.CreateTemp(paths.ProfileDir(), "."+name+"-")
	if err != nil {
		return "", err
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return "", err
	}
	defer os.Remove(temporaryPath)
	if err := manifest.Save(temporaryPath); err != nil {
		return "", err
	}
	// link(2) is exclusive: unlike a rename it cannot replace a file another
	// Magus process (or the user) created after our initial validation.
	if err := os.Link(temporaryPath, path); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return "", fmt.Errorf("profile %q already exists; choose a new name", name)
		}
		return "", err
	}
	return path, nil
}

func loadMacProfile(paths Paths, name string) (Manifest, string, error) {
	path, err := macProfilePath(paths, name)
	if err != nil {
		return Manifest{}, "", err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return Manifest{}, "", err
	}
	if !info.Mode().IsRegular() {
		return Manifest{}, "", fmt.Errorf("profile %q is not a regular file", name)
	}
	manifest, err := LoadManifest(path)
	if err != nil {
		return Manifest{}, "", err
	}
	if err := manifestPlatformCheck(manifest, "darwin"); err != nil {
		return Manifest{}, "", err
	}
	if err := manifest.Validate(); err != nil {
		return Manifest{}, "", err
	}
	return manifest, path, nil
}

func listMacProfiles(paths Paths) ([]string, error) {
	entries, err := os.ReadDir(paths.ProfileDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	profiles := []string{}
	for _, entry := range entries {
		if entry.Type()&fs.ModeSymlink != 0 || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".toml")
		if profileNamePattern.MatchString(name) {
			profiles = append(profiles, name)
		}
	}
	sort.Strings(profiles)
	return profiles, nil
}

// mergeMacProfile adds a saved profile to the existing declared selection. A
// profile is a reusable starting point, not an instruction to discard choices
// already made on this Mac. Singular appearance preferences intentionally take
// the profile's value when it supplies one.
func mergeMacProfile(current, profile Manifest) (Manifest, error) {
	if err := manifestPlatformCheck(current, "darwin"); err != nil {
		return Manifest{}, err
	}
	if err := current.Validate(); err != nil {
		return Manifest{}, err
	}
	if err := manifestPlatformCheck(profile, "darwin"); err != nil {
		return Manifest{}, err
	}
	if err := profile.Validate(); err != nil {
		return Manifest{}, err
	}

	merged := current
	merged.Mac.Packages = appendUnique(current.Mac.Packages, profile.Mac.Packages)
	merged.Mac.Settings = appendUnique(current.Mac.Settings, profile.Mac.Settings)
	merged.Mac.AppConfigs = appendUnique(current.Mac.AppConfigs, profile.Mac.AppConfigs)
	if profile.Mac.Terminal != "" {
		merged.Mac.Terminal = profile.Mac.Terminal
	}
	if profile.Mac.ModernShell != nil {
		merged.Mac.ModernShell = profile.Mac.ModernShell
	}
	return merged, merged.Validate()
}

func appendUnique(current, additions []string) []string {
	result := append([]string(nil), current...)
	seen := make(map[string]bool, len(result)+len(additions))
	for _, id := range result {
		seen[id] = true
	}
	for _, id := range additions {
		if !seen[id] {
			result = append(result, id)
			seen[id] = true
		}
	}
	return result
}

func runMacProfileCLI(args []string, paths Paths, manifestPath string, dry, asJSON, plain bool, timeout time.Duration) int {
	if asJSON {
		fmt.Fprintln(os.Stderr, "magus: profile commands do not support --json")
		return 2
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "magus: usage: magus profile <list|save|show|use> [name]")
		return 2
	}
	fail := func(err error) int {
		fmt.Fprintln(os.Stderr, "magus:", err)
		return 1
	}
	switch args[0] {
	case "list":
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "magus: usage: magus profile list")
			return 2
		}
		profiles, err := listMacProfiles(paths)
		if err != nil {
			return fail(err)
		}
		for _, name := range profiles {
			fmt.Println(name)
		}
		return 0
	case "save":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "magus: usage: magus profile save NAME")
			return 2
		}
		profilePath, err := macProfilePath(paths, args[1])
		if err != nil {
			return fail(err)
		}
		manifest, err := LoadManifest(manifestPath)
		if err != nil {
			return fail(fmt.Errorf("load current selection: %w", err))
		}
		if dry {
			if err := manifestPlatformCheck(manifest, "darwin"); err != nil {
				return fail(err)
			}
			if err := manifest.Validate(); err != nil {
				return fail(err)
			}
			fmt.Printf("Would save profile %q to %s\n", args[1], profilePath)
			return 0
		}
		if _, err := saveMacProfile(paths, args[1], manifest); err != nil {
			return fail(err)
		}
		fmt.Printf("Saved profile %q to %s\n", args[1], profilePath)
		return 0
	case "show":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "magus: usage: magus profile show NAME")
			return 2
		}
		_, profilePath, err := loadMacProfile(paths, args[1])
		if err != nil {
			return fail(err)
		}
		contents, err := os.ReadFile(profilePath)
		if err != nil {
			return fail(err)
		}
		_, _ = os.Stdout.Write(contents)
		return 0
	case "use":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "magus: usage: magus profile use NAME")
			return 2
		}
		if !isTerminal(os.Stdin) {
			fmt.Fprintln(os.Stderr, "magus: open an interactive terminal to review a profile")
			return 2
		}
		profile, _, err := loadMacProfile(paths, args[1])
		if err != nil {
			return fail(err)
		}
		current, err := LoadManifest(manifestPath)
		if err == ErrNoManifest {
			current = newMacManifest()
			err = nil
		}
		if err != nil {
			return fail(fmt.Errorf("load current selection: %w", err))
		}
		merged, err := mergeMacProfile(current, profile)
		if err != nil {
			return fail(err)
		}
		if plain {
			setTerminalProfile(termenv.Ascii)
		}
		if err := runMacTUI(paths, manifestPath, merged, dry, timeout); err != nil {
			return fail(err)
		}
		return 0
	default:
		fmt.Fprintf(os.Stderr, "magus: unknown profile command %q\n", args[0])
		return 2
	}
}
