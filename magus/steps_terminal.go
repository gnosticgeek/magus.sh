package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// kittyStep installs kitty into ~/.local/kitty.app via the official installer,
// then owns the two artifacts that make it usable: symlinks on PATH and a
// launcher entry.
//
// It is one step rather than three because the three are meaningless apart —
// a symlink to a binary that an atomic update removed is drift, not a separate
// concern, and Check has to see all three to answer honestly.
type kittyStep struct{}

func (kittyStep) ID() string       { return "terminal:kitty" }
func (kittyStep) Describe() string { return "install kitty into ~/.local/kitty.app" }

// kittyInstaller is upstream's own install script. It installs entirely into
// $HOME and needs no root, which is why kitty is the default terminal.
const kittyInstaller = "https://sw.kovidgoyal.net/kitty/installer.sh"

func (k kittyStep) binary(c *Context) string {
	return filepath.Join(c.Paths.AppDir("kitty"), "bin", "kitty")
}

// Check derives state entirely from the filesystem: the binary, both symlinks,
// and the desktop entry. Any one of them missing is drift, and Apply is
// idempotent enough to repair whichever it was.
func (k kittyStep) Check(c *Context) (State, error) {
	if !isExecutable(k.binary(c)) {
		return StateMissing, nil
	}
	for _, name := range []string{"kitty", "kitten"} {
		link := filepath.Join(c.Paths.Bin, name)
		target, err := os.Readlink(link)
		if err != nil {
			return StateDrifted, nil // absent, or a real file rather than a link
		}
		want := filepath.Join(c.Paths.AppDir("kitty"), "bin", name)
		if target != want {
			return StateDrifted, nil
		}
	}
	if !desktopEntryCurrent(c, "kitty.desktop", k.binary(c)) {
		return StateDrifted, nil
	}
	return StateOK, nil
}

func (k kittyStep) Apply(c *Context) error {
	if !isExecutable(k.binary(c)) {
		if !have("curl") {
			return fmt.Errorf("curl is required to install kitty")
		}
		// The installer reads the script on stdin; launch=n stops it opening a
		// window on a machine that may not have a display at all.
		if c.DryRun {
			c.Report.Detail("would run: curl -L %s | sh /dev/stdin launch=n", kittyInstaller)
		} else if err := c.Run("sh", "-c",
			fmt.Sprintf("curl -fsSL %q | sh /dev/stdin launch=n", kittyInstaller)); err != nil {
			return err
		}
	}

	for _, name := range []string{"kitty", "kitten"} {
		target := filepath.Join(c.Paths.AppDir("kitty"), "bin", name)
		if err := forceSymlink(c, target, filepath.Join(c.Paths.Bin, name)); err != nil {
			return err
		}
	}

	return writeDesktopEntry(c, "kitty.desktop", desktopEntry{
		Name:       "kitty",
		Comment:    "Fast, feature-rich, GPU-based terminal",
		Exec:       k.binary(c),
		Icon:       filepath.Join(c.Paths.AppDir("kitty"), "share", "icons", "hicolor", "256x256", "apps", "kitty.png"),
		Terminal:   false,
		Categories: []string{"System", "TerminalEmulator"},
	})
}

func (k kittyStep) Remove(c *Context) error {
	for _, name := range []string{"kitty", "kitten"} {
		link := filepath.Join(c.Paths.Bin, name)
		// Only remove links that point into our app dir. A real file at that
		// path, or a link somewhere else, is the user's own and not ours to take.
		if target, err := os.Readlink(link); err == nil &&
			strings.HasPrefix(target, c.Paths.AppDir("kitty")) {
			if err := removePath(c, link); err != nil {
				return err
			}
		}
	}
	if err := removePath(c, filepath.Join(c.Paths.Apps, "kitty.desktop")); err != nil {
		return err
	}
	return removePath(c, c.Paths.AppDir("kitty"))
}

// keepKonsoleStep is the "keep Konsole" choice made explicit. Modelling it as a
// step that does nothing, rather than as an absent step, means `doctor` can say
// out loud that the terminal choice was honoured.
type keepKonsoleStep struct{}

func (keepKonsoleStep) ID() string                    { return "terminal:konsole" }
func (keepKonsoleStep) Describe() string              { return "keep the stock Konsole — nothing to install" }
func (keepKonsoleStep) Check(*Context) (State, error) { return StateOK, nil }
func (keepKonsoleStep) Apply(*Context) error          { return nil }
func (keepKonsoleStep) Remove(*Context) error         { return nil }

// notYetStep stands in for a manifest choice the reconciler cannot honour yet.
// It reports itself rather than being silently absent — a user who chose Ghostty
// and got kitty's absence with no explanation would rightly not trust the tool.
type notYetStep struct {
	id     string
	choice string
	reason string
}

func (s notYetStep) ID() string { return s.id }
func (s notYetStep) Describe() string {
	return fmt.Sprintf("%s is not implemented yet — %s", s.choice, s.reason)
}

func (s notYetStep) Why() string { return "not implemented yet: " + s.reason }

func (s notYetStep) Check(*Context) (State, error) { return StateNotApplicable, nil }
func (s notYetStep) Apply(*Context) error          { return nil }
func (s notYetStep) Remove(*Context) error         { return nil }

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Mode()&0o111 != 0
}

// forceSymlink replaces whatever is at link with a symlink to target. It
// removes an existing symlink but refuses to clobber a real file — that would
// be someone else's binary, and destroying it is not a trade we make.
func forceSymlink(c *Context, target, link string) error {
	if c.DryRun {
		c.Report.Detail("would link %s -> %s", link, target)
		return nil
	}
	if info, err := os.Lstat(link); err == nil {
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("%s exists and is not a symlink — leaving it alone", link)
		}
		if err := os.Remove(link); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return err
	}
	return os.Symlink(target, link)
}

func removePath(c *Context, path string) error {
	if c.DryRun {
		c.Report.Detail("would remove %s", path)
		return nil
	}
	return os.RemoveAll(path)
}
