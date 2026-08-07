package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// Version is the manifest schema version magus writes. It is recorded in every
// manifest so a later release can migrate one written by an earlier one —
// including retiring something an earlier version installed, which is what makes
// "converge to the current intended state" different from "don't crash on
// re-run" (§3).
const Version = "0.3.0"

// Manifest is the declarative record of what this machine should look like.
// It is the only thing the wizard produces and the only thing the reconciler
// reads. The future GUI reads and writes this same file rather than being a
// second implementation.
type Manifest struct {
	Magus         MagusSection         `toml:"magus"`
	Choices       ChoicesSection       `toml:"choices"`
	Optimisations OptimisationsSection `toml:"optimisations"`
}

type MagusSection struct {
	Version string `toml:"version"`
	Device  string `toml:"device"`
}

type ChoicesSection struct {
	Terminal string   `toml:"terminal"`
	Browser  string   `toml:"browser"`
	Bundles  []string `toml:"bundles"`
	Theme    string   `toml:"theme"`
}

// OptimisationsSection is the §4 step-4 set, scoped by device.
//
// The split is not cosmetic. A Deck and a Steam Machine want genuinely
// different things — one is a battery-powered handheld held at arm's length,
// the other is mains-powered and three metres from a television — and applying
// either set to the other device is the exact mistake §2 diagnoses. Steps that
// do not apply to the running device report StateNotApplicable rather than
// being silently absent, so `doctor` can say so out loud.
//
// Everything here is userland. The catalogue also carries a Wi-Fi power-save
// tweak and a Btrfs /home conversion; both are deliberately excluded — the
// first needs root, the second is irreversible.
type OptimisationsSection struct {
	// --- any KDE device ---

	// BalooOff disables KDE's file indexer, which can pin the CPU for hours
	// after a fresh install. Reclaims battery on a Deck and quiets the fans.
	BalooOff bool `toml:"baloo_off"`
	// DoubleClick turns off KDE's single-click activation.
	DoubleClick bool `toml:"double_click"`
	// CursorSize in pixels; 0 leaves it alone. A Deck wants a bigger cursor
	// because it is driven by thumbs, a Machine because it is across a room.
	CursorSize int `toml:"cursor_size"`

	// --- any Steam device ---

	// ProtonGE installs the latest GE-Proton into compatibilitytools.d.
	ProtonGE bool `toml:"proton_ge"`

	// --- steam machine only ---

	// HDMIFullRange fixes the washed-out blacks limited-range HDMI causes on a
	// TV. HDMICEC lets the TV wake with the machine. PowerProfile should not be
	// "performance" on something running off a battery.
	HDMIFullRange bool   `toml:"hdmi_full_range"`
	CEC           bool   `toml:"cec"`
	PowerProfile  string `toml:"power_profile"`
}

// Valid choice sets. Anything outside these is a typo, and a typo that silently
// installs nothing is worse than one that stops the run.
var (
	validTerminals = []string{"kitty", "ghostty", "alacritty", "konsole"}
	validBrowsers  = []string{"firefox", "brave", "chrome", "vivaldi"}
	validBundles   = []string{"essentials", "gaming", "creative", "dev", "comms"}
	validProfiles  = []string{"performance", "balanced", "power-saver"}
	validThemes    = []string{"tokyo-night", "gruvbox", "catppuccin", "nord"}
)

// DefaultManifest is the opinionated set — what `magus run --defaults` produces
// and what holding Enter through the wizard produces. Every default here is the
// brief's §4 default.
func DefaultManifest(d Device) Manifest {
	m := Manifest{
		Magus: MagusSection{Version: Version, Device: string(d.Kind)},
		Choices: ChoicesSection{
			Terminal: "kitty",
			Browser:  "firefox",
			Bundles:  []string{"essentials", "gaming"},
			Theme:    "tokyo-night",
		},
		Optimisations: defaultOptimisations(d),
	}
	return m
}

// defaultOptimisations is where the two devices actually diverge. Everything
// KDE-shaped applies to both; the display and power settings are specific to a
// machine plugged into a television.
func defaultOptimisations(d Device) OptimisationsSection {
	o := OptimisationsSection{
		BalooOff:    true,
		DoubleClick: true,
		ProtonGE:    true,
	}
	switch d.Kind {
	case DeviceMachine:
		// Three metres away on a TV: bigger cursor, full-range colour, CEC so
		// the telly wakes with it, and no reason to conserve power.
		o.CursorSize = 48
		o.HDMIFullRange = true
		o.CEC = true
		o.PowerProfile = "performance"
	case DeviceDeck:
		// A handheld driven by thumbs. Its display is its own, so the HDMI
		// fixes are meaningless, and forcing a performance profile on a battery
		// is the handheld-vs-console mistake pointed the other way.
		o.CursorSize = 32
		o.PowerProfile = "balanced"
	default:
		// SteamOS on someone else's hardware, or a plain Linux box. Take the
		// KDE tweaks, leave anything hardware-specific alone.
		o.CursorSize = 0
		o.ProtonGE = false
		o.PowerProfile = "balanced"
	}
	return o
}

// Any reports whether the manifest asks for any optimisation at all. Used
// instead of testing one representative field: CEC is false on a Deck even when
// every optimisation it *can* take is enabled, so keying off it would tell a
// Deck owner they had declined a set they had actually accepted.
func (o OptimisationsSection) Any() bool {
	return o.BalooOff || o.DoubleClick || o.CursorSize != 0 || o.ProtonGE ||
		o.HDMIFullRange || o.CEC || o.PowerProfile == "performance"
}

// noOptimisations is what "skip them" means: nothing changed anywhere, and a
// power profile that is explicitly the neutral one rather than empty.
func noOptimisations() OptimisationsSection {
	return OptimisationsSection{PowerProfile: "balanced"}
}

// LoadManifest reads the manifest from disk. A missing file is reported as
// ErrNoManifest so callers can distinguish "never set up" — which `run` handles
// by launching the wizard — from a genuine read failure.
func LoadManifest(path string) (Manifest, error) {
	var m Manifest
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return m, ErrNoManifest
		}
		return m, err
	}
	if err := toml.Unmarshal(b, &m); err != nil {
		return m, fmt.Errorf("parse %s: %w", path, err)
	}
	return m, nil
}

// ErrNoManifest means this machine has not been set up yet.
var ErrNoManifest = errors.New("no manifest")

// Save writes the manifest atomically. The manifest is the one file whose loss
// costs the user their choices, so a half-written one is never acceptable —
// hence writeFileAtomic rather than os.WriteFile.
func (m Manifest) Save(path string) error {
	var buf bytes.Buffer
	buf.WriteString("# magus manifest — the declarative record of what this machine should look like.\n")
	buf.WriteString("# Edit it and run `magus reconcile` to converge. Safe to keep in version control.\n\n")
	enc := toml.NewEncoder(&buf)
	enc.Indent = ""
	if err := enc.Encode(m); err != nil {
		return err
	}
	return writeFileAtomic(path, buf.Bytes(), 0o644)
}

// Validate reports every problem with the manifest at once rather than the first
// one. A user who mistyped two bundle names should learn both in one run.
func (m Manifest) Validate() error {
	var problems []string

	if m.Magus.Version == "" {
		problems = append(problems, "magus.version is empty")
	}
	if !oneOf(m.Choices.Terminal, validTerminals) {
		problems = append(problems, fmt.Sprintf("choices.terminal %q is not one of %s",
			m.Choices.Terminal, strings.Join(validTerminals, ", ")))
	}
	if !oneOf(m.Choices.Browser, validBrowsers) {
		problems = append(problems, fmt.Sprintf("choices.browser %q is not one of %s",
			m.Choices.Browser, strings.Join(validBrowsers, ", ")))
	}
	for _, b := range m.Choices.Bundles {
		if !oneOf(b, validBundles) {
			problems = append(problems, fmt.Sprintf("choices.bundles contains %q, not one of %s",
				b, strings.Join(validBundles, ", ")))
		}
	}
	// The theme drives nothing yet (v0.4), but it is recorded now — so catch a
	// typo at write time rather than months later when theming lands.
	if t := m.Choices.Theme; t != "" && !oneOf(t, validThemes) {
		problems = append(problems, fmt.Sprintf("choices.theme %q is not one of %s",
			t, strings.Join(validThemes, ", ")))
	}
	if p := m.Optimisations.PowerProfile; p != "" && !oneOf(p, validProfiles) {
		problems = append(problems, fmt.Sprintf("optimisations.power_profile %q is not one of %s",
			p, strings.Join(validProfiles, ", ")))
	}
	// A negative or absurd cursor would be written straight into kcminputrc.
	if c := m.Optimisations.CursorSize; c != 0 && (c < 16 || c > 128) {
		problems = append(problems, fmt.Sprintf(
			"optimisations.cursor_size %d is out of range (16–128, or 0 to leave it alone)", c))
	}

	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("invalid manifest:\n  - %s", strings.Join(problems, "\n  - "))
}

// HasBundle reports whether a bundle was selected.
func (m Manifest) HasBundle(name string) bool { return oneOf(name, m.Choices.Bundles) }

// Migrate brings an older manifest forward to the current schema, returning
// whether anything changed so the caller can decide to re-save.
//
// This is where retiring a step lives: a future version that drops a component
// an earlier one installed clears its field here, and the reconciler's Remove
// pass takes the artifact away. Migrating on read is what keeps every other code
// path able to assume a current-schema manifest.
func (m *Manifest) Migrate() bool {
	changed := false

	// 0.2.0 → 0.3.0: the optimisations section gained the KDE and Proton-GE
	// fields, which did not exist when the manifest was written. Zero values
	// would read as "the user said no" — but they never said anything, so take
	// this release's defaults for the device the manifest records. That is what
	// "converge to the current intended state" means when the intent itself has
	// grown (§3).
	if m.Magus.Version == "0.2.0" {
		def := defaultOptimisations(Device{Kind: DeviceKind(m.Magus.Device)})
		m.Optimisations.BalooOff = def.BalooOff
		m.Optimisations.DoubleClick = def.DoubleClick
		m.Optimisations.CursorSize = def.CursorSize
		m.Optimisations.ProtonGE = def.ProtonGE
		changed = true
	}

	if m.Magus.Version != Version {
		m.Magus.Version = Version
		changed = true
	}
	// A pre-0.2 manifest has no theme; the wizard now always sets one.
	if m.Choices.Theme == "" {
		m.Choices.Theme = "tokyo-night"
		changed = true
	}
	return changed
}

func oneOf(v string, set []string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}
