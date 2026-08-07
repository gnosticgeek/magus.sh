package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// desktopEntry is a launcher entry we write into ~/.local/share/applications.
//
// The one rule that matters: Exec must be an absolute path. The graphical
// session's PATH does not include ~/.local/bin — that is a .bashrc addition, and
// nothing launching a .desktop sources .bashrc — so a bare command name works
// from a terminal and fails from the launcher (§8).
type desktopEntry struct {
	Name       string
	Comment    string
	Exec       string // absolute path, always
	Icon       string
	Terminal   bool
	Categories []string
}

// render produces the file contents. Values are escaped per the Desktop Entry
// spec: backslash, and the leading-space case that would otherwise be trimmed.
func (e desktopEntry) render() string {
	var b strings.Builder
	b.WriteString("[Desktop Entry]\n")
	b.WriteString("Type=Application\n")
	fmt.Fprintf(&b, "Name=%s\n", escapeDesktopValue(e.Name))
	if e.Comment != "" {
		fmt.Fprintf(&b, "Comment=%s\n", escapeDesktopValue(e.Comment))
	}
	fmt.Fprintf(&b, "Exec=%s\n", escapeDesktopValue(e.Exec))
	if e.Icon != "" {
		fmt.Fprintf(&b, "Icon=%s\n", escapeDesktopValue(e.Icon))
	}
	fmt.Fprintf(&b, "Terminal=%t\n", e.Terminal)
	if len(e.Categories) > 0 {
		fmt.Fprintf(&b, "Categories=%s;\n", strings.Join(e.Categories, ";"))
	}
	// Deliberately no TryExec. A TryExec that misses makes the entry vanish from
	// the launcher silently rather than erroring, which turns a repairable
	// problem into an invisible one — and `magus doctor` is the better place to
	// discover that a binary went missing.
	b.WriteString("X-Magus-Managed=true\n")
	return b.String()
}

// escapeDesktopValue escapes the characters the spec gives meaning to inside a
// value. Paths chosen by a user are exactly where this bites.
func escapeDesktopValue(v string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		"\n", `\n`,
		"\t", `\t`,
		"\r", `\r`,
	)
	return r.Replace(v)
}

// magusManagedMarker distinguishes entries magus wrote from entries the user or
// a package wrote. Nothing without it is ever modified or deleted.
const magusManagedMarker = "X-Magus-Managed=true"

func writeDesktopEntry(c *Context, filename string, e desktopEntry) error {
	path := filepath.Join(c.Paths.Apps, filename)
	if c.DryRun {
		c.Report.Detail("would write %s", path)
		return nil
	}
	if existing, err := os.ReadFile(path); err == nil &&
		!strings.Contains(string(existing), magusManagedMarker) {
		// Someone else's launcher lives here. Overwriting it would be exactly the
		// class of destruction the brief warns about: only remove what you can
		// verify is yours.
		return fmt.Errorf("%s exists and was not written by magus — leaving it alone", path)
	}
	return writeFileAtomic(path, []byte(e.render()), 0o644)
}

// desktopEntryCurrent reports whether the entry exists, is ours, and points at
// the expected binary. Anything else is drift for the owning step to repair.
func desktopEntryCurrent(c *Context, filename, wantExec string) bool {
	b, err := os.ReadFile(filepath.Join(c.Paths.Apps, filename))
	if err != nil {
		return false
	}
	body := string(b)
	if !strings.Contains(body, magusManagedMarker) {
		// Not ours. Treating it as current is the right call — the step must not
		// fight the user for the file, and writeDesktopEntry would refuse anyway.
		return true
	}
	return strings.Contains(body, "Exec="+escapeDesktopValue(wantExec)+"\n")
}
