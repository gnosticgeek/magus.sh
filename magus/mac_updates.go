package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
)

type macUpdatesDone struct{ err error }
type macSelfUpdateDone struct{ err error }

type macSelfUpdateChecked struct {
	available  bool
	latest     string
	generation asyncGeneration
}

const magusInstallerURL = "https://magus.sh/install"
const magusLatestReleaseURL = "https://api.github.com/repos/gnosticgeek/magus.sh/releases/latest"

func (m *macModel) checkMagusUpdate() tea.Cmd {
	if m.selfUpdateCancel != nil {
		m.selfUpdateCancel()
	}
	generation := m.selfUpdateGeneration.next()
	ctx, cancel := context.WithCancel(context.Background())
	m.selfUpdateCancel = cancel
	return func() tea.Msg {
		defer cancel()
		checkCtx, timeout := context.WithTimeout(ctx, 10*time.Second)
		defer timeout()
		output, err := runMacCommand(checkCtx, nil, "/usr/bin/curl", "--proto", "=https", "--tlsv1.2", "-fsSL", magusLatestReleaseURL)
		if err != nil {
			return macSelfUpdateChecked{generation: generation}
		}
		latest, ok := parseMagusLatestRelease([]byte(output))
		if !ok {
			return macSelfUpdateChecked{generation: generation}
		}
		return macSelfUpdateChecked{available: magusVersionNewer(latest, buildVersion), latest: latest, generation: generation}
	}
}

// parseMagusLatestRelease keeps the network response boundary small and
// testable. GitHub's /releases/latest endpoint is intentionally the sole
// authority for an offered update.
func parseMagusLatestRelease(data []byte) (string, bool) {
	var release struct {
		TagName string `json:"tag_name"`
	}
	if json.Unmarshal(data, &release) != nil || strings.TrimSpace(release.TagName) == "" {
		return "", false
	}
	return strings.TrimSpace(release.TagName), true
}

func (m *macModel) magUpdateSummary() string {
	if m.magUpdateAvailable {
		return "Update available: " + m.magUpdateLatest
	}
	if m.magUpdateLatest != "" {
		return "Magus is up to date (" + m.magUpdateLatest + ")."
	}
	return "Download and verify the latest Magus release from GitHub."
}

func magusVersionNewer(latest, current string) bool {
	parse := func(version string) ([]int, bool) {
		version = strings.TrimSpace(version)
		if !strings.HasPrefix(version, "v") {
			return nil, false
		}
		version = strings.TrimPrefix(version, "v")
		parts := strings.Split(version, ".")
		if len(parts) != 3 {
			return nil, false
		}
		parsed := make([]int, len(parts))
		for i, part := range parts {
			if part == "" {
				return nil, false
			}
			for _, r := range part {
				if r < '0' || r > '9' {
					return nil, false
				}
			}
			value, err := strconv.Atoi(part)
			if err != nil {
				return nil, false
			}
			parsed[i] = value
		}
		return parsed, true
	}
	newer, okNewer := parse(latest)
	if current == "dev" {
		return okNewer
	}
	running, okRunning := parse(current)
	if !okNewer || !okRunning {
		return false
	}
	for i := range newer {
		if newer[i] != running[i] {
			return newer[i] > running[i]
		}
	}
	return false
}

type macSelfUpdateCommand struct {
	timeout     time.Duration
	in          io.Reader
	out, errOut io.Writer
}

func (c *macSelfUpdateCommand) SetStdin(r io.Reader)  { c.in = r }
func (c *macSelfUpdateCommand) SetStdout(w io.Writer) { c.out = w }
func (c *macSelfUpdateCommand) SetStderr(w io.Writer) { c.errOut = w }
func (c *macSelfUpdateCommand) Run() error {
	parent, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(parent, c.timeout)
	defer cancel()
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate current executable: %w", err)
	}
	download := exec.CommandContext(ctx, "/usr/bin/curl", "--proto", "=https", "--tlsv1.2", "-fsSL", magusInstallerURL)
	var script bytes.Buffer
	download.Stdout, download.Stderr = &script, c.errOut
	if err := download.Run(); err != nil {
		return fmt.Errorf("download installer: %w", err)
	}
	cmd := exec.CommandContext(ctx, "/bin/sh")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = bytes.NewReader(script.Bytes()), c.out, c.errOut
	cmd.Env = append(os.Environ(), "MAGUS_BIN_DIR="+filepath.Dir(executable), "MAGUS_NO_PATH=1", "MAGUS_NO_LAUNCH=1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run verified installer: %w", err)
	}
	return nil
}

func (m *macModel) updateMagus() tea.Cmd {
	if m.preview {
		m.screen = macScreenMenu
		m.notice = "Preview only — Magus was not updated."
		return nil
	}
	timeout := m.timeout
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	return tea.Exec(&macSelfUpdateCommand{timeout: timeout}, func(err error) tea.Msg {
		return macSelfUpdateDone{err: err}
	})
}

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
