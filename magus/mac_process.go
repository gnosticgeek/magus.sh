package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type commandFunc func(context.Context, string, ...string) (string, error)

// boundedCommandOutput drains both pipes even when the retained output is full.
// Logs never carry terminal control sequences into the parent TUI.
type boundedCommandOutput struct {
	mu   sync.Mutex
	text string
	log  func(string)
}

func (w *boundedCommandOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	s := string(p)
	w.text += s
	if len(w.text) > 4*1024*1024 {
		w.text = w.text[len(w.text)-4*1024*1024:]
	}
	if w.log != nil {
		for _, line := range strings.Split(s, "\n") {
			if strings.TrimSpace(line) != "" {
				w.log(cleanLog(line))
			}
		}
	}
	return len(p), nil
}
func cleanLog(s string) string {
	// ANSI parser removes escape sequences, then strip remaining control chars.
	s = stripTerminal(s)
	if len(s) > 2000 {
		s = s[:2000]
	}
	return s
}
func runMacCommand(ctx context.Context, log func(string), name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = 2 * time.Second
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1", "HOMEBREW_NO_ENV_HINTS=1", "HOMEBREW_NO_ASK=1", "LC_ALL=C")
	w := &boundedCommandOutput{log: log}
	cmd.Stdout = w
	cmd.Stderr = w
	err := cmd.Run()
	if err != nil {
		return w.text, fmt.Errorf("%s: %w: %s", filepath.Base(name), err, cleanLog(w.text))
	}
	return w.text, nil
}
func (c *Context) macCommand(name string, args ...string) (string, error) {
	parent := c.Parent
	if parent == nil {
		parent = context.Background()
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Minute
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	if c.Execute != nil {
		return c.Execute(ctx, name, args...)
	}
	return runMacCommand(ctx, c.OnLog, name, args...)
}
func findBrew() string {
	if p, err := exec.LookPath("brew"); err == nil {
		return p
	}
	for _, p := range []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew"} {
		if st, err := os.Stat(p); err == nil && st.Mode()&0111 != 0 {
			return p
		}
	}
	return ""
}
func (c *Context) brewPath() string {
	if c.Brew != "" {
		return c.Brew
	}
	return findBrew()
}

// flock releases automatically on process exit, including crashes. Leave the
// inode in place: deleting a lock file permits two different locks on one path.
func acquireMacLock(paths Paths) (func(), error) {
	if err := os.MkdirAll(paths.State, 0700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(paths.State, "install.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, fmt.Errorf("another Magus operation is running")
	}
	return func() { syscall.Flock(int(f.Fd()), syscall.LOCK_UN); f.Close() }, nil
}
