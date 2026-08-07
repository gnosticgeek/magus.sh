package main

import "fmt"

// browserApps maps a manifest choice to its Flathub application id. Flatpak for
// every browser, without exception — that is the single consistent rule the
// brief asks for (§15) rather than deciding case by case.
var browserApps = map[string]struct{ appID, label string }{
	"firefox": {"org.mozilla.firefox", "Firefox"},
	"brave":   {"com.brave.Browser", "Brave"},
	"chrome":  {"com.google.Chrome", "Google Chrome"},
	"vivaldi": {"com.vivaldi.Vivaldi", "Vivaldi"},
}

// bundleApp is one application within a bundle.
type bundleApp struct {
	appID string
	label string
}

// bundles is the §4 step-3 set. Everything here is a user flatpak, so a bundle
// is data rather than code — adding an app is one line, and it inherits the
// idempotency and removability of flatpakStep for free.
//
// Deliberately absent: Decky Loader, whose installer requires root and so falls
// outside the userland-only rule; and MangoHud, which installs as a Vulkan layer
// extension whose branch must match the runtime — both want their own step
// rather than being wedged in here.
var bundles = map[string][]bundleApp{
	"essentials": {
		{"org.kde.ark", "Ark (archives)"},
		{"org.kde.kate", "Kate (text editor)"},
		{"io.mpv.Mpv", "mpv (media player)"},
		{"org.kde.gwenview", "Gwenview (image viewer)"},
	},
	"gaming": {
		{"net.davidotek.pupgui2", "ProtonUp-Qt"},
		{"com.heroicgameslauncher.hgl", "Heroic Games Launcher"},
		{"net.lutris.Lutris", "Lutris"},
	},
	"creative": {
		{"org.gimp.GIMP", "GIMP"},
		{"org.kde.krita", "Krita"},
		{"org.inkscape.Inkscape", "Inkscape"},
		{"com.obsproject.Studio", "OBS Studio"},
	},
	"dev": {
		{"com.visualstudio.code", "VS Code"},
	},
	"comms": {
		{"com.discordapp.Discord", "Discord"},
		{"org.signal.Signal", "Signal"},
		{"im.riot.Riot", "Element"},
	},
}

// bundleOrder fixes the order bundles are applied in. Ranging over the map would
// give a different order every run, which makes output impossible to compare
// between runs and uninstall order non-deterministic.
var bundleOrder = []string{"essentials", "gaming", "creative", "dev", "comms"}

// StepsFor turns a manifest into the ordered plan the reconciler executes.
//
// This function is the whole of "the wizard is a manifest builder and nothing
// more": everything downstream of here depends only on the manifest, so the
// TUI, `--defaults`, a hand-edited file and the eventual GUI all converge
// through identical code.
//
// It reads the device from the manifest rather than from live detection, so the
// plan is a pure function of the file. A manifest carried to a different machine
// still describes the same intent; it is each step's Check that decides whether
// that intent applies to the hardware actually present.
func StepsFor(m Manifest) []Step {
	steps := []Step{flathubStep{}}

	switch m.Choices.Terminal {
	case "kitty":
		steps = append(steps, kittyStep{})
	case "konsole":
		steps = append(steps, keepKonsoleStep{})
	case "ghostty", "alacritty":
		steps = append(steps, notYetStep{
			id:     "terminal:" + m.Choices.Terminal,
			choice: m.Choices.Terminal,
			reason: "only kitty and konsole are wired up in v0.2",
		})
	}

	if b, ok := browserApps[m.Choices.Browser]; ok {
		steps = append(steps, flatpakStep{id: "browser:" + m.Choices.Browser, appID: b.appID, label: b.label})
	}

	for _, name := range bundleOrder {
		if !m.HasBundle(name) {
			continue
		}
		for _, app := range bundles[name] {
			steps = append(steps, flatpakStep{
				id:    fmt.Sprintf("bundle:%s/%s", name, app.appID),
				appID: app.appID,
				label: app.label,
			})
		}
	}

	// --- optimisations (§4 step 4) ---
	// Ordered after the installs because a couple of them touch things the
	// installs create, and because a failed tweak should never cost someone
	// their browser.
	steps = append(steps, kdeSteps(m)...)
	if m.Optimisations.ProtonGE {
		steps = append(steps, protonGEStep{})
	}
	steps = append(steps, machineOptimisations(m)...)

	// Theming (§4 step 5) is v0.4. The manifest carries the choice already.

	return steps
}

// machineOptimisations are the display and power settings that only make sense
// on something plugged into a television and a wall.
//
// They are emitted whenever the manifest asks for them, even on a Deck — the
// step's own Check is what declines. That way `doctor` on a Deck with a
// Machine's manifest says "not applicable" rather than quietly omitting them,
// which is the difference between a tool that explains itself and one that
// looks broken.
func machineOptimisations(m Manifest) []Step {
	var out []Step
	if m.Optimisations.HDMIFullRange {
		out = append(out, notYetOptimisation{
			id:     "optimise:hdmi-full-range",
			label:  "HDMI colour range → full RGB",
			reason: "needs a verified DRM property on real hardware — §10",
		})
	}
	if m.Optimisations.CEC {
		out = append(out, notYetOptimisation{
			id:     "optimise:cec",
			label:  "HDMI-CEC",
			reason: "SteamOS added CEC controls; the userland hook is unverified — §10",
		})
	}
	if p := m.Optimisations.PowerProfile; p != "" && p != "balanced" {
		out = append(out, notYetOptimisation{
			id:     "optimise:power-profile",
			label:  "power profile → " + p,
			reason: "needs powerprofilesctl confirmed on hardware — §10",
		})
	}
	return out
}
