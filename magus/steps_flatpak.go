package main

import (
	"fmt"
	"strings"
)

// flathubRemote is added per-user, never system-wide. A user remote lives under
// ~/.local/share/flatpak, which is on /home and therefore survives an atomic
// update; a system remote does not, and adding one needs root we do not take.
const flathubRepo = "https://dl.flathub.org/repo/flathub.flatpakrepo"

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

// Check asks flatpak itself rather than looking for a path. `flatpak info`
// exiting zero is the authoritative answer to "is this installed for this user",
// and it stays correct across flatpak's own layout changes.
func (s flatpakStep) Check(c *Context) (State, error) {
	if !have("flatpak") {
		return StateNotApplicable, nil
	}
	if _, err := c.Output("flatpak", "info", "--user", s.appID); err != nil {
		// A non-zero exit here means "not installed for this user". flatpak does
		// not distinguish that from other failures by exit code, but every other
		// failure mode (a broken installation, an unreadable repo) also warrants
		// running install again, so treating it as missing is both safe and
		// idempotent.
		return StateMissing, nil
	}
	return StateOK, nil
}

func (s flatpakStep) Apply(c *Context) error {
	return c.Run("flatpak", "install", "--user", "--noninteractive", "--assumeyes", "flathub", s.appID)
}

func (s flatpakStep) Remove(c *Context) error {
	if !have("flatpak") {
		return nil
	}
	return c.Run("flatpak", "uninstall", "--user", "--noninteractive", "--assumeyes", s.appID)
}
