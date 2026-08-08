package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

// buildVersion is the release tag, injected at link time by the release
// workflow (-X main.buildVersion=v0.3.0). A build from source says so rather
// than claiming a version it isn't.
var buildVersion = "dev"

// usage is the help text. It lists exactly the five verbs from the brief (§5) —
// if a sixth ever appears here without appearing there, one of the two is wrong.
const usage = `magus — the first hour of owning a Steam Machine, done for you.

usage:
  magus                    launch the interactive TUI
  magus run                ask the five questions, write a manifest, converge
  magus run --defaults     no questions; write the opinionated manifest and converge
  magus reconcile          converge to the existing manifest, no questions
  magus doctor             report drift and breakage; changes nothing
  magus uninstall          reverse what magus installed
  magus version            print the binary and manifest schema versions

flags:
  --defaults               skip the wizard and take every default
  --dry-run                report what would change without changing it
  --json                   machine-readable result on stdout; log stays on stderr
  --plain                  no colour, for logs and pipes
  --manifest <path>        use a manifest other than ~/.config/magus/manifest.toml
  --timeout <duration>     per-command timeout (default 15m)

exit codes:
  0  everything converged, or doctor found no drift
  1  a step failed, or doctor found drift
  2  usage error
`

// runCLI dispatches the non-TUI verbs. It returns the process exit code rather
// than exiting, so that every path through it is testable.
func runCLI(args []string) int {
	verb := args[0]

	fs := flag.NewFlagSet("magus "+verb, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	var (
		defaults     = fs.Bool("defaults", false, "skip the wizard and take every default")
		dryRun       = fs.Bool("dry-run", false, "report what would change without changing it")
		asJSON       = fs.Bool("json", false, "machine-readable result on stdout")
		plain        = fs.Bool("plain", false, "no colour")
		manifestPath = fs.String("manifest", "", "path to the manifest")
		timeout      = fs.Duration("timeout", 15*time.Minute, "per-command timeout")
	)
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}

	rep := NewReporter()
	if *plain {
		rep.Plain = true
	}
	// In JSON mode the human log is noise on a channel a caller is parsing
	// alongside stdout, so silence it. Everything it would have said is in the
	// document instead.
	if *asJSON {
		rep.Out = io.Discard
	}

	paths, err := NewPaths()
	if err != nil {
		rep.Die("cannot resolve home directory: %v", err)
	}
	if err := paths.EnsureDirs(); err != nil {
		rep.Die("cannot create %s: %v", paths.Config, err)
	}
	// Clear temps left by a previous run that died mid-write, before anything
	// else has a chance to add more.
	paths.SweepTemps()

	if *manifestPath == "" {
		*manifestPath = paths.ManifestPath()
	}
	device := DetectDevice()

	switch verb {
	case "version":
		// Two different numbers, and conflating them would make a support
		// question unanswerable: buildVersion is which binary you are running,
		// Version is the manifest schema it reads and writes.
		if *asJSON {
			return newJSONOutput("version", device, *manifestPath, false).emit(os.Stdout, 0)
		}
		fmt.Printf("magus %s (manifest schema %s)\n", buildVersion, Version)
		return 0

	case "run":
		return cmdRun(rep, paths, device, *manifestPath, *defaults, *dryRun, *asJSON, *timeout)

	case "reconcile":
		return withManifest(rep, paths, device, "reconcile", *manifestPath, *dryRun, *asJSON, *timeout,
			func(c *Context, steps []Step) Summary {
				c.Report.Section("reconciling %d steps on %s", len(steps), device.Describe())
				return Reconcile(c, steps)
			})

	case "doctor":
		return cmdDoctor(rep, paths, device, *manifestPath, *asJSON, *timeout)

	case "uninstall":
		return withManifest(rep, paths, device, "uninstall", *manifestPath, *dryRun, *asJSON, *timeout,
			func(c *Context, steps []Step) Summary {
				c.Report.Section("removing what magus installed")
				return Uninstall(c, steps)
			})

	case "help", "-h", "--help":
		fmt.Print(usage)
		return 0

	default:
		fmt.Fprintf(os.Stderr, "magus: unknown command %q\n\n%s", verb, usage)
		return 2
	}
}

// cmdRun writes a manifest if there is not one already, then converges.
func cmdRun(rep *Reporter, paths Paths, device Device, path string, defaults, dryRun, asJSON bool, timeout time.Duration) int {
	out := newJSONOutput("run", device, path, dryRun).withTooling("flatpak", "curl")
	die := func(code int, format string, a ...any) int {
		if asJSON {
			return out.fail(code, format, a...).emit(os.Stdout, code)
		}
		rep.Die(format, a...)
		return code
	}

	m, err := LoadManifest(path)
	switch {
	case errors.Is(err, ErrNoManifest):
		if !defaults {
			// The wizard needs a terminal to draw on. Without one — piped,
			// cron, a CI run — there is no way to ask, and silently taking
			// defaults the user never agreed to would be worse than stopping.
			if asJSON || !isTerminal(os.Stdin) {
				rep.Warn("no manifest at %s, and no terminal to run the wizard on", path)
				rep.Detail("run `magus run --defaults` to take the opinionated set")
				return die(2, "no manifest at %s, and the wizard needs a terminal — use --defaults", path)
			}
			answered, confirmed, werr := RunWizard(device)
			if werr != nil {
				return die(1, "wizard: %v", werr)
			}
			if !confirmed {
				rep.Warn("cancelled — nothing was written")
				return die(130, "cancelled — nothing was written") // 130: interrupted by the user
			}
			m = answered
			rep.Section("writing your manifest")
			if dryRun {
				rep.Detail("would write %s", path)
			} else if err := m.Save(path); err != nil {
				rep.Die("cannot write manifest: %v", err)
			} else {
				rep.OK("wrote %s", path)
			}
			break
		}
		m = DefaultManifest(device)
		rep.Section("writing a default manifest for %s", device.Describe())
		if dryRun {
			rep.Detail("would write %s", path)
		} else if err := m.Save(path); err != nil {
			return die(1, "cannot write manifest: %v", err)
		} else {
			rep.OK("wrote %s", path)
		}
	case err != nil:
		return die(1, "%v", err)
	default:
		rep.OK("using existing manifest %s", path)
		if m.Migrate() && !dryRun {
			if err := m.Save(path); err != nil {
				rep.Warn("could not save migrated manifest: %v", err)
			} else {
				rep.OK("migrated manifest to schema %s", Version)
			}
		}
	}

	out.withManifest(m, device)
	if !out.Manifest.Valid {
		return die(1, "%s", out.Manifest.Invalid)
	}

	ctx := &Context{Manifest: m, Device: device, Paths: paths, Report: rep, DryRun: dryRun, Timeout: timeout}
	steps := StepsFor(m)
	rep.Section("converging %d steps", len(steps))
	sum := Reconcile(ctx, steps)
	sum.Print(rep)
	out.withSummary(sum)

	code := 0
	if sum.Failed() {
		code = 1
	}
	if asJSON {
		return out.emit(os.Stdout, code)
	}
	return code
}

// cmdDoctor reports drift and changes nothing. It exits non-zero when it finds
// drift so that it is usable as a check in a script or a systemd unit.
func cmdDoctor(rep *Reporter, paths Paths, device Device, path string, asJSON bool, timeout time.Duration) int {
	out := newJSONOutput("doctor", device, path, true).withTooling("flatpak", "curl")

	// die reports a fatal condition through whichever channel the caller is
	// reading. In JSON mode it must NOT exit, or stdout would carry nothing and
	// a parser would see a closed pipe instead of an explanation.
	die := func(code int, format string, a ...any) int {
		if asJSON {
			return out.fail(code, format, a...).emit(os.Stdout, code)
		}
		rep.Die(format, a...)
		return code // unreachable; rep.Die exits
	}

	rep.Section("device")
	rep.OK("%s", device.Describe())
	if !device.Confident {
		rep.Detail("detection is a heuristic here — see STEAM_MACHINE_BRIEF.md §10")
	}
	rep.Detail("dmi vendor=%q product=%q os=%q", device.Vendor, device.Product, device.OSID)

	rep.Section("tooling")
	for _, t := range out.Tooling {
		if t.Present {
			rep.OK("%s present", t.Name)
		} else {
			rep.Warn("%s missing — steps that need it will skip", t.Name)
		}
	}

	m, err := LoadManifest(path)
	if errors.Is(err, ErrNoManifest) {
		rep.Section("manifest")
		rep.Warn("no manifest at %s — this machine has not been set up", path)
		if asJSON {
			return out.emit(os.Stdout, 1)
		}
		return 1
	}
	if err != nil {
		return die(1, "%v", err)
	}
	out.withManifest(m, device)

	rep.Section("manifest")
	rep.OK("%s (schema %s, device %s)", path, m.Magus.Version, m.Magus.Device)
	if !out.Manifest.Valid {
		rep.Warn("%s", out.Manifest.Invalid)
		if asJSON {
			return out.emit(os.Stdout, 1)
		}
		return 1
	}
	if out.Manifest.DeviceMismatch {
		// Worth saying out loud: a manifest copied from another machine, or a
		// device that detects differently than it used to.
		rep.Warn("manifest says device %q but this machine detects as %q",
			m.Magus.Device, device.Kind)
	}

	ctx := &Context{Manifest: m, Device: device, Paths: paths, Report: rep, DryRun: true, Timeout: timeout}
	steps := StepsFor(m)
	rep.Section("steps")
	sum := Doctor(ctx, steps)
	out.withSummary(sum)

	drift := out.Summary.NeedsAttention
	rep.Section("summary")
	if drift == 0 {
		rep.OK("no drift — %d steps all correct", len(sum.Results))
	} else {
		rep.Warn("%d of %d steps need attention — run `magus reconcile`", drift, len(sum.Results))
	}

	code := 0
	if drift > 0 {
		code = 1
	}
	if asJSON {
		return out.emit(os.Stdout, code)
	}
	return code
}

// withManifest is the shared shape of reconcile and uninstall: load, validate,
// build the plan, hand it to an action, print the summary.
func withManifest(rep *Reporter, paths Paths, device Device, command, path string, dryRun, asJSON bool,
	timeout time.Duration, action func(*Context, []Step) Summary) int {

	out := newJSONOutput(command, device, path, dryRun).withTooling("flatpak", "curl")
	die := func(code int, format string, a ...any) int {
		if asJSON {
			return out.fail(code, format, a...).emit(os.Stdout, code)
		}
		rep.Die(format, a...)
		return code
	}

	m, err := LoadManifest(path)
	if errors.Is(err, ErrNoManifest) {
		rep.Warn("no manifest at %s", path)
		rep.Detail("run `magus run --defaults` first")
		return die(2, "no manifest at %s — run `magus run --defaults` first", path)
	}
	if err != nil {
		return die(1, "%v", err)
	}
	if m.Migrate() && !dryRun {
		if err := m.Save(path); err != nil {
			rep.Warn("could not save migrated manifest: %v", err)
		}
	}
	out.withManifest(m, device)
	if !out.Manifest.Valid {
		return die(1, "%s", out.Manifest.Invalid)
	}

	ctx := &Context{Manifest: m, Device: device, Paths: paths, Report: rep, DryRun: dryRun, Timeout: timeout}
	sum := action(ctx, StepsFor(m))
	sum.Print(rep)
	out.withSummary(sum)

	code := 0
	if sum.Failed() {
		code = 1
	}
	if asJSON {
		return out.emit(os.Stdout, code)
	}
	return code
}
