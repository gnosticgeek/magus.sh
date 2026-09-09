package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/muesli/termenv"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

type macOutcome struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}
type macEvent struct {
	Kind, Text string
	Index      int
	Outcome    macOutcome
	Outcomes   []macOutcome
}
type macSession struct {
	events    chan macEvent
	decisions chan string
	cancel    context.CancelFunc
}

func outcomeFor(r Result, dry bool) macOutcome {
	o := macOutcome{ID: r.Step.ID(), Name: r.Step.Describe(), Status: "already present"}
	switch {
	case r.Err != nil:
		o.Status = "failed"
		o.Detail = r.Err.Error()
	case r.Before == StateNotApplicable:
		o.Status = "skipped"
		o.Detail = whyOf(r.Step)
	case r.Changed:
		if dry {
			o.Status = "would install"
		} else {
			o.Status = "installed"
		}
	}
	return o
}
func executeMacStep(c *Context, s Step, restore bool) Result {
	if !restore {
		return Reconcile(c, []Step{s}).Results[0]
	}
	r := Result{Step: s}
	r.Before, r.Err = s.Check(c)
	if r.Err != nil {
		return r
	}
	r.Err = s.Remove(c)
	r.Changed = r.Err == nil
	r.After = StateOK
	return r
}
func restoreSteps(c *Context) ([]Step, error) {
	snaps, err := loadSnapshots(c)
	if err != nil {
		return nil, err
	}
	var steps []Step
	for _, s := range macSettings {
		if _, ok := snaps[s.ID]; ok {
			steps = append(steps, preferenceStep{s})
		}
	}
	if _, err := os.Stat(shellBackup(c.Paths)); err == nil {
		steps = append(steps, shellStep{})
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	for _, p := range appConfigPresets {
		step := appConfigStep{IDValue: p.ID}
		if _, err := os.Stat(step.backup(c.Paths)); err == nil {
			steps = append(steps, step)
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return steps, nil
}
func saveOutcomes(paths Paths, out []macOutcome) error {
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(paths.State, "last-run.json"), b, 0600)
}

// Both the TUI worker and headless commands call executeMacStep. Only the TUI
// pauses for a retry decision. The worker owns all mutations until it exits.
func startMacSession(paths Paths, path string, m Manifest, restore, dry bool, timeout time.Duration) *macSession {
	parent, cancel := context.WithCancel(context.Background())
	session := &macSession{make(chan macEvent, 64), make(chan string, 1), cancel}
	go func() {
		defer close(session.events)
		defer cancel()
		send := func(e macEvent) { session.events <- e }
		c := &Context{Paths: paths, Manifest: m, Device: Device{Kind: DeviceMac}, DryRun: dry, Timeout: timeout, Parent: parent, Report: &Reporter{Out: io.Discard, Plain: true}}
		c.OnLog = func(line string) {
			select {
			case session.events <- macEvent{Kind: "log", Text: line}:
			default:
			}
		}
		var unlock func()
		var err error
		if !dry {
			unlock, err = acquireMacLock(paths)
			if err == nil {
				defer unlock()
			}
		}
		if err != nil {
			send(macEvent{Kind: "fatal", Text: err.Error()})
			return
		}
		steps := macSteps(m)
		if restore {
			steps, err = restoreSteps(c)
		}
		if err == nil && !dry {
			err = m.Save(path)
		}
		if err != nil {
			send(macEvent{Kind: "fatal", Text: err.Error()})
			return
		}
		outcomes := make([]macOutcome, len(steps))
		for i, s := range steps {
			outcomes[i] = macOutcome{ID: s.ID(), Name: s.Describe(), Status: "unfinished"}
		}
		persist := func() bool {
			if dry {
				return true
			}
			if err := saveOutcomes(paths, outcomes); err != nil {
				send(macEvent{Kind: "fatal", Text: "Could not save outcomes: " + err.Error()})
				return false
			}
			return true
		}
		if !persist() {
			return
		}
		send(macEvent{Kind: "plan", Outcomes: append([]macOutcome(nil), outcomes...)})
		for i, s := range steps {
			if parent.Err() != nil {
				break
			}
			for {
				send(macEvent{Kind: "active", Index: i, Text: s.Describe()})
				r := executeMacStep(c, s, restore)
				o := outcomeFor(r, dry)
				if restore && r.Err == nil {
					o.Status = "restored"
					if dry {
						o.Status = "would restore"
					}
				}
				if parent.Err() != nil {
					o.Status = "unfinished"
					o.Detail = "Stopped; next run will inspect actual state."
				}
				outcomes[i] = o
				if restore && r.Err == nil && !dry {
					var ids []string
					for _, id := range m.Mac.Settings {
						if "setting:"+id != s.ID() {
							ids = append(ids, id)
						}
					}
					m.Mac.Settings = ids
					m.Mac.AppConfigs = withoutAppConfig(m.Mac.AppConfigs, s.ID())
					if s.ID() == "shell:configure" {
						m.Mac.ModernShell = nil
					}
					if err := m.Save(path); err != nil {
						send(macEvent{Kind: "fatal", Text: err.Error()})
						return
					}
				}
				if !persist() {
					return
				}
				send(macEvent{Kind: "result", Index: i, Outcome: o})
				if r.Err == nil || parent.Err() != nil {
					break
				}
				send(macEvent{Kind: "failure", Index: i, Text: o.Detail})
				choice := "stop"
				select {
				case choice = <-session.decisions:
				case <-parent.Done():
				}
				if choice == "retry" {
					continue
				}
				if choice == "stop" {
					cancel()
				}
				break
			}
		}
		send(macEvent{Kind: "finished", Outcomes: outcomes})
	}()
	return session
}

func runMacCLI(verb string, paths Paths, path string, defaults, dry, asJSON, plain bool, timeout time.Duration) int {
	if path == "" {
		path = paths.ManifestPath()
	}
	d := DetectDevice()
	out := newJSONOutput(verb, d, path, dry || verb == "doctor").withTooling("brew", "curl")
	rep := NewReporter()
	rep.Plain = plain || rep.Plain
	if asJSON {
		rep.Out = io.Discard
	}
	fail := func(code int, err error) int {
		if asJSON {
			return out.fail(code, "%v", err).emit(os.Stdout, code)
		}
		fmt.Fprintln(os.Stderr, "magus:", err)
		return code
	}
	switch verb {
	case "help", "--help", "-h":
		fmt.Print(usage)
		return 0
	case "version":
		if asJSON {
			return out.emit(os.Stdout, 0)
		}
		fmt.Printf("magus %s (manifest schema %s)\n", buildVersion, Version)
		return 0
	case "run", "preview", "reconcile", "doctor", "restore", "uninstall":
	default:
		return fail(2, fmt.Errorf("unknown command %q", verb))
	}
	if verb == "uninstall" {
		return fail(2, fmt.Errorf("Mac packages are left in place; use magus restore to restore recorded settings"))
	}
	preview := verb == "preview"
	if preview && asJSON {
		return fail(2, fmt.Errorf("preview is interactive; use reconcile --dry-run --json"))
	}
	m, err := LoadManifest(path)
	if err == ErrNoManifest && (verb == "run" || preview) {
		m = newMacManifest()
		err = nil
	}
	if err != nil {
		return fail(1, err)
	}
	if err = manifestPlatformCheck(m, "darwin"); err != nil {
		return fail(1, err)
	}
	if err = m.Validate(); err != nil {
		return fail(1, err)
	}
	out.withManifest(m, d)
	if preview || (verb == "run" && !defaults && !asJSON) {
		if plain {
			setTerminalProfile(termenv.Ascii)
		}
		if !isTerminal(os.Stdin) {
			return fail(2, fmt.Errorf("open an interactive terminal; use reconcile for unattended execution"))
		}
		if err := runMacTUI(paths, path, m, preview || dry, timeout); err != nil {
			return fail(1, err)
		}
		return 0
	}
	if runtime.GOOS != "darwin" {
		return fail(2, fmt.Errorf("Mac execution requires macOS"))
	}
	if verb == "run" && !defaults {
		return fail(2, fmt.Errorf("use reconcile for unattended execution or run --defaults for an empty initial selection"))
	}
	parent, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	c := &Context{Paths: paths, Manifest: m, Device: d, Report: rep, DryRun: dry || verb == "doctor", Timeout: timeout, Parent: parent}
	var unlock func()
	if !c.DryRun {
		unlock, err = acquireMacLock(paths)
		if err != nil {
			return fail(1, err)
		}
		defer unlock()
		if err = m.Save(path); err != nil {
			return fail(1, err)
		}
	}
	steps := macSteps(m)
	if verb == "restore" {
		steps, err = restoreSteps(c)
		if err != nil {
			return fail(1, err)
		}
	}
	var sum Summary
	if verb == "doctor" {
		sum = Doctor(c, steps)
	} else {
		var outcomes []macOutcome
		for _, s := range steps {
			if parent.Err() != nil {
				sum.Results = append(sum.Results, Result{Step: s, Err: parent.Err()})
				outcomes = append(outcomes, macOutcome{ID: s.ID(), Name: s.Describe(), Status: "unfinished"})
				continue
			}
			r := executeMacStep(c, s, verb == "restore")
			sum.Results = append(sum.Results, r)
			o := outcomeFor(r, dry)
			if verb == "restore" && r.Err == nil {
				o.Status = "restored"
				if !dry {
					var ids []string
					for _, id := range m.Mac.Settings {
						if "setting:"+id != s.ID() {
							ids = append(ids, id)
						}
					}
					m.Mac.Settings = ids
					m.Mac.AppConfigs = withoutAppConfig(m.Mac.AppConfigs, s.ID())
					if s.ID() == "shell:configure" {
						m.Mac.ModernShell = nil
					}
					if err = m.Save(path); err != nil {
						return fail(1, err)
					}
				}
			}
			outcomes = append(outcomes, o)
		}
		if !dry {
			if err := saveOutcomes(paths, outcomes); err != nil {
				return fail(1, err)
			}
		}
	}
	out.withSummary(sum)
	code := 0
	if sum.Failed() || verb == "doctor" && out.Summary.NeedsAttention > 0 {
		code = 1
	}
	if asJSON {
		return out.emit(os.Stdout, code)
	}
	sum.Print(rep)
	return code
}
