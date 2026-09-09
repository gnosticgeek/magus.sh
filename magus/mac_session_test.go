package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMacSessionRetryAndResume(t *testing.T) {
	c := macTestContext(t)
	dir := t.TempDir()
	script := `#!/bin/sh
case "$1" in
list) if [ -f "$MAGUS_FIXTURE/installed" ]; then cat "$MAGUS_FIXTURE/installed"; fi ;;
install)
 if [ ! -f "$MAGUS_FIXTURE/attempt" ]; then touch "$MAGUS_FIXTURE/attempt"; echo 'simulated network failure' >&2; exit 1; fi
 echo "$3" >> "$MAGUS_FIXTURE/installed" ;;
*) exit 2 ;;
esac
`
	if err := os.WriteFile(filepath.Join(dir, "brew"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("MAGUS_FIXTURE", dir)
	m := newMacManifest()
	m.Mac.Packages = []string{"git"}
	run := func(retry bool) []macOutcome {
		session := startMacSession(c.Paths, c.Paths.ManifestPath(), m, false, false, time.Second)
		defer session.cancel()
		var final []macOutcome
		timeout := time.After(5 * time.Second)
		for {
			select {
			case e, ok := <-session.events:
				if !ok {
					return final
				}
				if e.Kind == "fatal" {
					t.Fatal(e.Text)
				}
				if e.Kind == "failure" {
					if !retry {
						t.Fatal("rerun failed")
					}
					session.decisions <- "retry"
				}
				if e.Kind == "finished" {
					final = e.Outcomes
				}
			case <-timeout:
				t.Fatal("session hung")
			}
		}
	}
	first := run(true)
	if len(first) != 1 || first[0].Status != "installed" {
		t.Fatalf("first: %+v", first)
	}
	second := run(false)
	if len(second) != 1 || second[0].Status != "already present" {
		t.Fatalf("second: %+v", second)
	}
	if _, err := os.Stat(filepath.Join(c.Paths.State, "last-run.json")); err != nil {
		t.Fatal(err)
	}
}
func TestMacSessionStopLeavesUnfinished(t *testing.T) {
	c := macTestContext(t)
	dir := t.TempDir()
	script := "#!/bin/sh\nif [ \"$1\" = list ]; then exit 0; fi\necho failed >&2\nexit 1\n"
	if err := os.WriteFile(filepath.Join(dir, "brew"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	m := newMacManifest()
	m.Mac.Packages = []string{"git", "jq"}
	session := startMacSession(c.Paths, c.Paths.ManifestPath(), m, false, false, time.Second)
	defer session.cancel()
	var outcomes []macOutcome
	for e := range session.events {
		if e.Kind == "failure" {
			session.decisions <- "stop"
		}
		if e.Kind == "finished" {
			outcomes = e.Outcomes
		}
		if e.Kind == "fatal" {
			t.Fatal(e.Text)
		}
	}
	if len(outcomes) != 2 || outcomes[0].Status != "failed" || outcomes[1].Status != "unfinished" {
		t.Fatalf("unexpected results: %+v", outcomes)
	}
}
