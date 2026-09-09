package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
)

type macUpdatesDone struct{ err error }

// Homebrew owns the terminal during maintenance, so prompts and download logs
// remain visible. This deliberately does not use the installation basket.
type macUpdateCommand struct {
	items       []macUpgrade
	refresh     bool
	brew        string
	timeout     time.Duration
	in          io.Reader
	out, errOut io.Writer
}

func (c *macUpdateCommand) SetStdin(r io.Reader)  { c.in = r }
func (c *macUpdateCommand) SetStdout(w io.Writer) { c.out = w }
func (c *macUpdateCommand) SetStderr(w io.Writer) { c.errOut = w }
func (c *macUpdateCommand) Run() error {
	parent, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(parent, c.timeout)
	defer cancel()
	var commands [][]string
	if c.refresh {
		commands = append(commands, []string{"update"})
	} else {
		if len(c.items) == 0 {
			return fmt.Errorf("no reviewed updates")
		}
		current, err := readMacUpdates(ctx, c.brew)
		if err != nil {
			return err
		}
		if !sameMacUpdates(c.items, current) {
			return fmt.Errorf("available versions changed; reopen Update all and review again")
		}
		for _, item := range c.items {
			commands = append(commands, []string{"upgrade", "--" + item.Kind, item.ID})
		}
	}
	for i, args := range commands {
		fmt.Fprintf(c.out, "\nMagus: %d/%d · Homebrew %s (Ctrl+C to stop)\n", i+1, len(commands), args)
		cmd := exec.CommandContext(ctx, c.brew, args...)
		// Keep the foreground process group so password prompts can read the
		// terminal and Ctrl+C reaches Homebrew and its children directly.
		cmd.WaitDelay = 2 * time.Second
		cmd.Stdin, cmd.Stdout, cmd.Stderr = c.in, c.out, c.errOut
		cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_INSTALL_CLEANUP=1")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Homebrew %s: %w", args, err)
		}
	}
	return nil
}

func (m *macModel) updateAll() tea.Cmd {
	if m.preview {
		m.notice = "Preview only — no updates were run."
		return nil
	}
	if !m.updateReview.ready || m.updateReview.loading || m.updateReview.err != nil || len(m.updateReview.items) == 0 {
		m.notice = "Check and review available updates first."
		return nil
	}
	brew := findBrew()
	if brew == "" {
		m.notice = "Install Homebrew before checking for updates."
		return nil
	}
	unlock, err := acquireMacLock(m.paths)
	if err != nil {
		m.notice = err.Error()
		return nil
	}
	timeout := m.timeout
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	items := append([]macUpgrade(nil), m.updateReview.items...)
	return tea.Exec(&macUpdateCommand{brew: brew, timeout: timeout, items: items}, func(err error) tea.Msg {
		unlock()
		return macUpdatesDone{err: err}
	})
}
