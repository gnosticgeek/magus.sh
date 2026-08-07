package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The wizard is a manifest builder and nothing more (§5). It never touches the
// filesystem, never installs anything, and hands a Manifest back to the caller
// — which is what lets `--defaults`, a hand-edited file and the eventual GUI all
// converge through identical code.
//
// The state machine below is deliberately free of any Bubble Tea types so it can
// be driven from a test without a TTY. wizardModel is a thin rendering shell
// over it.

// optionCol is the column the option notes start at. Wide enough for the
// longest name so a note never runs into it.
const optionCol = 18

// Option is one answer to a Question.
type Option struct {
	ID   string
	Name string
	Note string
}

// Question is one of the five decisions from §4.
type Question struct {
	ID     string
	Title  string
	Prompt string
	// Multi means several options can be on at once (the bundles question).
	Multi   bool
	Options []Option
}

// wizardQuestions builds the question set. The option IDs are drawn from the
// same valid* slices the manifest validates against, so a new terminal or
// bundle cannot be offered here without also being accepted there.
func wizardQuestions(d Device) []Question {
	// The optimisations differ per device, so the option's own note has to as
	// well — offering a Deck "full RGB, CEC, performance profile" would be
	// describing a machine it isn't.
	optimiseNote := "the desktop tweaks below"
	if d.Kind == DeviceMachine {
		optimiseNote = "desktop tweaks, plus the TV and power settings"
	}
	return []Question{
		{
			ID: "terminal", Title: "terminal",
			Prompt: "Which terminal should this machine use?",
			Options: []Option{
				{"kitty", "kitty", "GPU-accelerated · installs to ~/.local"},
				{"ghostty", "Ghostty", "newer, fast — not wired up in v0.2"},
				{"alacritty", "Alacritty", "minimal — not wired up in v0.2"},
				{"konsole", "Keep Konsole", "change nothing"},
			},
		},
		{
			ID: "browser", Title: "browser",
			Prompt: "Which browser?",
			Options: []Option{
				{"firefox", "Firefox", "open source · the default"},
				{"brave", "Brave", "chromium, blocks ads by default"},
				{"chrome", "Google Chrome", "chromium, Google account sync"},
				{"vivaldi", "Vivaldi", "chromium, heavily customisable"},
			},
		},
		{
			ID: "bundles", Title: "apps", Multi: true,
			Prompt: "Which bundles?  (space toggles)",
			Options: []Option{
				{"essentials", "Essentials", "archives, editor, media player, images"},
				{"gaming", "Gaming extras", "ProtonUp-Qt, Heroic, Lutris"},
				{"creative", "Creative", "GIMP, Krita, Inkscape, OBS"},
				{"dev", "Dev", "VS Code"},
				{"comms", "Comms", "Discord, Signal, Element"},
			},
		},
		{
			ID: "optimise", Title: "optimisations",
			Prompt: "These aren't really questions. Here's what will be done:",
			Options: []Option{
				{"yes", "Apply them", optimiseNote},
				{"no", "Skip them", "change no settings at all"},
			},
		},
		{
			ID: "theme", Title: "theme",
			Prompt: "Pick a palette. One choice drives everything.",
			Options: []Option{
				{"tokyo-night", "Tokyo Night", "the default"},
				{"gruvbox", "Gruvbox", "warm, high contrast"},
				{"catppuccin", "Catppuccin", "soft pastel"},
				{"nord", "Nord", "cool, muted"},
			},
		},
	}
}

// optimisationSummary describes what the optimisations answer buys on a given
// device, split into what actually runs today and what is only recorded.
//
// It is derived from defaultOptimisations so the screen cannot claim something
// the manifest would not set — the two would otherwise drift the first time a
// default changed.
func optimisationSummary(d Device) (built, pending []string) {
	o := defaultOptimisations(d)
	if o.BalooOff {
		built = append(built, "Disable Baloo                 stops the indexer pinning the CPU")
	}
	if o.DoubleClick {
		built = append(built, "Double-click to open          instead of KDE's single click")
	}
	switch {
	case o.CursorSize >= 40:
		built = append(built, fmt.Sprintf("Cursor size %dpx              you're three metres away", o.CursorSize))
	case o.CursorSize > 0:
		built = append(built, fmt.Sprintf("Cursor size %dpx              easier to track with thumbs", o.CursorSize))
	}
	if o.ProtonGE {
		built = append(built, "Proton-GE                     more games run, fewer tweaks")
	}
	if o.HDMIFullRange {
		pending = append(pending, "HDMI colour range → full RGB  blacks are grey without it")
	}
	if o.CEC {
		pending = append(pending, "HDMI-CEC                      the TV wakes with the machine")
	}
	if o.PowerProfile == "performance" {
		pending = append(pending, "Performance power profile     it's plugged into a wall")
	}
	return built, pending
}

// Wizard is the pure state machine behind the five questions.
type Wizard struct {
	Questions []Question
	Index     int // == len(Questions) means the review screen
	Cursor    int
	M         Manifest
	// Device is what the questions are being asked about. Held here rather than
	// on the rendering model so there is one source of truth — a screen that
	// took its own copy could disagree with the manifest being built.
	Device Device

	// Done is set when the user confirms at the review screen; Cancelled when
	// they quit. Both stop the program, but only Done yields a manifest worth
	// writing.
	Done      bool
	Cancelled bool
}

// NewWizard starts from the device's defaults, so holding enter all the way
// through produces exactly what `--defaults` would have.
func NewWizard(d Device) *Wizard {
	w := &Wizard{Questions: wizardQuestions(d), M: DefaultManifest(d), Device: d}
	w.syncCursor()
	return w
}

// Current returns the question being asked. Only valid when !w.AtReview().
func (w *Wizard) Current() Question { return w.Questions[w.Index] }

// AtReview reports whether every question has been answered.
func (w *Wizard) AtReview() bool { return w.Index >= len(w.Questions) }

// Selected reports whether an option is currently chosen.
func (w *Wizard) Selected(q Question, optID string) bool {
	switch q.ID {
	case "terminal":
		return w.M.Choices.Terminal == optID
	case "browser":
		return w.M.Choices.Browser == optID
	case "bundles":
		return w.M.HasBundle(optID)
	case "theme":
		return w.M.Choices.Theme == optID
	case "optimise":
		return w.M.Optimisations.Any() == (optID == "yes")
	}
	return false
}

// syncCursor puts the cursor on the currently-chosen option, so each question
// opens showing what the default already is rather than always starting at the
// top. For a multi question there is no single choice, so it starts at 0.
func (w *Wizard) syncCursor() {
	w.Cursor = 0
	if w.AtReview() {
		return
	}
	q := w.Current()
	if q.Multi {
		return
	}
	for i, o := range q.Options {
		if w.Selected(q, o.ID) {
			w.Cursor = i
			return
		}
	}
}

func (w *Wizard) Up() {
	if w.AtReview() || w.Cursor == 0 {
		return
	}
	w.Cursor--
}

func (w *Wizard) Down() {
	if w.AtReview() || w.Cursor >= len(w.Current().Options)-1 {
		return
	}
	w.Cursor++
}

// Toggle flips the option under the cursor. Only meaningful on a Multi
// question; a no-op elsewhere so the key can be bound unconditionally.
func (w *Wizard) Toggle() {
	if w.AtReview() || !w.Current().Multi {
		return
	}
	id := w.Current().Options[w.Cursor].ID
	if w.M.HasBundle(id) {
		out := make([]string, 0, len(w.M.Choices.Bundles))
		for _, b := range w.M.Choices.Bundles {
			if b != id {
				out = append(out, b)
			}
		}
		w.M.Choices.Bundles = out
		return
	}
	// Keep bundles in the canonical order rather than click order, so the
	// manifest is stable no matter how the user got there.
	w.M.Choices.Bundles = append(w.M.Choices.Bundles, id)
	sortBundles(w.M.Choices.Bundles)
}

// sortBundles orders in place by the canonical bundleOrder.
func sortBundles(bs []string) {
	rank := make(map[string]int, len(bundleOrder))
	for i, b := range bundleOrder {
		rank[b] = i
	}
	for i := 1; i < len(bs); i++ {
		for j := i; j > 0 && rank[bs[j]] < rank[bs[j-1]]; j-- {
			bs[j], bs[j-1] = bs[j-1], bs[j]
		}
	}
}

// Accept applies the current selection and advances. On the review screen it
// finishes the wizard.
func (w *Wizard) Accept() {
	if w.AtReview() {
		w.Done = true
		return
	}
	q := w.Current()
	if !q.Multi {
		w.apply(q, q.Options[w.Cursor].ID)
	}
	w.Index++
	w.syncCursor()
}

// Back steps to the previous question, or does nothing at the first one.
func (w *Wizard) Back() {
	if w.Index == 0 {
		return
	}
	w.Index--
	w.syncCursor()
}

func (w *Wizard) apply(q Question, optID string) {
	switch q.ID {
	case "terminal":
		w.M.Choices.Terminal = optID
	case "browser":
		w.M.Choices.Browser = optID
	case "theme":
		w.M.Choices.Theme = optID
	case "optimise":
		// All or nothing, and device-appropriate either way. Setting only the
		// TV fields here would leave the KDE tweaks on after the user declined
		// — they'd have said no and had their settings changed anyway.
		if optID == "yes" {
			w.M.Optimisations = defaultOptimisations(w.Device)
		} else {
			w.M.Optimisations = noOptimisations()
		}
	}
}

// AppCount returns how many bundle apps the current selection implies, for the
// footer on the bundles question.
func (w *Wizard) AppCount() int {
	n := 0
	for _, b := range w.M.Choices.Bundles {
		n += len(bundles[b])
	}
	return n
}

// ---------------------------------------------------------------------------
// Bubble Tea shell
// ---------------------------------------------------------------------------

type wizardModel struct {
	w             *Wizard
	width, height int
}

func (m wizardModel) Init() tea.Cmd { return nil }

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.w.Cancelled = true
			return m, tea.Quit
		case "up", "k":
			m.w.Up()
		case "down", "j":
			m.w.Down()
		case " ":
			m.w.Toggle()
		case "enter":
			m.w.Accept()
			if m.w.Done {
				return m, tea.Quit
			}
		case "esc":
			m.w.Back()
		}
	}
	return m, nil
}

func (m wizardModel) View() string {
	if m.w.AtReview() {
		return m.viewReview()
	}
	return m.viewQuestion()
}

func (m wizardModel) viewQuestion() string {
	w := m.w
	q := w.Current()

	header := sDim.Render("── ") +
		sMuted.Render(fmt.Sprintf("question %d of %d", w.Index+1, len(w.Questions))) +
		sDim.Render(" "+strings.Repeat("─", maxInt(2, 34-len(q.Title)))+" ") +
		sBright.Render(q.Title)

	var b strings.Builder
	b.WriteString("  " + sBright.Render(q.Prompt) + "\n\n")

	for i, o := range q.Options {
		focused := i == w.Cursor
		chosen := w.Selected(q, o.ID)

		var marker string
		switch {
		case q.Multi && chosen:
			marker = sAccent.Render("[×]")
		case q.Multi:
			marker = sDim.Render("[ ]")
		case chosen:
			marker = sAccent.Render("●")
		default:
			marker = sDim.Render("○")
		}

		caret := "  "
		if focused {
			caret = sCursor.Render("▸ ")
		}
		name := sMuted.Render(pad(o.Name, optionCol))
		switch {
		case focused:
			name = sBright.Render(pad(o.Name, optionCol))
		case chosen:
			name = sText.Render(pad(o.Name, optionCol))
		}
		b.WriteString(caret + marker + " " + name + " " + sMuted.Render(o.Note) + "\n")
	}

	b.WriteString("\n  " + rule(56) + "\n")
	switch q.ID {
	case "optimise":
		built, pending := optimisationSummary(m.w.Device)
		for _, l := range built {
			b.WriteString("  " + sAccent.Render("✓") + " " + sMuted.Render(l) + "\n")
		}
		for _, l := range pending {
			b.WriteString("  " + sDim.Render("· "+l) + "\n")
		}
		if len(pending) > 0 {
			b.WriteString("\n  " + sDim.Render("· not built yet — the manifest records your answer") + "\n")
		}
	case "bundles":
		n := w.AppCount()
		b.WriteString("  " + sText.Render(fmt.Sprintf("%d apps", n)) +
			sDim.Render(" · ") + sMuted.Render("flatpak, sandboxed, removable") + "\n")
	case "theme":
		b.WriteString("  " + sDim.Render("not built yet — v0.4. The manifest records your answer.") + "\n")
	default:
		b.WriteString("  " + sAccent.Render("default") +
			sDim.Render(" · press enter to accept") + "\n")
	}

	hints := []Hint{
		{Key: "↑↓", Action: "choose", Kind: HintNormal},
	}
	if q.Multi {
		hints = append(hints,
			Hint{Key: "space", Action: "toggle", Kind: HintPrimary},
			Hint{Key: "enter", Action: "continue", Kind: HintNormal})
	} else {
		hints = append(hints, Hint{Key: "enter", Action: "accept", Kind: HintPrimary})
	}
	if w.Index > 0 {
		hints = append(hints, Hint{Key: "esc", Action: "back", Kind: HintNormal})
	}
	hints = append(hints, Hint{Key: "q", Action: "quit", Kind: HintSystem})

	return wrapFrame(m.width, m.height, header, b.String(), "", statusBar(hints))
}

func (m wizardModel) viewReview() string {
	w := m.w
	header := sDim.Render("── ") + sMuted.Render("review") + sDim.Render(" "+strings.Repeat("─", 46))

	var b strings.Builder
	rows := [][2]string{
		{"terminal", w.M.Choices.Terminal},
		{"browser", w.M.Choices.Browser},
		{"bundles", strings.Join(w.M.Choices.Bundles, ", ")},
		{"theme", w.M.Choices.Theme},
		{"optimise", map[bool]string{true: "applied", false: "skipped"}[w.M.Optimisations.CEC]},
	}
	for _, r := range rows {
		v := r[1]
		if v == "" {
			v = "none"
		}
		b.WriteString("  " + sMuted.Render(pad(r[0], 11)) + sText.Render(v) + "\n")
	}

	b.WriteString("\n  " + sMuted.Render("This writes one file:") + "\n")
	b.WriteString("  " + sText.Render("~/.config/magus/manifest.toml") + "\n\n")
	b.WriteString("  " + sMuted.Render(fmt.Sprintf(
		"Then %d steps converge this machine to it.", len(StepsFor(w.M)))) + "\n")
	b.WriteString("  " + sMuted.Render("Re-running later repairs whatever has drifted.") + "\n")

	hints := []Hint{
		{Key: "enter", Action: "write it and converge", Kind: HintPrimary},
		{Key: "esc", Action: "back", Kind: HintNormal},
		{Key: "q", Action: "quit", Kind: HintSystem},
	}
	return wrapFrame(m.width, m.height, header, b.String(), "", statusBar(hints))
}

// RunWizard puts the five questions on screen and returns the manifest the user
// built. confirmed is false when they quit, in which case nothing should be
// written.
func RunWizard(d Device) (Manifest, bool, error) {
	w := NewWizard(d)
	p := tea.NewProgram(wizardModel{w: w, width: 80, height: 24}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return Manifest{}, false, err
	}
	if w.Cancelled || !w.Done {
		return Manifest{}, false, nil
	}
	return w.M, true, nil
}

// wrapFrame is wrapScreen without the Model dependency, so screens that aren't
// part of the stage-picker can use the same layout.
func wrapFrame(width, height int, header, body, footer, bar string) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	out := lipgloss.JoinVertical(lipgloss.Left,
		sBright.Render(header), "", body, "", sDim.Render(footer),
	)
	padLines := height - (strings.Count(out, "\n") + 1) - (strings.Count(bar, "\n") + 1) - 1
	if padLines < 0 {
		padLines = 0
	}
	return lipgloss.NewStyle().Width(width).Render(out) + "\n" +
		strings.Repeat("\n", padLines) +
		lipgloss.NewStyle().Width(width).Render(bar)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
