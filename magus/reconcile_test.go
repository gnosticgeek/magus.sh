package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testContext builds a Context rooted at a temp home, so real steps can be
// exercised without touching the developer's own dotfiles. This is the harness
// the brief asks for in §9: every action expressible without the wizard.
func testContext(t *testing.T, m Manifest) *Context {
	t.Helper()
	home := t.TempDir()
	// Clear XDG overrides so the layout is derived from the temp home alone;
	// a developer with XDG_CONFIG_HOME set would otherwise write outside it.
	for _, v := range []string{"XDG_DATA_HOME", "XDG_STATE_HOME", "XDG_CONFIG_HOME"} {
		t.Setenv(v, "")
	}
	paths := newPathsUnder(home)
	if err := paths.EnsureDirs(); err != nil {
		t.Fatal(err)
	}
	return &Context{
		Manifest: m,
		Device:   Device{Kind: DeviceMachine},
		Paths:    paths,
		Report:   &Reporter{Out: io.Discard, Plain: true},
	}
}

// fakeStep is a scriptable Step for exercising the engine itself.
type fakeStep struct {
	id string
	// states is consumed one entry per Check call, so a test can say
	// "missing, then ok" to model a successful apply.
	states    []State
	checkErr  error
	applyErr  error
	removeErr error

	checks  int
	applies int
	removes int
}

func (s *fakeStep) ID() string       { return s.id }
func (s *fakeStep) Describe() string { return s.id + " (fake)" }

func (s *fakeStep) Check(*Context) (State, error) {
	s.checks++
	if s.checkErr != nil {
		return StateUnknown, s.checkErr
	}
	if len(s.states) == 0 {
		return StateOK, nil
	}
	st := s.states[0]
	if len(s.states) > 1 {
		s.states = s.states[1:]
	}
	return st, nil
}

func (s *fakeStep) Apply(*Context) error  { s.applies++; return s.applyErr }
func (s *fakeStep) Remove(*Context) error { s.removes++; return s.removeErr }

func TestReconcileSkipsStepsAlreadyCorrect(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	st := &fakeStep{id: "already-ok", states: []State{StateOK}}

	sum := Reconcile(c, []Step{st})

	if st.applies != 0 {
		t.Errorf("Apply called %d times on an already-correct step, want 0", st.applies)
	}
	if sum.Results[0].Changed {
		t.Error("result marked Changed for a step that needed nothing")
	}
	if sum.Failed() {
		t.Error("summary reports failure for a clean run")
	}
}

func TestReconcileAppliesThenVerifies(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	st := &fakeStep{id: "repairs", states: []State{StateMissing, StateOK}}

	sum := Reconcile(c, []Step{st})

	if st.applies != 1 {
		t.Errorf("Apply called %d times, want 1", st.applies)
	}
	if st.checks != 2 {
		t.Errorf("Check called %d times, want 2 (probe then verify)", st.checks)
	}
	if !sum.Results[0].Changed || sum.Failed() {
		t.Errorf("want a clean change, got changed=%v failed=%v", sum.Results[0].Changed, sum.Failed())
	}
}

// A step whose Apply exits zero without producing its artifact is exactly the
// failure the verify pass exists to catch — an installer that lies is worse than
// one that errors, because the next run trusts it.
func TestReconcileCatchesApplyThatDidNothing(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	st := &fakeStep{id: "lies", states: []State{StateMissing, StateMissing}}

	sum := Reconcile(c, []Step{st})

	if !sum.Failed() {
		t.Fatal("summary should report failure when the artifact is still missing after Apply")
	}
	if !strings.Contains(sum.Results[0].Err.Error(), "still missing") {
		t.Errorf("error should name the surviving state, got %v", sum.Results[0].Err)
	}
}

// Applying on an unknown state is how an installer clobbers something it did not
// put there. A failed probe must not become a licence to write.
func TestReconcileDoesNotApplyWhenCheckFails(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	st := &fakeStep{id: "unprobeable", checkErr: errors.New("permission denied")}

	sum := Reconcile(c, []Step{st})

	if st.applies != 0 {
		t.Errorf("Apply called %d times after a failed Check, want 0", st.applies)
	}
	if !sum.Failed() {
		t.Error("a failed probe should surface as a failure")
	}
}

// One failing step must not cost the user the rest of the run: convergence
// degrades, it does not abort.
func TestReconcileContinuesPastAFailingStep(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	bad := &fakeStep{id: "bad", states: []State{StateMissing}, applyErr: errors.New("network")}
	good := &fakeStep{id: "good", states: []State{StateMissing, StateOK}}

	sum := Reconcile(c, []Step{bad, good})

	if good.applies != 1 {
		t.Error("a step after a failure should still run")
	}
	if !sum.Failed() {
		t.Error("the run should still report failure overall")
	}
	if len(sum.Results) != 2 {
		t.Errorf("got %d results, want 2", len(sum.Results))
	}
}

func TestReconcileSkipsNotApplicable(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	st := &fakeStep{id: "n/a", states: []State{StateNotApplicable}}

	sum := Reconcile(c, []Step{st})

	if st.applies != 0 {
		t.Error("Apply must not run for a not-applicable step")
	}
	if sum.Failed() {
		t.Error("not-applicable is not a failure")
	}
}

// Dry-run must probe honestly but never write — it is the run a user makes
// before they trust us.
func TestDryRunNeverApplies(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	c.DryRun = true

	kitty := kittyStep{}
	if err := kitty.Apply(c); err != nil {
		t.Fatalf("dry-run Apply errored: %v", err)
	}
	for _, p := range []string{
		filepath.Join(c.Paths.Bin, "kitty"),
		filepath.Join(c.Paths.Apps, "kitty.desktop"),
	} {
		if _, err := os.Lstat(p); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("dry run created %s", p)
		}
	}
}

func TestUninstallRunsInReverseOrder(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	var order []string
	mk := func(id string) *recordingStep {
		return &recordingStep{fakeStep: fakeStep{id: id, states: []State{StateOK}}, order: &order}
	}
	a, b := mk("first"), mk("second")

	Uninstall(c, []Step{a, b})

	if len(order) != 2 || order[0] != "second" || order[1] != "first" {
		t.Errorf("removal order = %v, want [second first]", order)
	}
}

type recordingStep struct {
	fakeStep
	order *[]string
}

func (s *recordingStep) Remove(c *Context) error {
	*s.order = append(*s.order, s.id)
	return s.fakeStep.Remove(c)
}

// --- kitty step: state derived from the filesystem, never from bookkeeping ---

// installFakeKitty lays down the artifacts a real install would produce.
func installFakeKitty(t *testing.T, c *Context) {
	t.Helper()
	binDir := filepath.Join(c.Paths.AppDir("kitty"), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"kitty", "kitten"} {
		target := filepath.Join(binDir, name)
		if err := os.WriteFile(target, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := (kittyStep{}).Apply(c); err != nil {
		t.Fatal(err)
	}
}

func TestKittyCheckReportsOKWhenFullyInstalled(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	installFakeKitty(t, c)

	got, err := kittyStep{}.Check(c)
	if err != nil {
		t.Fatal(err)
	}
	if got != StateOK {
		t.Errorf("Check = %v, want ok", got)
	}
}

// The post-atomic-update case: SteamOS replaced the image and the symlink now
// dangles. The step must see drift and repair it, not report success because it
// "installed this last time".
func TestKittyCheckDetectsDriftFromMissingSymlink(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	installFakeKitty(t, c)

	if err := os.Remove(filepath.Join(c.Paths.Bin, "kitten")); err != nil {
		t.Fatal(err)
	}

	got, err := kittyStep{}.Check(c)
	if err != nil {
		t.Fatal(err)
	}
	if got != StateDrifted {
		t.Errorf("Check = %v, want drifted", got)
	}
}

func TestKittyCheckDetectsDriftFromStaleDesktopExec(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	installFakeKitty(t, c)

	entry := filepath.Join(c.Paths.Apps, "kitty.desktop")
	body, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	stale := strings.Replace(string(body), "Exec="+filepath.Join(c.Paths.AppDir("kitty"), "bin", "kitty"),
		"Exec=/usr/bin/kitty", 1)
	if err := os.WriteFile(entry, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}

	got, _ := kittyStep{}.Check(c)
	if got != StateDrifted {
		t.Errorf("Check = %v, want drifted for a .desktop pointing at /usr", got)
	}
}

// Re-applying a correct install must produce the same filesystem — that is what
// "idempotent" has to mean for a tool whose normal case is re-running.
func TestKittyApplyIsIdempotent(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	installFakeKitty(t, c)

	before := snapshot(t, c.Paths.Bin, c.Paths.Apps)
	if err := (kittyStep{}).Apply(c); err != nil {
		t.Fatalf("second Apply errored: %v", err)
	}
	if after := snapshot(t, c.Paths.Bin, c.Paths.Apps); after != before {
		t.Errorf("second Apply changed the filesystem:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// A real binary at ~/.local/bin/kitty is someone else's install. Destroying it
// is not a trade magus makes.
func TestKittyApplyRefusesToClobberARealFile(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	theirs := filepath.Join(c.Paths.Bin, "kitty")
	if err := os.WriteFile(theirs, []byte("their own build"), 0o755); err != nil {
		t.Fatal(err)
	}
	binDir := filepath.Join(c.Paths.AppDir("kitty"), "bin")
	os.MkdirAll(binDir, 0o755)
	os.WriteFile(filepath.Join(binDir, "kitty"), []byte("#!/bin/sh\n"), 0o755)

	if err := (kittyStep{}).Apply(c); err == nil {
		t.Fatal("Apply should refuse to replace a real file")
	}
	body, _ := os.ReadFile(theirs)
	if string(body) != "their own build" {
		t.Error("Apply destroyed a file it did not own")
	}
}

// Uninstall must leave alone anything magus did not put there.
func TestKittyRemoveLeavesForeignSymlinks(t *testing.T) {
	c := testContext(t, DefaultManifest(Device{Kind: DeviceMachine}))
	elsewhere := filepath.Join(c.Paths.Home, "somewhere", "kitty")
	os.MkdirAll(filepath.Dir(elsewhere), 0o755)
	os.WriteFile(elsewhere, []byte("#!/bin/sh\n"), 0o755)
	link := filepath.Join(c.Paths.Bin, "kitty")
	if err := os.Symlink(elsewhere, link); err != nil {
		t.Fatal(err)
	}

	if err := (kittyStep{}).Remove(c); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(link); err != nil {
		t.Error("Remove deleted a symlink pointing outside magus's app dir")
	}
}

// snapshot renders a stable description of the given dirs for comparison.
func snapshot(t *testing.T, dirs ...string) string {
	t.Helper()
	var b strings.Builder
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			p := filepath.Join(dir, e.Name())
			if target, err := os.Readlink(p); err == nil {
				fmt.Fprintf(&b, "%s -> %s\n", p, target)
				continue
			}
			body, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(&b, "%s = %s\n", p, body)
		}
	}
	return b.String()
}
