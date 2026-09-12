package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// flathubRemote is added per-user, never system-wide. A user remote lives under
// ~/.local/share/flatpak, which is on /home and therefore survives an atomic
// update; a system remote does not, and adding one needs root we do not take.
const flathubRepo = "https://dl.flathub.org/repo/flathub.flatpakrepo"
const flatpakOwnershipContents = "Installed by Magus.\n"

// flathubStep ensures the user-scoped Flathub remote exists. Every flatpak step
// depends on it, so it is ordered first in StepsFor.
type flathubStep struct{}

func (flathubStep) ID() string       { return "flathub" }
func (flathubStep) Describe() string { return "add the Flathub remote for this user" }

func (flathubStep) Check(c *Context) (State, error) {
	if !have("flatpak") {
		// SteamOS ships flatpak, so this only happens off-device. The bundles
		// simply cannot be installed; that is not an error worth failing on.
		return StateNotApplicable, nil
	}
	out, err := c.Output("flatpak", "remotes", "--user", "--columns=name")
	if err != nil {
		return StateUnknown, err
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "flathub" {
			return StateOK, nil
		}
	}
	return StateMissing, nil
}

func (flathubStep) Apply(c *Context) error {
	return c.Run("flatpak", "remote-add", "--user", "--if-not-exists", "flathub", flathubRepo)
}

// Remove is deliberately a no-op. The remote is shared infrastructure — the user
// may well have flatpaks of their own from it, and taking it away because they
// uninstalled magus would break them.
func (flathubStep) Remove(c *Context) error {
	c.Report.Detail("leaving the flathub remote in place — other apps may use it")
	return nil
}

// flatpakStep installs one flatpak application. Browsers and every bundle app go
// through this: sandboxed, user-scoped, no root, and it survives atomic updates
// because it lives on /home.
type flatpakStep struct {
	id    string // magus step id, e.g. "browser" or "bundle:essentials/vlc"
	appID string // flatpak application id, e.g. org.mozilla.firefox
	label string // human name for output
}

func (s flatpakStep) ID() string { return s.id }
func (s flatpakStep) Describe() string {
	return fmt.Sprintf("install %s (%s) as a user flatpak", s.label, s.appID)
}

// Check asks Flatpak for the complete user application list. Unlike
// `flatpak info`, this keeps "not installed" distinct from a broken probe, so a
// read failure can never become permission to install and claim ownership.
func (s flatpakStep) Check(c *Context) (State, error) {
	if !have("flatpak") {
		return StateNotApplicable, nil
	}
	out, err := c.Output("flatpak", "list", "--user", "--app", "--columns=application")
	if err != nil {
		return StateUnknown, err
	}
	for _, installed := range strings.Split(out, "\n") {
		if strings.TrimSpace(installed) == s.appID {
			return StateOK, nil
		}
	}
	return StateMissing, nil
}

func (s flatpakStep) Apply(c *Context) error {
	if err := c.Run("flatpak", "install", "--user", "--noninteractive", "--assumeyes", "flathub", s.appID); err != nil {
		return err
	}
	if c.DryRun {
		return nil
	}
	return writeFileAtomic(flatpakOwnershipPath(c, s.appID), []byte(flatpakOwnershipContents), 0o600)
}

func (s flatpakStep) Remove(c *Context) error {
	if !have("flatpak") {
		return nil
	}
	marker := flatpakOwnershipPath(c, s.appID)
	owned, err := validOwnershipMarker(marker, flatpakOwnershipContents)
	if err != nil {
		return err
	}
	if !owned {
		return errNotReversible
	}
	if err := c.Run("flatpak", "uninstall", "--user", "--noninteractive", "--assumeyes", s.appID); err != nil {
		return err
	}
	if c.DryRun {
		return nil
	}
	return os.Remove(marker)
}

func flatpakOwnershipPath(c *Context, appID string) string {
	return filepath.Join(c.Paths.State, "ownership", "flatpak", appID)
}
