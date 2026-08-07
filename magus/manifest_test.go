package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.toml")

	want := DefaultManifest(Device{Kind: DeviceMachine})
	if err := want.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := LoadManifest(path)
	if err != nil {
		t.Fatal(err)
	}

	if got.Magus != want.Magus {
		t.Errorf("magus section = %+v, want %+v", got.Magus, want.Magus)
	}
	if got.Choices.Terminal != "kitty" || got.Choices.Browser != "firefox" {
		t.Errorf("choices = %+v", got.Choices)
	}
	if strings.Join(got.Choices.Bundles, ",") != "essentials,gaming" {
		t.Errorf("bundles = %v", got.Choices.Bundles)
	}
	if got.Optimisations != want.Optimisations {
		t.Errorf("optimisations = %+v, want %+v", got.Optimisations, want.Optimisations)
	}
}

func TestLoadManifestReportsMissingDistinctly(t *testing.T) {
	_, err := LoadManifest(filepath.Join(t.TempDir(), "nope.toml"))
	if !errors.Is(err, ErrNoManifest) {
		t.Errorf("err = %v, want ErrNoManifest — `run` branches on this", err)
	}
}

// The manifest is the one file whose loss costs the user their choices. A
// crash mid-write must leave the previous one intact, never a truncated file.
func TestSaveIsAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.toml")

	first := DefaultManifest(Device{Kind: DeviceMachine})
	if err := first.Save(path); err != nil {
		t.Fatal(err)
	}
	second := first
	second.Choices.Browser = "brave"
	if err := second.Save(path); err != nil {
		t.Fatal(err)
	}

	got, err := LoadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Choices.Browser != "brave" {
		t.Errorf("browser = %q, want brave", got.Choices.Browser)
	}
	// No temp files may survive a successful write.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), tempMarker) {
			t.Errorf("atomic write left a temp file behind: %s", e.Name())
		}
	}
}

func TestValidateReportsEveryProblemAtOnce(t *testing.T) {
	m := Manifest{
		Magus:   MagusSection{Version: Version},
		Choices: ChoicesSection{Terminal: "xterm", Browser: "netscape", Bundles: []string{"essentials", "nonsense"}},
	}
	err := m.Validate()
	if err == nil {
		t.Fatal("want an error")
	}
	for _, want := range []string{"xterm", "netscape", "nonsense"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %q; got:\n%v", want, err)
		}
	}
}

func TestValidateAcceptsTheDefaults(t *testing.T) {
	for _, d := range []Device{{Kind: DeviceMachine}, {Kind: DeviceDeck}, {Kind: DeviceOther}} {
		if err := DefaultManifest(d).Validate(); err != nil {
			t.Errorf("default manifest for %s is invalid: %v", d.Kind, err)
		}
	}
}

// TV-specific defaults are wrong on a handheld — the same mistake the brief
// diagnoses, pointed the other way.
func TestDeckDefaultsDoNotTakeTheTVOptimisations(t *testing.T) {
	m := DefaultManifest(Device{Kind: DeviceDeck})
	if m.Optimisations.HDMIFullRange || m.Optimisations.CEC {
		t.Error("Deck defaults should not enable the HDMI/CEC optimisations")
	}
	if m.Optimisations.PowerProfile == "performance" {
		t.Error("Deck defaults should not force the performance power profile")
	}
}

func TestMigrateFillsSchemaGaps(t *testing.T) {
	m := Manifest{
		Magus:   MagusSection{Version: "0.1.0"},
		Choices: ChoicesSection{Terminal: "kitty", Browser: "firefox"},
	}
	if !m.Migrate() {
		t.Fatal("Migrate should report a change for a 0.1.0 manifest")
	}
	if m.Magus.Version != Version {
		t.Errorf("version = %q, want %q", m.Magus.Version, Version)
	}
	if m.Choices.Theme == "" {
		t.Error("Migrate should fill in the theme a pre-0.2 manifest lacks")
	}
	if m.Migrate() {
		t.Error("Migrate on an already-current manifest should report no change")
	}
}

// --- device detection ---

func TestClassify(t *testing.T) {
	tests := []struct {
		name              string
		vendor, prod, os_ string
		want              DeviceKind
		confident         bool
	}{
		{"deck lcd", "Valve", "Jupiter", "steamos", DeviceDeck, true},
		{"deck oled", "Valve", "Galileo", "steamos", DeviceDeck, true},
		{"deck case-insensitive", "valve", "jupiter", "steamos", DeviceDeck, true},
		// Unverified: no Steam Machine DMI product name has been confirmed on
		// real hardware (§10), so any other Valve board is a heuristic match.
		{"unknown valve board", "Valve", "Fremont", "steamos", DeviceMachine, false},
		{"steamos on other hardware", "ASUS", "ROG Ally", "steamos", DeviceSteamOS, true},
		{"holo id", "ASUS", "ROG Ally", "holo", DeviceSteamOS, true},
		{"a plain linux box", "Dell", "XPS 13", "arch", DeviceOther, true},
		{"nothing readable", "", "", "", DeviceOther, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classify(tt.vendor, tt.prod, tt.os_)
			if got.Kind != tt.want {
				t.Errorf("Kind = %q, want %q", got.Kind, tt.want)
			}
			if got.Confident != tt.confident {
				t.Errorf("Confident = %v, want %v", got.Confident, tt.confident)
			}
		})
	}
}

func TestDetectDeviceHonoursOverride(t *testing.T) {
	t.Setenv("MAGUS_DEVICE", "steam-machine")
	if got := DetectDevice(); got.Kind != DeviceMachine {
		t.Errorf("Kind = %q, want steam-machine — the headless harness depends on this", got.Kind)
	}
}

// --- the plan ---

// The plan must be identical run to run: ranging over a map would make output
// incomparable between runs and uninstall order non-deterministic.
func TestStepsForIsDeterministic(t *testing.T) {
	m := DefaultManifest(Device{Kind: DeviceMachine})
	m.Choices.Bundles = []string{"comms", "essentials", "gaming", "dev", "creative"}

	first := stepIDs(StepsFor(m))
	for i := 0; i < 20; i++ {
		if got := stepIDs(StepsFor(m)); got != first {
			t.Fatalf("plan differs between runs:\n%s\nvs\n%s", first, got)
		}
	}
}

func TestStepsForPutsFlathubFirst(t *testing.T) {
	steps := StepsFor(DefaultManifest(Device{Kind: DeviceMachine}))
	if len(steps) == 0 || steps[0].ID() != "flathub" {
		t.Fatal("the flathub remote must be ordered before anything that installs a flatpak")
	}
}

func TestStepsForHonoursTheManifest(t *testing.T) {
	m := DefaultManifest(Device{Kind: DeviceMachine})
	m.Choices.Terminal = "konsole"
	m.Choices.Browser = "brave"
	m.Choices.Bundles = []string{"dev"}

	ids := stepIDs(StepsFor(m))
	for _, want := range []string{"terminal:konsole", "browser:brave", "bundle:dev/com.visualstudio.code"} {
		if !strings.Contains(ids, want) {
			t.Errorf("plan is missing %q; got:\n%s", want, ids)
		}
	}
	if strings.Contains(ids, "terminal:kitty") {
		t.Error("plan installs kitty despite the manifest choosing konsole")
	}
	if strings.Contains(ids, "bundle:gaming") {
		t.Error("plan includes a bundle the manifest did not select")
	}
}

// Every bundle app id must be unique across the whole catalogue — a duplicate
// would install fine but uninstall twice, and the second would look like a
// failure.
func TestBundleAppIDsAreUnique(t *testing.T) {
	seen := map[string]string{}
	for _, name := range bundleOrder {
		for _, app := range bundles[name] {
			if prev, dup := seen[app.appID]; dup {
				t.Errorf("%s appears in both %s and %s", app.appID, prev, name)
			}
			seen[app.appID] = name
		}
	}
}

func TestBundleOrderCoversEveryBundle(t *testing.T) {
	if len(bundleOrder) != len(bundles) {
		t.Fatalf("bundleOrder has %d entries but there are %d bundles", len(bundleOrder), len(bundles))
	}
	for _, name := range bundleOrder {
		if _, ok := bundles[name]; !ok {
			t.Errorf("bundleOrder names %q, which is not a bundle", name)
		}
		if !oneOf(name, validBundles) {
			t.Errorf("bundle %q is not accepted by manifest validation", name)
		}
	}
}

func stepIDs(steps []Step) string {
	var b strings.Builder
	for _, s := range steps {
		b.WriteString(s.ID())
		b.WriteByte('\n')
	}
	return b.String()
}
