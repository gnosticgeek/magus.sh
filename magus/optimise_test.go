package main

import (
	"os"
	"strings"
	"testing"
)

func planIDs(m Manifest) string {
	var b strings.Builder
	for _, s := range StepsFor(m) {
		b.WriteString(s.ID())
		b.WriteByte('\n')
	}
	return b.String()
}

// The two devices are the same tool with different opinions. Everything
// KDE-shaped applies to both; only the TV and power settings differ.
func TestDeckAndMachineShareTheDesktopOptimisations(t *testing.T) {
	deck := DefaultManifest(Device{Kind: DeviceDeck})
	machine := DefaultManifest(Device{Kind: DeviceMachine})

	for _, m := range []Manifest{deck, machine} {
		if !m.Optimisations.BalooOff || !m.Optimisations.DoubleClick {
			t.Errorf("both devices should take the KDE tweaks, got %+v", m.Optimisations)
		}
		if !m.Optimisations.ProtonGE {
			t.Error("both Steam devices should take Proton-GE")
		}
		if m.Optimisations.CursorSize == 0 {
			t.Error("both devices want a bigger cursor, for different reasons")
		}
	}
}

func TestOnlyTheMachineTakesTheTVOptimisations(t *testing.T) {
	deck := DefaultManifest(Device{Kind: DeviceDeck}).Optimisations
	machine := DefaultManifest(Device{Kind: DeviceMachine}).Optimisations

	if deck.HDMIFullRange || deck.CEC {
		t.Errorf("a Deck's display is its own — HDMI fixes are meaningless: %+v", deck)
	}
	if deck.PowerProfile != "balanced" {
		t.Errorf("power_profile = %q on a Deck, want balanced — it runs on a battery", deck.PowerProfile)
	}
	if !machine.HDMIFullRange || !machine.CEC || machine.PowerProfile != "performance" {
		t.Errorf("the Machine should take all three: %+v", machine)
	}
}

// A handheld is held at arm's length and a console sits across a room, so the
// same setting wants a different number rather than a different feature.
func TestCursorSizeDiffersByDistanceNotByFeature(t *testing.T) {
	deck := DefaultManifest(Device{Kind: DeviceDeck}).Optimisations.CursorSize
	machine := DefaultManifest(Device{Kind: DeviceMachine}).Optimisations.CursorSize
	if !(machine > deck && deck > 24) {
		t.Errorf("want stock(24) < deck(%d) < machine(%d)", deck, machine)
	}
}

func TestPlanIncludesTheDesktopStepsOnBothDevices(t *testing.T) {
	for _, kind := range []DeviceKind{DeviceDeck, DeviceMachine} {
		ids := planIDs(DefaultManifest(Device{Kind: kind}))
		for _, want := range []string{
			"optimise:double-click", "optimise:baloo-off",
			"optimise:cursor-size", "optimise:proton-ge",
		} {
			if !strings.Contains(ids, want) {
				t.Errorf("%s plan is missing %q:\n%s", kind, want, ids)
			}
		}
	}
}

func TestPlanOmitsTVStepsOnADeck(t *testing.T) {
	ids := planIDs(DefaultManifest(Device{Kind: DeviceDeck}))
	for _, unwanted := range []string{
		"optimise:hdmi-full-range", "optimise:cec", "optimise:power-profile",
	} {
		if strings.Contains(ids, unwanted) {
			t.Errorf("a Deck plan should not contain %q:\n%s", unwanted, ids)
		}
	}
}

func TestPlanIncludesTVStepsOnAMachine(t *testing.T) {
	ids := planIDs(DefaultManifest(Device{Kind: DeviceMachine}))
	for _, want := range []string{
		"optimise:hdmi-full-range", "optimise:cec", "optimise:power-profile",
	} {
		if !strings.Contains(ids, want) {
			t.Errorf("a Machine plan should contain %q:\n%s", want, ids)
		}
	}
}

// Turning an optimisation off in the manifest must remove its step, not leave a
// step that decides at runtime to do nothing — otherwise `doctor` lists work
// the user explicitly declined.
func TestOptimisationsCanBeTurnedOffIndividually(t *testing.T) {
	m := DefaultManifest(Device{Kind: DeviceDeck})
	m.Optimisations.BalooOff = false
	m.Optimisations.CursorSize = 0

	ids := planIDs(m)
	if strings.Contains(ids, "optimise:baloo-off") {
		t.Error("baloo step present despite baloo_off = false")
	}
	if strings.Contains(ids, "optimise:cursor-size") {
		t.Error("cursor step present despite cursor_size = 0")
	}
	if !strings.Contains(ids, "optimise:double-click") {
		t.Error("turning two off should not have removed the third")
	}
}

// Skipping optimisations in the wizard has to clear every field, on both
// devices — a half-cleared set would silently apply some of what was declined.
func TestSkippingOptimisationsLeavesNoSteps(t *testing.T) {
	for _, kind := range []DeviceKind{DeviceDeck, DeviceMachine} {
		w := NewWizard(Device{Kind: kind})
		for w.Current().ID != "optimise" {
			w.Accept()
		}
		w.Cursor = 1 // "Skip them"
		w.Accept()
		holdEnter(w)

		if ids := planIDs(w.M); strings.Contains(ids, "optimise:") {
			t.Errorf("%s: declining optimisations still planned steps:\n%s", kind, ids)
		}
	}
}

func TestSkippingOptimisationsClearsEveryField(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	for w.Current().ID != "optimise" {
		w.Accept()
	}
	w.Cursor = 1
	w.Accept()

	o := w.M.Optimisations
	if o.BalooOff || o.DoubleClick || o.ProtonGE || o.HDMIFullRange || o.CEC || o.CursorSize != 0 {
		t.Errorf("skip should clear everything, got %+v", o)
	}
	if o.PowerProfile != "balanced" {
		t.Errorf("power_profile = %q, want balanced", o.PowerProfile)
	}
}

// --- validation ---

func TestCursorSizeIsRangeChecked(t *testing.T) {
	for _, tc := range []struct {
		size int
		ok   bool
	}{{0, true}, {24, true}, {32, true}, {128, true}, {15, false}, {129, false}, {-1, false}} {
		m := DefaultManifest(Device{Kind: DeviceDeck})
		m.Optimisations.CursorSize = tc.size
		err := m.Validate()
		if tc.ok && err != nil {
			t.Errorf("cursor_size %d should be valid: %v", tc.size, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("cursor_size %d should be rejected", tc.size)
		}
	}
}

// --- migration ---

// A manifest written by v0.2 has none of the optimisation fields. Zero values
// would read as "the user declined", but they were never asked — so migration
// takes this release's defaults for the device the manifest records.
func TestMigrationFillsNewOptimisationsPerDevice(t *testing.T) {
	for _, kind := range []DeviceKind{DeviceDeck, DeviceMachine} {
		old := Manifest{
			Magus:   MagusSection{Version: "0.2.0", Device: string(kind)},
			Choices: ChoicesSection{Terminal: "kitty", Browser: "firefox", Theme: "nord"},
		}
		if !old.Migrate() {
			t.Fatalf("%s: migrating a 0.2.0 manifest should report a change", kind)
		}
		if old.Magus.Version != Version {
			t.Errorf("version = %q, want %q", old.Magus.Version, Version)
		}
		want := defaultOptimisations(Device{Kind: kind})
		got := old.Optimisations
		if got.BalooOff != want.BalooOff || got.DoubleClick != want.DoubleClick ||
			got.CursorSize != want.CursorSize || got.ProtonGE != want.ProtonGE {
			t.Errorf("%s: migrated optimisations = %+v, want the new fields from %+v", kind, got, want)
		}
		if old.Choices.Theme != "nord" {
			t.Error("migration must not overwrite a choice the user made")
		}
		if err := old.Validate(); err != nil {
			t.Errorf("%s: migrated manifest is invalid: %v", kind, err)
		}
	}
}

func TestMigrationIsIdempotent(t *testing.T) {
	m := DefaultManifest(Device{Kind: DeviceDeck})
	if m.Migrate() {
		t.Error("a current manifest should need no migration")
	}
}

// --- the KDE steps themselves ---

func TestKDEStepsAreNotApplicableWithoutKDE(t *testing.T) {
	if kdeTool("kreadconfig") != "" {
		t.Skip("KDE is installed here; this asserts the off-KDE behaviour")
	}
	c := testContext(t, DefaultManifest(Device{Kind: DeviceDeck}))
	for _, s := range kdeSteps(c.Manifest) {
		got, err := s.Check(c)
		if err != nil {
			t.Errorf("%s: Check errored without KDE: %v", s.ID(), err)
		}
		if got != StateNotApplicable {
			t.Errorf("%s: Check = %v without KDE, want n/a", s.ID(), got)
		}
	}
}

// Proton-GE has nothing to extend when Steam isn't installed, which is the
// normal case on a development machine.
func TestProtonGEIsNotApplicableWithoutSteam(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceDeck}))
	got, err := protonGEStep{}.Check(c)
	if err != nil {
		t.Fatal(err)
	}
	if got != StateNotApplicable {
		t.Errorf("Check = %v with no Steam install, want n/a", got)
	}
}

func TestProtonGEDetectsAnExistingInstall(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceDeck}))
	mkdirAll(t, c.Paths.Home+"/.steam/root/compatibilitytools.d/GE-Proton9-20")

	got, err := protonGEStep{}.Check(c)
	if err != nil {
		t.Fatal(err)
	}
	if got != StateOK {
		t.Errorf("Check = %v with GE-Proton present, want ok", got)
	}
}

// Uninstall must take away GE-Proton without touching a compatibility tool
// somebody else put there.
func TestProtonGERemoveLeavesForeignTools(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceDeck}))
	base := c.Paths.Home + "/.steam/root/compatibilitytools.d"
	mkdirAll(t, base+"/GE-Proton9-20")
	mkdirAll(t, base+"/Proton-Sarek")

	if err := (protonGEStep{}).Remove(c); err != nil {
		t.Fatal(err)
	}
	if dirExists(base + "/GE-Proton9-20") {
		t.Error("Remove left GE-Proton behind")
	}
	if !dirExists(base + "/Proton-Sarek") {
		t.Error("Remove deleted a compatibility tool magus did not install")
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
