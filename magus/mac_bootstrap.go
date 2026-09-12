package main

import (
	"context"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
)

// Homebrew bootstrap owns its temporary file, cancellation function, and
// process-wide installation lock as one lifecycle. Every exit path calls
// endBootstrap, making cleanup explicit and idempotent.
func (m *macModel) endBootstrap() {
	if m.bootstrapCancel != nil {
		m.bootstrapCancel()
		m.bootstrapCancel = nil
	}
	if m.bootstrapFile != "" {
		os.Remove(m.bootstrapFile)
		m.bootstrapFile = ""
	}
	if m.bootstrapUnlock != nil {
		m.bootstrapUnlock()
		m.bootstrapUnlock = nil
	}
}
func (m *macModel) bootstrap() tea.Cmd {
	unlock, err := acquireMacLock(m.paths)
	if err != nil {
		m.notice = err.Error()
		return nil
	}
	m.bootstrapUnlock = unlock
	parent, cancel := context.WithCancel(context.Background())
	m.bootstrapCancel = cancel
	m.screen = macScreenBootstrap
	return func() tea.Msg {
		f, err := os.CreateTemp("", "magus-homebrew-*.sh")
		if err != nil {
			return macBootstrap{err: err}
		}
		path := f.Name()
		f.Close()
		ctx, cancel := context.WithTimeout(parent, 2*time.Minute)
		defer cancel()
		_, err = runMacCommand(ctx, nil, "/usr/bin/curl", "--proto", "=https", "--tlsv1.2", "-fsSL", "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh", "-o", path)
		if err != nil {
			os.Remove(path)
			return macBootstrap{err: err}
		}
		return macBootstrap{path: path}
	}
}
