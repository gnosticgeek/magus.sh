package main

import (
	"strings"
	"testing"
)

// holdEnter walks the whole wizard pressing only enter, which is the path a
// user takes when they trust the defaults.
func holdEnter(w *Wizard) {
	for i := 0; i < len(w.Questions)+1 && !w.Done; i++ {
		w.Accept()
	}
}

// The central promise of §4: every question has a default strong enough that
// holding enter produces a well-configured machine — and specifically the same
// machine `--defaults` would have produced.
func TestHoldingEnterMatchesDefaults(t *testing.T) {
	d := Device{Kind: DeviceMachine}
	w := NewWizard(d)
	holdEnter(w)

	if !w.Done {
		t.Fatal("holding enter should finish the wizard")
	}
	want := DefaultManifest(d)
	got := w.M

	if got.Choices.Terminal != want.Choices.Terminal ||
		got.Choices.Browser != want.Choices.Browser ||
		got.Choices.Theme != want.Choices.Theme {
		t.Errorf("choices = %+v, want %+v", got.Choices, want.Choices)
	}
	if strings.Join(got.Choices.Bundles, ",") != strings.Join(want.Choices.Bundles, ",") {
		t.Errorf("bundles = %v, want %v", got.Choices.Bundles, want.Choices.Bundles)
	}
	if got.Optimisations != want.Optimisations {
		t.Errorf("optimisations = %+v, want %+v", got.Optimisations, want.Optimisations)
	}
}

// Whatever the user builds must satisfy the same validation a hand-edited file
// does — the wizard must not be able to produce a manifest the reconciler will
// then reject.
func TestEveryReachableAnswerValidates(t *testing.T) {
	d := Device{Kind: DeviceMachine}
	base := NewWizard(d)

	for qi, q := range base.Questions {
		for oi := range q.Options {
			w := NewWizard(d)
			// Walk to question qi, taking defaults on the way.
			for w.Index < qi {
				w.Accept()
			}
			w.Cursor = oi
			if q.Multi {
				w.Toggle()
			}
			holdEnter(w)

			if err := w.M.Validate(); err != nil {
				t.Errorf("q%d=%s opt=%s produced an invalid manifest: %v",
					qi, q.ID, q.Options[oi].ID, err)
			}
		}
	}
}

func TestCursorOpensOnTheCurrentChoice(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})

	// kitty is the default and is first, so this only proves the mechanism if
	// we move off it and come back.
	w.Cursor = 3 // "Keep Konsole"
	w.Accept()   // terminal := konsole, advance to browser
	w.Back()     // back to terminal

	if got := w.Current().Options[w.Cursor].ID; got != "konsole" {
		t.Errorf("cursor opened on %q, want konsole — the question should show what's chosen", got)
	}
}

func TestBackPreservesEarlierAnswers(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	w.Cursor = 1 // Ghostty
	w.Accept()
	w.Cursor = 1 // Brave
	w.Accept()

	w.Back() // to browser
	w.Back() // to terminal

	if w.M.Choices.Terminal != "ghostty" {
		t.Errorf("terminal = %q, want ghostty", w.M.Choices.Terminal)
	}
	if w.M.Choices.Browser != "brave" {
		t.Errorf("browser = %q, want brave — going back must not clear it", w.M.Choices.Browser)
	}
}

func TestBackAtFirstQuestionIsANoOp(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	w.Back()
	if w.Index != 0 {
		t.Errorf("Index = %d after Back at the first question, want 0", w.Index)
	}
}

func TestToggleAddsAndRemovesBundles(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	for w.Current().ID != "bundles" {
		w.Accept()
	}

	// essentials is on by default; cursor 0 is essentials.
	w.Cursor = 0
	w.Toggle()
	if w.M.HasBundle("essentials") {
		t.Error("toggling an enabled bundle should switch it off")
	}
	w.Toggle()
	if !w.M.HasBundle("essentials") {
		t.Error("toggling again should switch it back on")
	}
}

// The manifest must be byte-identical regardless of the order the user clicked
// bundles in, or two machines with the same choices produce different files.
func TestBundleOrderIsCanonicalNotClickOrder(t *testing.T) {
	build := func(order []int) []string {
		w := NewWizard(Device{Kind: DeviceMachine})
		for w.Current().ID != "bundles" {
			w.Accept()
		}
		// Clear the defaults first.
		w.M.Choices.Bundles = nil
		for _, i := range order {
			w.Cursor = i
			w.Toggle()
		}
		return w.M.Choices.Bundles
	}

	a := strings.Join(build([]int{4, 2, 0}), ",") // comms, creative, essentials
	b := strings.Join(build([]int{0, 2, 4}), ",") // essentials, creative, comms
	if a != b {
		t.Errorf("bundle order depends on click order: %q vs %q", a, b)
	}
	if a != "essentials,creative,comms" {
		t.Errorf("bundles = %q, want canonical essentials,creative,comms", a)
	}
}

func TestToggleIsANoOpOnSingleChoiceQuestions(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	before := w.M.Choices.Terminal
	w.Cursor = 2
	w.Toggle() // terminal is not Multi
	if w.M.Choices.Terminal != before {
		t.Errorf("Toggle changed a single-choice answer: %q → %q", before, w.M.Choices.Terminal)
	}
}

func TestOptimiseAnswerDrivesAllThreeFields(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	for w.Current().ID != "optimise" {
		w.Accept()
	}
	w.Cursor = 1 // "Skip optimisations"
	w.Accept()

	o := w.M.Optimisations
	if o.HDMIFullRange || o.CEC {
		t.Errorf("skipping should clear the display optimisations, got %+v", o)
	}
	if o.PowerProfile != "balanced" {
		t.Errorf("power_profile = %q, want balanced when optimisations are skipped", o.PowerProfile)
	}
}

func TestReviewIsReachedOnlyAfterEveryQuestion(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	for i := 0; i < len(w.Questions); i++ {
		if w.AtReview() {
			t.Fatalf("reached review after %d of %d questions", i, len(w.Questions))
		}
		w.Accept()
	}
	if !w.AtReview() {
		t.Error("should be at review once every question is answered")
	}
	if w.Done {
		t.Error("review must be confirmed before the wizard is Done")
	}
	w.Accept()
	if !w.Done {
		t.Error("accepting at review should finish the wizard")
	}
}

// A Deck must not be handed the mains-powered TV defaults, and the wizard has
// to inherit that from DefaultManifest rather than hard-coding its own set.
func TestWizardStartsFromTheDeviceDefaults(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceDeck})
	if w.M.Optimisations.CEC || w.M.Optimisations.HDMIFullRange {
		t.Error("Deck wizard should not start with the TV optimisations on")
	}
	// But it should still open on "yes": a Deck takes the desktop optimisations
	// even though it takes none of the TV ones. Keying this off a single TV
	// field would tell a Deck owner they had declined a set they had accepted.
	q := w.Questions[3] // optimisations
	if !w.Selected(q, "yes") {
		t.Error("a Deck does want the desktop optimisations — the question should open on 'yes'")
	}
}

func TestCursorCannotLeaveTheOptionList(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	for i := 0; i < 20; i++ {
		w.Up()
	}
	if w.Cursor != 0 {
		t.Errorf("Cursor = %d after many Ups, want 0", w.Cursor)
	}
	for i := 0; i < 20; i++ {
		w.Down()
	}
	if want := len(w.Current().Options) - 1; w.Cursor != want {
		t.Errorf("Cursor = %d after many Downs, want %d", w.Cursor, want)
	}
}

// Every option the wizard offers must be one the manifest accepts. Without
// this, adding an option here and forgetting manifest.go ships a wizard that
// builds a manifest the reconciler refuses.
func TestWizardOptionsMatchManifestValidation(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	sets := map[string][]string{
		"terminal": validTerminals,
		"browser":  validBrowsers,
		"bundles":  validBundles,
		"theme":    validThemes,
	}
	for _, q := range w.Questions {
		want, ok := sets[q.ID]
		if !ok {
			continue
		}
		if len(q.Options) != len(want) {
			t.Errorf("%s offers %d options but %d are valid", q.ID, len(q.Options), len(want))
		}
		for _, o := range q.Options {
			if !oneOf(o.ID, want) {
				t.Errorf("%s offers %q, which the manifest does not accept", q.ID, o.ID)
			}
		}
	}
}

// --- rendering -------------------------------------------------------------
// Bubble Tea reads /dev/tty directly, so the program itself can't be driven
// from a pipe. View() is a pure function of the state though, so rendering can
// still be checked without a terminal.

func renderAt(t *testing.T, w *Wizard) string {
	t.Helper()
	return wizardModel{w: w, width: 80, height: 24}.View()
}

func TestQuestionViewShowsPromptOptionsAndProgress(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	out := renderAt(t, w)

	for _, want := range []string{
		"question 1 of 5", "terminal",
		"Which terminal should this machine use?",
		"kitty", "Ghostty", "Alacritty", "Keep Konsole",
		"accept", "quit",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("question view is missing %q\n---\n%s", want, out)
		}
	}
	// The first question has nothing to go back to.
	if strings.Contains(out, "back") {
		t.Error("the first question should not offer esc/back")
	}
}

func TestBundlesViewShowsCheckboxesAndAppCount(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	for w.Current().ID != "bundles" {
		w.Accept()
	}
	out := renderAt(t, w)

	if !strings.Contains(out, "[×]") || !strings.Contains(out, "[ ]") {
		t.Errorf("bundles view should render checkboxes\n---\n%s", out)
	}
	if !strings.Contains(out, "7 apps") {
		t.Errorf("bundles view should show the app count for the defaults\n---\n%s", out)
	}
	if !strings.Contains(out, "toggle") {
		t.Error("bundles view should offer space/toggle")
	}
}

func TestReviewViewShowsChoicesAndStepCount(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	for !w.AtReview() {
		w.Accept()
	}
	out := renderAt(t, w)

	for _, want := range []string{
		"review", "kitty", "firefox", "essentials, gaming", "tokyo-night",
		"manifest.toml", "converge",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("review view is missing %q\n---\n%s", want, out)
		}
	}
}

// The unimplemented questions must say so on screen, not just in the manifest.
func TestUnbuiltStepsAreDisclosedInTheView(t *testing.T) {
	w := NewWizard(Device{Kind: DeviceMachine})
	for w.Current().ID != "optimise" {
		w.Accept()
	}
	if out := renderAt(t, w); !strings.Contains(out, "not built yet") {
		t.Errorf("optimisations view should disclose that it isn't built\n---\n%s", out)
	}
	w.Accept()
	if out := renderAt(t, w); !strings.Contains(out, "not built yet") {
		t.Errorf("theme view should disclose that it isn't built\n---\n%s", out)
	}
}
