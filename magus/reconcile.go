package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// State is what a step's Check found on the filesystem. It is deliberately
// derived from artifacts and never from a record of what magus did last time —
// on an OS that periodically replaces parts of your install behind your back,
// a "what I installed" file is a lie waiting to happen (§3).
type State int

const (
	// StateUnknown means the probe itself failed — we could not determine the
	// state, which is not the same as knowing it is wrong.
	StateUnknown State = iota
	// StateOK means the artifact exists and is correct. Apply is skipped.
	StateOK
	// StateMissing means the artifact is absent. This is the fresh-install case
	// and equally the post-atomic-update case.
	StateMissing
	// StateDrifted means the artifact exists but no longer matches intent —
	// a stale symlink, a .desktop with the wrong Exec.
	StateDrifted
	// StateNotApplicable means this step does not apply to this machine or this
	// manifest. Not a failure; it is simply not our business.
	StateNotApplicable
)

func (s State) String() string {
	switch s {
	case StateOK:
		return "ok"
	case StateMissing:
		return "missing"
	case StateDrifted:
		return "drifted"
	case StateNotApplicable:
		return "n/a"
	default:
		return "unknown"
	}
}

// NeedsApply reports whether convergence has work to do for this state.
func (s State) NeedsApply() bool { return s == StateMissing || s == StateDrifted }

// Step is one unit of convergence. The contract:
//
//   - Check must not mutate anything, and must derive its answer from the
//     filesystem (or from probing the system), never from stored bookkeeping.
//   - Apply must be safe to run when Check said StateOK — the engine will not
//     call it in that case, but a step that depends on not being called twice
//     is a step that will eventually break.
//   - Remove must be safe to run when the artifact is already gone.
type Step interface {
	// ID is stable and used in output, in --only filters, and in future
	// manifest migrations. Never reuse a retired ID for something else.
	ID() string
	// Describe is the human-readable one-liner shown during a run.
	Describe() string
	// Check inspects the machine. Returning StateNotApplicable is how a step
	// declines a device or a manifest it has no business touching.
	Check(*Context) (State, error)
	// Apply converges the machine toward the manifest.
	Apply(*Context) error
	// Remove reverses Apply for `magus uninstall`. A step that genuinely cannot
	// be reversed returns errNotReversible and says why.
	Remove(*Context) error
}

// errNotReversible is returned by Remove for steps that own nothing removable.
var errNotReversible = fmt.Errorf("not reversible")

// explainer lets a step attach a reason to its not-applicable line. Without it
// a skipped step is a bare "n/a", which reads as "magus ignored this" rather
// than "here is why it doesn't apply to you".
//
// It exists so that Check can stay side-effect free: a step that printed its
// own explanation would emit a line and then have the engine print another.
type explainer interface {
	Why() string
}

// whyOf returns a step's reason for being skipped, or "".
func whyOf(s Step) string {
	if e, ok := s.(explainer); ok {
		return e.Why()
	}
	return ""
}

// Context is everything a step is allowed to know about the run. Passing it
// explicitly rather than reaching for globals is what lets the test harness in
// §9 drive real steps against a temp HOME.
type Context struct {
	Manifest Manifest
	Device   Device
	Paths    Paths
	Report   *Reporter

	// DryRun makes Run a no-op that logs the command it would have executed.
	// Check still runs for real — reporting what would change requires knowing
	// what is actually there.
	DryRun bool

	// Timeout bounds any single command. A hung flatpak download should cost
	// one step, not the whole run.
	Timeout time.Duration
}

// Run executes a command, honouring DryRun and Timeout. Steps shell out rather
// than reimplementing flatpak and friends — magus orchestrates package tools, it
// is not one (§12).
func (c *Context) Run(name string, args ...string) error {
	pretty := name + " " + strings.Join(args, " ")
	if c.DryRun {
		c.Report.Detail("would run: %s", pretty)
		return nil
	}
	c.Report.Detail("$ %s", pretty)

	timeout := c.Timeout
	if timeout == 0 {
		timeout = 15 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	// The graphical session's PATH has no ~/.local/bin (§8), and a step may
	// well need a binary an earlier step just installed there. Fix it up front
	// rather than discovering it at the one call site that breaks.
	cmd.Env = append(os.Environ(), "PATH="+c.Paths.Bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("%s: timed out after %s", name, timeout)
		}
		return fmt.Errorf("%s: %w\n%s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Output runs a command and returns its stdout. Unlike Run it executes even
// under DryRun, because it is used for probing rather than mutating.
func (c *Context) Output(name string, args ...string) (string, error) {
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "PATH="+c.Paths.Bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// Result is the outcome of one step in one run.
type Result struct {
	Step    Step
	Before  State
	After   State
	Changed bool
	Err     error
}

// Summary aggregates a run for the closing report and the exit code.
type Summary struct {
	Results []Result
}

func (s Summary) counts() (ok, changed, skipped, failed int) {
	for _, r := range s.Results {
		switch {
		case r.Err != nil:
			failed++
		case r.Before == StateNotApplicable:
			skipped++
		case r.Changed:
			changed++
		default:
			ok++
		}
	}
	return
}

// Failed reports whether any step errored, which drives the process exit code.
func (s Summary) Failed() bool {
	_, _, _, failed := s.counts()
	return failed > 0
}

// Print writes the closing summary.
func (s Summary) Print(r *Reporter) {
	ok, changed, skipped, failed := s.counts()
	r.Section("summary")
	r.OK("%d changed · %d already correct · %d not applicable", changed, ok, skipped)
	if failed > 0 {
		r.Warn("%d failed", failed)
		for _, res := range s.Results {
			if res.Err != nil {
				r.Detail("%s: %v", res.Step.ID(), res.Err)
			}
		}
	}
}

// Reconcile converges the machine to the manifest.
//
// A failing step warns and the run continues. That is deliberate: the user who
// loses their browser install to a network blip should still get their terminal,
// their bundles and their theme, and a re-run costs them nothing because every
// step is idempotent. Preflight is where we die; convergence is where we degrade.
func Reconcile(ctx *Context, steps []Step) Summary {
	var sum Summary

	for _, st := range steps {
		res := Result{Step: st}

		before, err := st.Check(ctx)
		res.Before = before
		if err != nil {
			// A failed probe is not a licence to apply blindly — applying on an
			// unknown state is how an installer clobbers something it did not
			// put there. Report and move on.
			res.Err = fmt.Errorf("check: %w", err)
			ctx.Report.Warn("%s — could not determine state: %v", st.ID(), err)
			sum.Results = append(sum.Results, res)
			continue
		}

		switch before {
		case StateNotApplicable:
			res.After = before
			if why := whyOf(st); why != "" {
				ctx.Report.Detail("%s — %s", st.ID(), why)
			} else {
				ctx.Report.Detail("%s — not applicable", st.ID())
			}
			sum.Results = append(sum.Results, res)
			continue
		case StateOK:
			res.After = before
			ctx.Report.OK("%s — already correct", st.ID())
			sum.Results = append(sum.Results, res)
			continue
		}

		ctx.Report.Section("%s — %s", st.ID(), st.Describe())
		if err := st.Apply(ctx); err != nil {
			res.Err = err
			ctx.Report.Warn("%s failed: %v", st.ID(), err)
			sum.Results = append(sum.Results, res)
			continue
		}
		res.Changed = true

		// Verify by re-checking rather than trusting Apply's return. A step that
		// exits zero without producing its artifact is precisely the failure this
		// architecture exists to catch.
		if ctx.DryRun {
			res.After = StateOK // nothing was actually applied; don't claim drift
		} else {
			after, err := st.Check(ctx)
			res.After = after
			switch {
			case err != nil:
				res.Err = fmt.Errorf("verify: %w", err)
				ctx.Report.Warn("%s applied but could not be verified: %v", st.ID(), err)
			case after != StateOK:
				res.Err = fmt.Errorf("applied but still %s", after)
				ctx.Report.Warn("%s applied but is still %s", st.ID(), after)
			default:
				ctx.Report.OK("%s — done", st.ID())
			}
		}
		sum.Results = append(sum.Results, res)
	}

	return sum
}

// Doctor reports drift without changing anything. It is the command a user runs
// after an atomic update to find out what broke, and the one they are most
// likely to run when they do not trust us yet — so it must never mutate.
func Doctor(ctx *Context, steps []Step) Summary {
	var sum Summary
	for _, st := range steps {
		res := Result{Step: st}
		state, err := st.Check(ctx)
		res.Before, res.After, res.Err = state, state, err

		switch {
		case err != nil:
			ctx.Report.Warn("%-22s unknown — %v", st.ID(), err)
		case state == StateOK:
			ctx.Report.OK("%-22s ok", st.ID())
		case state == StateNotApplicable:
			if why := whyOf(st); why != "" {
				ctx.Report.Detail("%-22s n/a — %s", st.ID(), why)
			} else {
				ctx.Report.Detail("%-22s n/a", st.ID())
			}
		default:
			ctx.Report.Warn("%-22s %s — %s", st.ID(), state, st.Describe())
		}
		sum.Results = append(sum.Results, res)
	}
	return sum
}

// Uninstall reverses the steps, most-recently-applied first. Order matters:
// a step that installed a binary and a step that symlinked it must come apart
// in the opposite order they went together.
func Uninstall(ctx *Context, steps []Step) Summary {
	var sum Summary
	for i := len(steps) - 1; i >= 0; i-- {
		st := steps[i]
		res := Result{Step: st}

		state, err := st.Check(ctx)
		res.Before = state
		if err == nil && (state == StateMissing || state == StateNotApplicable) {
			ctx.Report.Detail("%s — nothing to remove", st.ID())
			res.After = StateMissing
			sum.Results = append(sum.Results, res)
			continue
		}

		ctx.Report.Section("removing %s", st.ID())
		if err := st.Remove(ctx); err != nil {
			if err == errNotReversible {
				ctx.Report.Warn("%s — %s, leaving in place", st.ID(), err)
				sum.Results = append(sum.Results, res)
				continue
			}
			res.Err = err
			ctx.Report.Warn("%s: %v", st.ID(), err)
		} else {
			res.Changed = true
			ctx.Report.OK("%s removed", st.ID())
		}
		sum.Results = append(sum.Results, res)
	}
	return sum
}
