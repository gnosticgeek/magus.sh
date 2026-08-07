package main

import (
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

	// Resolve the latest release tarball from the GitHub API, then stream it
	// straight into tar — no temp file to clean up, and nothing left behind if
	// the download fails part-way.
	const api = "https://api.github.com/repos/GloriousEggroll/proton-ge-custom/releases/latest"
	url, err := c.Output("sh", "-c",
		fmt.Sprintf(`curl -fsSL %q | grep -o '"browser_download_url": *"[^"]*\.tar\.gz"' | head -1 | cut -d'"' -f4`, api))
	if err != nil {
		return fmt.Errorf("could not reach the GE-Proton release feed: %w", err)
	}
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("could not find a GE-Proton tarball in the latest release")
	}
	c.Report.Detail("downloading %s", filepath.Base(url))
	return c.Run("sh", "-c", fmt.Sprintf(`curl -fsSL %q | tar -xzf - -C %q`, url, dir))
}

func (protonGEStep) Remove(c *Context) error {
	for _, d := range compatDirs(c) {
		entries, err := os.ReadDir(d)
		if err != nil {
			continue
		}
		for _, e := range entries {
			// Only remove GE-Proton directories. Anything else in here is a
			// compatibility tool someone else installed.
			if e.IsDir() && strings.HasPrefix(e.Name(), "GE-Proton") {
				if err := removePath(c, filepath.Join(d, e.Name())); err != nil {
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
