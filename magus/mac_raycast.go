package main

import (
	tea "charm.land/bubbletea/v2"
	"os/exec"
)

func raycastRows() []macRow {
	return []macRow{
		{ID: "raycast-install", Name: "Add Raycast to basket", Summary: "Install the app through Review & install.", Note: raycastGuide},
		{ID: "https://manual.raycast.com", Name: "Settings checklist / open manual", Summary: "Spotlight shortcut, clipboard, snippets, Hyper Key, search cleanup and Auto Quit.", Note: raycastGuide},
		{ID: "https://www.raycast.com/raycast/github", Name: "GitHub", Summary: "Open extension page for pull requests, issues and notifications."},
		{ID: "https://www.raycast.com/thomas/visual-studio-code", Name: "Visual Studio Code", Summary: "Open extension page for recent projects."},
		{ID: "https://www.raycast.com/nhojb/brew", Name: "Brew", Summary: "Open extension page for Homebrew management."},
		{ID: "https://www.raycast.com/rolandleth/kill-process", Name: "Kill Process", Summary: "Open extension page to inspect and quit processes."},
		{ID: "https://www.raycast.com/mooxl/coffee", Name: "Coffee", Summary: "Open extension page to keep your Mac awake."},
		{ID: "https://www.raycast.com/tonka3000/speedtest", Name: "Speedtest", Summary: "Open extension page for connection testing."},
		{ID: "https://manual.raycast.com/import-export", Name: "Import / export instructions", Summary: "Move your own Raycast configuration using its supported commands."},
		{ID: "export:raycast", Name: "Export Raycast setup guide", Summary: "Save the settings checklist and extension links as Markdown.", Note: raycastGuide},
		{ID: "review", Name: "Review & install", Summary: "Review your basket before installing."},
		{ID: "app-configs", Name: "Back to app setups"},
	}
}
func openSetupURL(url string) tea.Cmd {
	// Only URLs from the bundled rows can be opened; no shell interpolation.
	return func() tea.Msg {
		for _, row := range append(raycastRows(), firefoxExtensionRows()...) {
			if row.ID == url && len(url) > 8 && url[:8] == "https://" {
				return setupOpened{err: exec.Command("open", url).Run()}
			}
		}
		return nil
	}
}

type setupOpened struct{ err error }

func firefoxExtensionRows() []macRow {
	return []macRow{
		{ID: "https://addons.mozilla.org/firefox/addon/ublock-origin/", Name: "Firefox / uBlock Origin", Summary: "Open add-on page for ad and tracker blocking."},
		{ID: "https://addons.mozilla.org/firefox/addon/darkreader/", Name: "Firefox / Dark Reader", Summary: "Open add-on page for website dark mode."},
		{ID: "https://addons.mozilla.org/firefox/addon/bitwarden-password-manager/", Name: "Firefox / Bitwarden", Summary: "Open add-on page for password management."},
		{ID: "https://addons.mozilla.org/firefox/addon/multi-account-containers/", Name: "Firefox / Multi-Account Containers", Summary: "Open add-on page for separate browsing identities."},
		{ID: "https://addons.mozilla.org/firefox/addon/sponsorblock/", Name: "Firefox / SponsorBlock", Summary: "Open add-on page for skipping sponsored segments."},
	}
}
