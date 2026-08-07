package main

import (
	"bufio"
	"os"
	"strings"
)

// DeviceKind is the branch point between the Deck path and the Steam Machine
// path. Steps consult it to decide whether they apply at all — a Deck should
// never be told its power profile ought to be "performance".
type DeviceKind string

const (
	DeviceDeck    DeviceKind = "steam-deck"
	DeviceMachine DeviceKind = "steam-machine"
	// DeviceSteamOS is SteamOS on hardware Valve did not make — a handheld from
	// another vendor, or a desktop install. The userland and immutable-filesystem
	// mechanics all still apply; only the hardware-specific steps do not.
	DeviceSteamOS DeviceKind = "steamos"
	// DeviceOther is anything else: a normal Linux box, or a dev machine. magus
	// still runs, but hardware-gated steps skip themselves.
	DeviceOther DeviceKind = "other"
)

// Device is what detection produced, including the raw DMI strings it decided
// from. The raw values are kept because §10 of the brief is explicit that no
// Steam Machine DMI identifiers have been confirmed on real hardware yet — when
// someone runs `magus doctor` on one, these are the fields that settle it.
type Device struct {
	Kind      DeviceKind
	Vendor    string // /sys/class/dmi/id/sys_vendor
	Product   string // /sys/class/dmi/id/product_name
	OSID      string // ID= from /etc/os-release
	SteamOS   bool
	Confident bool // false when we fell back to a heuristic rather than a known match
}

// Known Deck board names. Jupiter is the LCD model, Galileo the OLED refresh.
// Both are stable, long-published values.
var deckProducts = map[string]bool{
	"jupiter":   true,
	"galileo":   true,
	"steamdeck": true,
}

// DetectDevice classifies the machine magus is running on.
//
// MAGUS_DEVICE overrides detection entirely. That is not a debug escape hatch —
// it is how the headless test harness in §9 exercises the Steam Machine path
// from a laptop, and it is the honest fallback if real hardware reports
// something detection does not yet recognise.
func DetectDevice() Device {
	if forced := os.Getenv("MAGUS_DEVICE"); forced != "" {
		return Device{Kind: DeviceKind(forced), Confident: true}
	}
	return classify(
		readDMI("sys_vendor"),
		readDMI("product_name"),
		readOSRelease("ID"),
	)
}

// classify is the pure decision function, split out so tests can drive it with
// DMI strings that no CI machine will ever actually report.
func classify(vendor, product, osID string) Device {
	d := Device{
		Vendor:  vendor,
		Product: product,
		OSID:    osID,
		SteamOS: strings.EqualFold(osID, "steamos") || strings.EqualFold(osID, "holo"),
	}

	isValve := strings.Contains(strings.ToLower(vendor), "valve")
	prod := strings.ToLower(strings.TrimSpace(product))

	switch {
	case isValve && deckProducts[prod]:
		d.Kind = DeviceDeck
		d.Confident = true

	case isValve:
		// Valve hardware that is not a known Deck board. Today that means a
		// Steam Machine, but the DMI product name for one is unverified (§10) —
		// so this is a heuristic, and says so. Anything hardware-specific should
		// still probe for the capability rather than trusting this label.
		d.Kind = DeviceMachine
		d.Confident = false

	case d.SteamOS:
		d.Kind = DeviceSteamOS
		d.Confident = true

	default:
		d.Kind = DeviceOther
		d.Confident = true
	}
	return d
}

// IsValveHardware reports whether hardware-specific steps are even worth
// attempting.
func (d Device) IsValveHardware() bool {
	return d.Kind == DeviceDeck || d.Kind == DeviceMachine
}

// Describe renders the one-line summary shown by `magus doctor`.
func (d Device) Describe() string {
	var b strings.Builder
	b.WriteString(string(d.Kind))
	if d.Product != "" {
		b.WriteString(" (" + strings.TrimSpace(d.Vendor) + " " + strings.TrimSpace(d.Product) + ")")
	}
	if !d.Confident {
		b.WriteString(" — heuristic, unverified")
	}
	return b.String()
}

func readDMI(field string) string {
	b, err := os.ReadFile("/sys/class/dmi/id/" + field)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// readOSRelease pulls a single key out of /etc/os-release, stripping the quoting
// the format permits.
func readOSRelease(key string) string {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return ""
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		name, value, ok := strings.Cut(line, "=")
		if !ok || name != key {
			continue
		}
		return strings.Trim(value, `"'`)
	}
	return ""
}
