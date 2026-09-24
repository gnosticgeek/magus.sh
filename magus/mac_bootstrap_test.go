package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMacBootstrapInstallerCancelledByEndBootstrap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "install.sh")
	if err := os.WriteFile(path, []byte("exit 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	downloadCancelled, unlocked := false, false
	m := &macModel{
		bootstrapCancel: func() { downloadCancelled = true },
		bootstrapUnlock: func() { unlocked = true },
		bootstrapFile:   path,
	}
	if cmd := m.runBootstrapInstaller(path); cmd == nil {
		t.Fatal("installer command is nil")
	}
	m.endBootstrap()
	if !downloadCancelled || !unlocked || m.bootstrapCancel != nil || m.bootstrapFile != "" {
		t.Fatalf("bootstrap lifecycle not released: cancelled=%v unlocked=%v", downloadCancelled, unlocked)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("installer file remains: %v", err)
	}
}
