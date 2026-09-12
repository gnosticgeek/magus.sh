package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// protonGEStep installs GE-Proton into Steam's compatibilitytools.d.
//
// Entirely userland: the directory lives under $HOME and survives an atomic
// update for the same reason everything else magus installs does.
type protonGEStep struct{}

const protonGEOwnershipMarker = ".magus-owned"
const protonGEOwnershipContents = "Installed by Magus.\n"

func (protonGEStep) ID() string { return "optimise:proton-ge" }
func (protonGEStep) Describe() string {
	return "install the latest GE-Proton into Steam's compatibilitytools.d"
}

// compatDirs are the places Steam looks for custom compatibility tools. The
// first that already exists wins; on a Deck that is ~/.steam/root.
func compatDirs(c *Context) []string {
	return []string{
		filepath.Join(c.Paths.Home, ".steam", "root", "compatibilitytools.d"),
		filepath.Join(c.Paths.Home, ".local", "share", "Steam", "compatibilitytools.d"),
		filepath.Join(c.Paths.Home, ".steam", "steam", "compatibilitytools.d"),
	}
}

// steamRoot returns the compatibilitytools.d to install into, and whether a
// Steam installation was found at all. Without Steam there is nothing to extend.
func steamRoot(c *Context) (string, bool) {
	for _, d := range compatDirs(c) {
		// The parent existing is what proves this is Steam's directory; the
		// compatibilitytools.d itself may not have been created yet.
		if info, err := os.Stat(filepath.Dir(d)); err == nil && info.IsDir() {
			return d, true
		}
	}
	return "", false
}

// installedProtonGE returns the name of any GE-Proton already present.
func installedProtonGE(c *Context) string {
	for _, d := range compatDirs(c) {
		entries, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "GE-Proton") {
				return e.Name()
			}
		}
	}
	return ""
}

func (protonGEStep) Check(c *Context) (State, error) {
	if !c.Manifest.Optimisations.ProtonGE {
		return StateNotApplicable, nil
	}
	if _, ok := steamRoot(c); !ok {
		// No Steam on this machine — a dev laptop, say. Not a failure.
		return StateNotApplicable, nil
	}
	if installedProtonGE(c) == "" {
		return StateMissing, nil
	}
	// Deliberately presence, not latest-version: re-running magus should not
	// pull a fresh multi-hundred-megabyte release every time upstream tags one.
	// ProtonUp-Qt (in the gaming bundle) is the right tool for upgrading.
	return StateOK, nil
}

func (protonGEStep) Apply(c *Context) error {
	if installedProtonGE(c) != "" {
		return nil
	}
	dir, ok := steamRoot(c)
	if !ok {
		return fmt.Errorf("no Steam installation found under %s", c.Paths.Home)
	}
	if !have("curl") || !have("tar") {
		return fmt.Errorf("curl and tar are required to install GE-Proton")
	}
	if c.DryRun {
		c.Report.Detail("would install the latest GE-Proton into %s", dir)
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Resolve the latest release tarball from the GitHub API. Download to a
	// temporary archive so the URL is passed as a process argument rather than
	// interpolated into a shell pipeline.
	const api = "https://api.github.com/repos/GloriousEggroll/proton-ge-custom/releases/latest"
	feed, err := c.Output("curl", "-fsSL", api)
	if err != nil {
		return fmt.Errorf("could not reach the GE-Proton release feed: %w", err)
	}
	url, ok := protonGETarballURL([]byte(feed))
	if !ok {
		return fmt.Errorf("could not find a GE-Proton tarball in the latest release")
	}
	c.Report.Detail("downloading %s", filepath.Base(url))
	archive, err := os.CreateTemp("", "magus-proton-ge-*.tar.gz")
	if err != nil {
		return err
	}
	archivePath := archive.Name()
	if err := archive.Close(); err != nil {
		os.Remove(archivePath)
		return err
	}
	defer os.Remove(archivePath)
	if err := c.Run("curl", "-fsSL", "--output", archivePath, url); err != nil {
		return err
	}
	if err := c.Run("tar", "-xzf", archivePath, "-C", dir); err != nil {
		return err
	}
	if c.DryRun {
		return nil
	}
	installed := installedProtonGE(c)
	if installed == "" {
		return fmt.Errorf("GE-Proton archive did not create an installation directory")
	}
	marker := filepath.Join(dir, installed, protonGEOwnershipMarker)
	return writeFileAtomic(marker, []byte(protonGEOwnershipContents), 0o600)
}

func protonGETarballURL(data []byte) (string, bool) {
	var release struct {
		Assets []struct {
			URL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if json.Unmarshal(data, &release) != nil {
		return "", false
	}
	const releasePrefix = "https://github.com/GloriousEggroll/proton-ge-custom/releases/download/"
	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.URL, releasePrefix) && strings.HasSuffix(asset.URL, ".tar.gz") {
			return asset.URL, true
		}
	}
	return "", false
}

func (protonGEStep) Remove(c *Context) error {
	for _, d := range compatDirs(c) {
		entries, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range entries {
			// The name identifies the tool; the marker proves Magus installed it.
			if !e.IsDir() || !strings.HasPrefix(e.Name(), "GE-Proton") {
				continue
			}
			path := filepath.Join(d, e.Name())
			owned, markerErr := validOwnershipMarker(filepath.Join(path, protonGEOwnershipMarker), protonGEOwnershipContents)
			if markerErr != nil {
				return markerErr
			}
			if owned {
				if err := removePath(c, path); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// notYetOptimisation stands in for a §4 optimisation whose mechanism has not
// been verified on hardware (§10). It reports itself rather than being absent,
// so `doctor` lists what the manifest asked for and did not get.
type notYetOptimisation struct {
	id     string
	label  string
	reason string
}

func (s notYetOptimisation) ID() string { return s.id }
func (s notYetOptimisation) Describe() string {
	return fmt.Sprintf("%s — not built yet (%s)", s.label, s.reason)
}

// Why satisfies explainer, so the engine prints the reason on the same line as
// the n/a rather than the step emitting one of its own.
func (s notYetOptimisation) Why() string { return "not built yet: " + s.reason }

func (s notYetOptimisation) Check(*Context) (State, error) { return StateNotApplicable, nil }
func (s notYetOptimisation) Apply(*Context) error          { return nil }
func (s notYetOptimisation) Remove(*Context) error         { return nil }
