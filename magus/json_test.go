package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

// decode round-trips the document the way a front end would, so a change that
// breaks JSON encoding fails here rather than in someone's GUI.
func decode(t *testing.T, o *jsonOutput) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	o.emit(&buf, 0)
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	return got
}

func TestJSONCarriesTheThreeVersionsSeparately(t *testing.T) {
	got := decode(t, newJSONOutput("doctor", Device{Kind: DeviceDeck}, "/tmp/m.toml", false))

	if got["schema"].(float64) != jsonSchema {
		t.Errorf("schema = %v, want %d", got["schema"], jsonSchema)
	}
	m := got["magus"].(map[string]any)
	if m["manifest_schema"] != Version {
		t.Errorf("manifest_schema = %v, want %q", m["manifest_schema"], Version)
	}
	// The contract version, the binary version and the manifest schema move for
	// different reasons; a front end needs all three to reason about
	// compatibility, so none may be dropped.
	if _, ok := m["version"]; !ok {
		t.Error("magus.version missing — a front end cannot tell which binary it is talking to")
	}
}

func TestJSONReportsDeviceConfidence(t *testing.T) {
	// An unverified Steam Machine guess must be visible as a guess (§10).
	guess := classify("Valve", "SomeUnknownBoard", "steamos")
	got := decode(t, newJSONOutput("doctor", guess, "/tmp/m.toml", true))

	d := got["device"].(map[string]any)
	if d["kind"] != string(DeviceMachine) {
		t.Errorf("kind = %v, want steam-machine", d["kind"])
	}
	if d["confident"] != false {
		t.Error("confident should be false for a heuristic match, so a GUI can hedge")
	}
}

func TestJSONFlagsTheStepsNeedingAttention(t *testing.T) {
	sum := Summary{Results: []Result{
		{Step: &fakeStep{id: "fine"}, Before: StateOK, After: StateOK},
		{Step: &fakeStep{id: "broken"}, Before: StateMissing, After: StateMissing},
		{Step: &fakeStep{id: "repaired"}, Before: StateMissing, After: StateOK, Changed: true},
		{Step: &fakeStep{id: "skipped"}, Before: StateNotApplicable, After: StateNotApplicable},
		{Step: &fakeStep{id: "errored"}, Before: StateOK, After: StateOK, Err: errors.New("boom")},
	}}

	out := newJSONOutput("reconcile", Device{Kind: DeviceDeck}, "/tmp/m.toml", false).withSummary(sum)
	byID := map[string]jsonStep{}
	for _, s := range out.Steps {
		byID[s.ID] = s
	}

	// A step that was missing and is now ok has been dealt with; only what is
	// still wrong should be flagged, or a GUI lights up after a good run.
	if byID["repaired"].NeedsAttention {
		t.Error("a repaired step should not need attention")
	}
	if !byID["broken"].NeedsAttention {
		t.Error("a still-missing step should need attention")
	}
	if !byID["errored"].NeedsAttention {
		t.Error("an errored step should need attention even when its state reads ok")
	}
	if byID["fine"].NeedsAttention || byID["skipped"].NeedsAttention {
		t.Error("ok and n/a steps should not need attention")
	}
	if out.Summary.NeedsAttention != 2 {
		t.Errorf("summary.needs_attention = %d, want 2", out.Summary.NeedsAttention)
	}
	if out.Summary.Total != 5 {
		t.Errorf("summary.total = %d, want 5", out.Summary.Total)
	}
}

// A skipped step with no explanation reads as "magus ignored this". The reason
// has to reach the front end.
func TestJSONCarriesTheReasonAStepWasSkipped(t *testing.T) {
	step := notYetOptimisation{id: "optimise:cec", label: "HDMI-CEC", reason: "unverified"}
	sum := Summary{Results: []Result{{Step: step, Before: StateNotApplicable, After: StateNotApplicable}}}

	out := newJSONOutput("doctor", Device{Kind: DeviceMachine}, "/tmp/m.toml", true).withSummary(sum)
	if out.Steps[0].Why == "" {
		t.Error("why is empty — a GUI would show an unexplained skipped row")
	}
}

// The whole point of putting the error in the document: a caller parsing stdout
// must not mistake a failure for "nothing to do".
func TestJSONFailureIsInTheDocumentNotJustTheExitCode(t *testing.T) {
	out := newJSONOutput("reconcile", Device{Kind: DeviceDeck}, "/tmp/m.toml", false)
	var buf bytes.Buffer
	code := out.fail(2, "no manifest at %s", "/tmp/m.toml").emit(&buf, 2)

	if code != 2 {
		t.Errorf("emit returned %d, want the code it was given", code)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("a failure must still produce valid JSON: %v", err)
	}
	if got["error"] == nil || got["error"] == "" {
		t.Error("error field is empty on a failed run")
	}
	if got["exit_code"].(float64) != 2 {
		t.Errorf("exit_code = %v, want 2", got["exit_code"])
	}
}

func TestJSONReportsAnInvalidManifestSeparatelyFromAFatalError(t *testing.T) {
	bad := DefaultManifest(Device{Kind: DeviceDeck})
	bad.Choices.Browser = "netscape"

	out := newJSONOutput("reconcile", Device{Kind: DeviceDeck}, "/tmp/m.toml", false).
		withManifest(bad, Device{Kind: DeviceDeck})

	if out.Manifest.Valid {
		t.Fatal("manifest should be invalid")
	}
	// User-fixable, so it belongs on the manifest rather than as a bare fatal.
	if out.Manifest.Invalid == "" {
		t.Error("invalid manifest should say what is wrong")
	}
}

// A manifest carried from another machine is worth surfacing rather than
// silently converging something built for different hardware.
func TestJSONFlagsADeviceMismatch(t *testing.T) {
	m := DefaultManifest(Device{Kind: DeviceMachine})
	out := newJSONOutput("doctor", Device{Kind: DeviceDeck}, "/tmp/m.toml", true).
		withManifest(m, Device{Kind: DeviceDeck})

	if !out.Manifest.DeviceMismatch {
		t.Error("a steam-machine manifest on a Deck should be flagged as a mismatch")
	}
}

func TestJSONStepsAreNeverNullForAnEmptyRun(t *testing.T) {
	// `null` and `[]` are different to iterate over; a front end should not have
	// to guard the empty case.
	got := decode(t, newJSONOutput("doctor", Device{Kind: DeviceDeck}, "/tmp/m.toml", true))
	if _, ok := got["steps"].([]any); !ok {
		t.Errorf("steps = %#v, want an empty array", got["steps"])
	}
}

func TestJSONToolingReportsWhatStepsDependOn(t *testing.T) {
	out := newJSONOutput("doctor", Device{Kind: DeviceDeck}, "/tmp/m.toml", true).
		withTooling("curl", "definitely-not-a-real-binary")

	if len(out.Tooling) != 2 {
		t.Fatalf("got %d tools, want 2", len(out.Tooling))
	}
	if out.Tooling[1].Present {
		t.Error("a missing tool should report present=false, so a GUI can explain the n/a rows")
	}
}
