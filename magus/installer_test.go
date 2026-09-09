package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallerVerifiedAssetsAndFailClosed(t *testing.T) {
	for _, tc := range []struct {
		os, arch, asset string
		bad             bool
	}{
		{"Darwin", "arm64", "magus-darwin-arm64", false}, {"Darwin", "x86_64", "magus-darwin-amd64", false},
		{"Linux", "x86_64", "magus-linux-amd64", false}, {"Linux", "aarch64", "magus-linux-arm64", false},
		{"Darwin", "arm64", "magus-darwin-arm64", true},
	} {
		t.Run(tc.asset+fmt.Sprint(tc.bad), func(t *testing.T) {
			dir := t.TempDir()
			fake := filepath.Join(dir, "tools")
			os.MkdirAll(fake, 0700)
			payload := []byte("#!/bin/sh\necho fixture-magus\n")
			os.WriteFile(filepath.Join(dir, "binary"), payload, 0600)
			checksum := fmt.Sprintf("%x  %s\n", sha256.Sum256(payload), tc.asset)
			if tc.bad {
				checksum = ""
			}
			os.WriteFile(filepath.Join(dir, "checksums"), []byte(checksum), 0600)
			os.WriteFile(filepath.Join(fake, "uname"), []byte("#!/bin/sh\ncase \"$1\" in -s) echo \"$FIXTURE_OS\";; -m) echo \"$FIXTURE_ARCH\";; esac\n"), 0700)
			os.WriteFile(filepath.Join(fake, "curl"), []byte("#!/bin/sh\ncase \"$4\" in */checksums.txt) cp \"$FIXTURE_DIR/checksums\" \"$3\";; */"+tc.asset+") cp \"$FIXTURE_DIR/binary\" \"$3\";; *) exit 1;; esac\n"), 0700)
			bin := filepath.Join(dir, "bin")
			os.MkdirAll(bin, 0700)
			target := filepath.Join(bin, "magus")
			os.WriteFile(target, []byte("previous"), 0700)
			cmd := exec.Command("/bin/sh", "scripts/install.sh")
			cmd.Env = append(os.Environ(), "PATH="+fake+":"+os.Getenv("PATH"), "HOME="+dir, "FIXTURE_DIR="+dir, "FIXTURE_OS="+tc.os, "FIXTURE_ARCH="+tc.arch, "MAGUS_VERSION=vtest", "MAGUS_BIN_DIR="+bin, "MAGUS_NO_PATH=1", "MAGUS_NO_LAUNCH=1")
			out, err := cmd.CombinedOutput()
			got, _ := os.ReadFile(target)
			if tc.bad {
				if err == nil || string(got) != "previous" {
					t.Fatalf("unverified replacement: %v %s", err, out)
				}
			} else if err != nil || string(got) != string(payload) || !strings.Contains(string(out), "checksum verified") {
				t.Fatalf("installer failed: %v %s", err, out)
			}
		})
	}
}
