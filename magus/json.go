package main

import (
	"encoding/json"
	"fmt"
	"io"
)

// The machine-readable contract.
//
// This exists so a GUI never has to compose a shell string. EmuDeck's Electron
// wrapper pipes arbitrary bash from its renderer to `exec`, which turns any
// front-end bug into arbitrary command execution; the lesson taken here — and
// the one omasteam reached independently (§8) — is that every action should be
// a *named* command the backend dispatches. A front end picks a verb and reads
// this; it never builds a command line.
//
// Output goes to stdout while the human log stays on stderr, so the two can be
// captured separately and neither corrupts the other.

// jsonSchema versions this contract, independently of both the binary's release
// tag and the manifest schema. Three numbers is a lot, but they move for
// different reasons and conflating any two makes a support question
// unanswerable. Bump this only on a breaking change to the shape below.
const jsonSchema = 1

type jsonOutput struct {
	Schema  int    `json:"schema"`
	Command string `json:"command"`

	Magus    jsonMagus    `json:"magus"`
	Device   jsonDevice   `json:"device"`
	Tooling  []jsonTool   `json:"tooling,omitempty"`
	Manifest jsonManifest `json:"manifest"`
	Steps    []jsonStep   `json:"steps"`
	Summary  jsonSummary  `json:"summary"`

	// Error is set when the command could not complete. It is reported here
	// rather than only on stderr, because a caller parsing stdout would
	// otherwise see a valid-looking empty result and treat a failure as "the
	// machine has nothing to do".
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exit_code"`
}

type jsonMagus struct {
	// Version is the binary's release tag; ManifestSchema is the manifest
	// format it reads and writes. A GUI checks the latter before trusting a
	// manifest it did not write.
	Version        string `json:"version"`
	ManifestSchema string `json:"manifest_schema"`
	DryRun         bool   `json:"dry_run"`
}

type jsonDevice struct {
	Kind string `json:"kind"`
	// Confident is false when detection fell back to a heuristic — currently
	// any Valve board that is not a known Deck (§10). A front end should
	// present an unconfident guess as a guess.
	Confident bool   `json:"confident"`
	Vendor    string `json:"vendor"`
	Product   string `json:"product"`
	OSID      string `json:"os_id"`
	SteamOS   bool   `json:"steamos"`
}

type jsonTool struct {
	Name    string `json:"name"`
	Present bool   `json:"present"`
}

type jsonManifest struct {
	Path    string `json:"path"`
	Present bool   `json:"present"`
	Valid   bool   `json:"valid"`
	// Invalid carries the validation problems, which are a user-fixable class
	// of error distinct from Error above.
	Invalid string `json:"invalid,omitempty"`
	Schema  string `json:"schema,omitempty"`
	Device  string `json:"device,omitempty"`
	// DeviceMismatch is true when the manifest records a different device than
	// the one detected — a manifest copied between machines.
	DeviceMismatch bool `json:"device_mismatch,omitempty"`
}

type jsonStep struct {
	ID       string `json:"id"`
	Describe string `json:"describe"`
	// State is what Check found before any work; StateAfter is what it found
	// after, and is empty when nothing was applied.
	State      string `json:"state"`
	StateAfter string `json:"state_after,omitempty"`
	// Why explains a not-applicable step, so a front end can show "skipped
	// because…" rather than an unexplained gap.
	Why     string `json:"why,omitempty"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
	// NeedsAttention is the single field a front end should colour on: true
	// when this step is drifted, missing, or errored.
	NeedsAttention bool `json:"needs_attention"`
}

type jsonSummary struct {
	Total          int `json:"total"`
	Changed        int `json:"changed"`
	OK             int `json:"ok"`
	NotApplicable  int `json:"not_applicable"`
	Failed         int `json:"failed"`
	NeedsAttention int `json:"needs_attention"`
}

// newJSONOutput seeds the parts that are known before any work happens, so an
// early failure still produces a document with the device and manifest path in
// it — which is exactly what someone debugging a failure needs.
func newJSONOutput(command string, d Device, manifestPath string, dryRun bool) *jsonOutput {
	return &jsonOutput{
		Schema:  jsonSchema,
		Command: command,
		Magus: jsonMagus{
			Version:        buildVersion,
			ManifestSchema: Version,
			DryRun:         dryRun,
		},
		Device: jsonDevice{
			Kind:      string(d.Kind),
			Confident: d.Confident,
			Vendor:    d.Vendor,
			Product:   d.Product,
			OSID:      d.OSID,
			SteamOS:   d.SteamOS,
		},
		Manifest: jsonManifest{Path: manifestPath},
		Steps:    []jsonStep{},
	}
}

// withTooling records whether the external tools the steps shell out to are
// present. A front end can explain "no flatpak, so the bundles are skipped"
// instead of showing a list of unexplained n/a rows.
func (o *jsonOutput) withTooling(names ...string) *jsonOutput {
	for _, n := range names {
		o.Tooling = append(o.Tooling, jsonTool{Name: n, Present: have(n)})
	}
	return o
}

func (o *jsonOutput) withManifest(m Manifest, detected Device) *jsonOutput {
	o.Manifest.Present = true
	o.Manifest.Schema = m.Magus.Version
	o.Manifest.Device = m.Magus.Device
	o.Manifest.DeviceMismatch = m.Magus.Device != string(detected.Kind)
	if err := m.Validate(); err != nil {
		o.Manifest.Valid = false
		o.Manifest.Invalid = err.Error()
	} else {
		o.Manifest.Valid = true
	}
	return o
}

// withSummary flattens a run into the step list and counts.
func (o *jsonOutput) withSummary(sum Summary) *jsonOutput {
	ok, changed, skipped, failed := sum.counts()
	o.Summary = jsonSummary{
		Total:         len(sum.Results),
		Changed:       changed,
		OK:            ok,
		NotApplicable: skipped,
		Failed:        failed,
	}
	for _, r := range sum.Results {
		s := jsonStep{
			ID:       r.Step.ID(),
			Describe: r.Step.Describe(),
			State:    r.Before.String(),
			Why:      whyOf(r.Step),
			Changed:  r.Changed,
		}
		if r.After != r.Before {
			s.StateAfter = r.After.String()
		}
		if r.Err != nil {
			s.Error = r.Err.Error()
		}
		// The state that matters to a caller is the one left behind: a step
		// that was missing and is now ok needs no attention.
		final := r.After
		if r.After == StateUnknown {
			final = r.Before
		}
		s.NeedsAttention = r.Err != nil || final.NeedsApply()
		if s.NeedsAttention {
			o.Summary.NeedsAttention++
		}
		o.Steps = append(o.Steps, s)
	}
	return o
}

// fail records a fatal error and the exit code it produced.
func (o *jsonOutput) fail(code int, format string, a ...any) *jsonOutput {
	o.Error = fmt.Sprintf(format, a...)
	o.ExitCode = code
	return o
}

// emit writes the document and returns the exit code, so call sites can end
// with `return out.emit(w, code)` and cannot forget to print.
func (o *jsonOutput) emit(w io.Writer, code int) int {
	o.ExitCode = code
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	// A write failure here has nowhere useful to go: stdout is the channel we
	// are failing to write to. The exit code still carries the outcome.
	_ = enc.Encode(o)
	return code
}
