package main

import (
	"errors"
	"flag"
	"fmt"
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
  magus version            print the manifest schema version

flags:
  --defaults               skip the wizard and take every default
  --dry-run                report what would change without changing it
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
		fmt.Printf("magus %s (manifest schema %s)\n", buildVersion, Version)
		return 0

	case "run":
		return cmdRun(rep, paths, device, *manifestPath, *defaults, *dryRun, *timeout)

	case "reconcile":
		return withManifest(rep, paths, device, *manifestPath, *dryRun, *timeout,
			func(c *Context, steps []Step) Summary {
				c.Report.Section("reconciling %d steps on %s", len(steps), device.Describe())
				return Reconcile(c, steps)
			})

	case "doctor":
		return cmdDoctor(rep, paths, device, *manifestPath, *timeout)

	case "uninstall":
		return withManifest(rep, paths, device, *manifestPath, *dryRun, *timeout,
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
func cmdRun(rep *Reporter, paths Paths, device Device, path string, defaults, dryRun bool, timeout time.Duration) int {
	m, err := LoadManifest(path)
	switch {
	case errors.Is(err, ErrNoManifest):
		if !defaults {
			// The wizard needs a terminal to draw on. Without one — piped,
			// cron, a CI run — there is no way to ask, and silently taking
			// defaults the user never agreed to would be worse than stopping.
			if !isTerminal(os.Stdin) {
				rep.Warn("no manifest at %s, and no terminal to run the wizard on", path)
				rep.Detail("run `magus run --defaults` to take the opinionated set")
				return 2
			}
			answered, confirmed, werr := RunWizard(device)
			if werr != nil {
				rep.Die("wizard: %v", werr)
			}
			if !confirmed {
				rep.Warn("cancelled — nothing was written")
				return 130 // conventional exit for "interrupted by the user"
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
			rep.Die("cannot write manifest: %v", err)
		} else {
			rep.OK("wrote %s", path)
		}
	case err != nil:
		rep.Die("%v", err)
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

	if err := m.Validate(); err != nil {
		rep.Die("%v", err)
	}

	ctx := &Context{Manifest: m, Device: device, Paths: paths, Report: rep, DryRun: dryRun, Timeout: timeout}
	steps := StepsFor(m)
	rep.Section("converging %d steps", len(steps))
	sum := Reconcile(ctx, steps)
	sum.Print(rep)
	if sum.Failed() {
		return 1
	}
	return 0
}

// cmdDoctor reports drift and changes nothing. It exits non-zero when it finds
// drift so that it is usable as a check in a script or a systemd unit.
func cmdDoctor(rep *Reporter, paths Paths, device Device, path string, timeout time.Duration) int {
	rep.Section("device")
	rep.OK("%s", device.Describe())
	if !device.Confident {
		rep.Detail("detection is a heuristic here — see STEAM_MACHINE_BRIEF.md §10")
	}
	rep.Detail("dmi vendor=%q product=%q os=%q", device.Vendor, device.Product, device.OSID)

	rep.Section("tooling")
	for _, t := range []string{"flatpak", "curl"} {
		if have(t) {
			rep.OK("%s present", t)
		} else {
			rep.Warn("%s missing — steps that need it will skip", t)
		}
	}

	m, err := LoadManifest(path)
	if errors.Is(err, ErrNoManifest) {
		rep.Section("manifest")
		rep.Warn("no manifest at %s — this machine has not been set up", path)
		return 1
	}
	if err != nil {
		rep.Die("%v", err)
	}

	rep.Section("manifest")
	rep.OK("%s (schema %s, device %s)", path, m.Magus.Version, m.Magus.Device)
	if err := m.Validate(); err != nil {
		rep.Warn("%v", err)
		return 1
	}
	if m.Magus.Device != string(device.Kind) {
		// Worth saying out loud: a manifest copied from another machine, or a
		// device that detects differently than it used to.
		rep.Warn("manifest says device %q but this machine detects as %q",
			m.Magus.Device, device.Kind)
	}

	ctx := &Context{Manifest: m, Device: device, Paths: paths, Report: rep, DryRun: true, Timeout: timeout}
	steps := StepsFor(m)
	rep.Section("steps")
	sum := Doctor(ctx, steps)

	drift := 0
	for _, r := range sum.Results {
		if r.Err != nil || r.After.NeedsApply() {
			drift++
		}
	}
	rep.Section("summary")
	if drift == 0 {
		rep.OK("no drift — %d steps all correct", len(sum.Results))
		return 0
	}
	rep.Warn("%d of %d steps need attention — run `magus reconcile`", drift, len(sum.Results))
	return 1
}

// withManifest is the shared shape of reconcile and uninstall: load, validate,
// build the plan, hand it to an action, print the summary.
func withManifest(rep *Reporter, paths Paths, device Device, path string, dryRun bool,
	timeout time.Duration, action func(*Context, []Step) Summary) int {

	m, err := LoadManifest(path)
	if errors.Is(err, ErrNoManifest) {
		rep.Warn("no manifest at %s", path)
		rep.Detail("run `magus run --defaults` first")
		return 2
	}
	if err != nil {
		rep.Die("%v", err)
	}
	if m.Migrate() && !dryRun {
		if err := m.Save(path); err != nil {
			rep.Warn("could not save migrated manifest: %v", err)
		}
	}
	if err := m.Validate(); err != nil {
		rep.Die("%v", err)
	}

	ctx := &Context{Manifest: m, Device: device, Paths: paths, Report: rep, DryRun: dryRun, Timeout: timeout}
	sum := action(ctx, StepsFor(m))
	sum.Print(rep)
	if sum.Failed() {
		return 1
	}
	return 0
}
