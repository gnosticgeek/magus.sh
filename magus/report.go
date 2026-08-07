package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/mattn/go-isatty"
)

// Reporter is the CLI's entire output vocabulary: four verbs and nothing else,
// matching the installer convention in the brief (§7). Resisting a fifth verb is
// the point — every line of a reconcile run should be classifiable at a glance.
//
// Output goes to stderr so that anything a subcommand prints to stdout stays
// machine-readable and pipeable.
type Reporter struct {
	Out   io.Writer
	Plain bool // no colour: not a TTY, or --plain
}

// NewReporter writes to stderr, colourised only when stderr is a terminal.
func NewReporter() *Reporter {
	return &Reporter{Out: os.Stderr, Plain: !isTerminal(os.Stderr)}
}

// style applies a lipgloss Render func unless colour is off. Render is variadic,
// hence the signature.
func (r *Reporter) style(s string, f func(...string) string) string {
	if r.Plain {
		return s
	}
	return f(s)
}

// Section announces a new phase of work.
func (r *Reporter) Section(format string, a ...any) {
	fmt.Fprintf(r.Out, "\n%s\n", r.style("── "+fmt.Sprintf(format, a...), sBright.Render))
}

// OK reports something that is now in the desired state — whether we changed it
// or found it already correct.
func (r *Reporter) OK(format string, a ...any) {
	fmt.Fprintf(r.Out, "  %s %s\n", r.style("✓", sAccent.Render), fmt.Sprintf(format, a...))
}

// Warn reports a degraded outcome that does not stop the run. Optional
// components warn and return rather than aborting, so one missing dependency
// never costs the user the rest of the reconcile.
func (r *Reporter) Warn(format string, a ...any) {
	fmt.Fprintf(r.Out, "  %s %s\n", r.style("!", sWarn.Render), fmt.Sprintf(format, a...))
}

// Die reports a fatal condition and exits non-zero. Reserved for preflight:
// once convergence has started, a failing step is a Warn and the run continues.
func (r *Reporter) Die(format string, a ...any) {
	fmt.Fprintf(r.Out, "  %s %s\n", r.style("✗", sWarn.Render), fmt.Sprintf(format, a...))
	os.Exit(1)
}

// Detail prints an indented continuation line under the verb above it.
func (r *Reporter) Detail(format string, a ...any) {
	fmt.Fprintf(r.Out, "    %s\n", r.style(fmt.Sprintf(format, a...), sDim.Render))
}

// have reports whether a command exists on PATH.
func have(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// isTerminal reports whether f is an interactive terminal.
//
// Not a ModeCharDevice test: /dev/null is a character device too, so
// `magus run < /dev/null` — which is what cron and CI do — would look
// interactive and the wizard would try to open a TTY that isn't there.
func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
