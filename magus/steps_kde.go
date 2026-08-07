package main

import (
	"fmt"
	"strconv"
	"strings"
)

// KDE settings steps.
//
// Two things make these different from the install steps:
//
//  1. KDE owns the config files, and its daemon rewrites them. So Check never
//     compares file contents — it asks kreadconfig for the *effective* value,
//     which is the honest reading of "derive state from the system" (§3).
//  2. The tool is versioned. SteamOS 3.8 ships Plasma 6 (kreadconfig6), older
//     images ship Plasma 5 (kreadconfig5). The catalogue's commands hardcode 5
//     and are now stale on current hardware, so probe instead of guessing.

// kdeTool returns the Plasma-6 binary if present, else the Plasma-5 one, else
// "" when neither is installed.
func kdeTool(base string) string {
	for _, suffix := range []string{"6", "5"} {
		if have(base + suffix) {
			return base + suffix
		}
	}
	// Some distributions ship an unversioned name.
	if have(base) {
		return base
	}
	return ""
}

// kdeConfigStep sets one key in one KDE config file.
//
// Remove restores KDE's own stock default rather than whatever the user had
// before us. That is a deliberate trade: remembering the prior value would mean
// keeping bookkeeping that can fall out of sync with reality, which is the
// failure mode this whole architecture exists to avoid. The stock value is
// stated in Describe so nothing is a surprise.
type kdeConfigStep struct {
	id      string
	label   string
	file    string // e.g. "kdeglobals"
	group   string // e.g. "KDE"
	key     string // e.g. "SingleClick"
	want    string // the value we set
	stock   string // KDE's default, restored on uninstall
	applies func(Manifest) bool
	// notify runs after the write to make a live session pick it up. Optional:
	// the write itself is what persists.
	notify func(*Context)
}

func (s kdeConfigStep) ID() string { return s.id }
func (s kdeConfigStep) Describe() string {
	return fmt.Sprintf("%s (%s %s/%s = %s; stock is %s)",
		s.label, s.file, s.group, s.key, s.want, s.stock)
}

func (s kdeConfigStep) Check(c *Context) (State, error) {
	if !s.applies(c.Manifest) {
		return StateNotApplicable, nil
	}
	read := kdeTool("kreadconfig")
	if read == "" {
		// Not a KDE machine, or KDE's CLI isn't installed. Not our business.
		return StateNotApplicable, nil
	}
	got, err := c.Output(read, "--file", s.file, "--group", s.group, "--key", s.key)
	if err != nil {
		// kreadconfig exits non-zero when the key is unset, which simply means
		// the stock default is in force — that is "missing", not a failure.
		return StateMissing, nil
	}
	if strings.TrimSpace(got) != s.want {
		return StateMissing, nil
	}
	return StateOK, nil
}

func (s kdeConfigStep) Apply(c *Context) error { return s.write(c, s.want) }

func (s kdeConfigStep) Remove(c *Context) error {
	if !s.applies(c.Manifest) {
		return nil
	}
	return s.write(c, s.stock)
}

func (s kdeConfigStep) write(c *Context, value string) error {
	write := kdeTool("kwriteconfig")
	if write == "" {
		return fmt.Errorf("kwriteconfig is not installed")
	}
	if err := c.Run(write, "--file", s.file, "--group", s.group, "--key", s.key, value); err != nil {
		return err
	}
	if s.notify != nil && !c.DryRun {
		s.notify(c)
	}
	return nil
}

// reconfigureKWin asks a running KWin to re-read its configuration. Best
// effort: with no session there is nothing to tell, and the write still stands
// for next login.
func reconfigureKWin(c *Context) {
	qdbus := kdeTool("qdbus")
	if qdbus == "" {
		return
	}
	if err := c.Run(qdbus, "org.kde.KWin", "/KWin", "reconfigure"); err != nil {
		c.Report.Detail("could not reload KWin live — takes effect at next login")
	}
}

// balooStep disables KDE's file indexer. It is its own step rather than a
// kdeConfigStep because turning it off means both writing the config and
// telling the running service to stop.
type balooStep struct{}

func (balooStep) ID() string { return "optimise:baloo-off" }
func (balooStep) Describe() string {
	return "disable Baloo, KDE's file indexer — it can pin the CPU for hours after a fresh install"
}

func (balooStep) Check(c *Context) (State, error) {
	if !c.Manifest.Optimisations.BalooOff {
		return StateNotApplicable, nil
	}
	read := kdeTool("kreadconfig")
	if read == "" {
		return StateNotApplicable, nil
	}
	got, err := c.Output(read, "--file", "baloofilerc",
		"--group", "Basic Settings", "--key", "Indexing-Enabled")
	if err != nil {
		return StateMissing, nil // unset means indexing is on, i.e. not yet done
	}
	if strings.TrimSpace(got) != "false" {
		return StateMissing, nil
	}
	return StateOK, nil
}

func (balooStep) Apply(c *Context) error {
	write := kdeTool("kwriteconfig")
	if write == "" {
		return fmt.Errorf("kwriteconfig is not installed")
	}
	if err := c.Run(write, "--file", "baloofilerc",
		"--group", "Basic Settings", "--key", "Indexing-Enabled", "false"); err != nil {
		return err
	}
	// Stop the running indexer too. Without a session there is nothing running,
	// so a failure here is not worth failing the step over.
	if ctl := kdeTool("balooctl"); ctl != "" && !c.DryRun {
		if err := c.Run(ctl, "disable"); err != nil {
			c.Report.Detail("balooctl disable did not run — config is set for next login")
		}
	}
	return nil
}

func (balooStep) Remove(c *Context) error {
	write := kdeTool("kwriteconfig")
	if write == "" {
		return nil
	}
	if err := c.Run(write, "--file", "baloofilerc",
		"--group", "Basic Settings", "--key", "Indexing-Enabled", "true"); err != nil {
		return err
	}
	if ctl := kdeTool("balooctl"); ctl != "" && !c.DryRun {
		_ = c.Run(ctl, "enable")
	}
	return nil
}

// kdeSteps returns the KDE tweaks the manifest asks for. Both devices get the
// same set; only the cursor size differs, and that is data rather than a branch.
func kdeSteps(m Manifest) []Step {
	var out []Step

	if m.Optimisations.DoubleClick {
		out = append(out, kdeConfigStep{
			id:      "optimise:double-click",
			label:   "double-click to open, not single",
			file:    "kdeglobals",
			group:   "KDE",
			key:     "SingleClick",
			want:    "false",
			stock:   "true",
			applies: func(m Manifest) bool { return m.Optimisations.DoubleClick },
			notify:  reconfigureKWin,
		})
	}
	if m.Optimisations.BalooOff {
		out = append(out, balooStep{})
	}
	if n := m.Optimisations.CursorSize; n != 0 {
		out = append(out, kdeConfigStep{
			id:      "optimise:cursor-size",
			label:   fmt.Sprintf("cursor size %dpx", n),
			file:    "kcminputrc",
			group:   "Mouse",
			key:     "cursorSize",
			want:    strconv.Itoa(n),
			stock:   "24",
			applies: func(m Manifest) bool { return m.Optimisations.CursorSize != 0 },
			notify:  reconfigureKWin,
		})
	}
	return out
}
