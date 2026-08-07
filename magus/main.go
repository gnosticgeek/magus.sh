package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Screen is the top-level state of the TUI wizard. It is distinct from Step,
// which is a unit of convergence in the reconciler.
type Screen int

const (
	StepSplash Screen = iota
	StepPick
	StepReview
	StepWrite
	StepRun
)

// PickView is the active sub-screen inside the Pick step.
type PickView int

const (
	PickMenu PickView = iota
	PickStage
	PickSearch
	PickInstalling
)

// InstallPhase is shared by stage-install and final-run flows.
type InstallPhase int

const (
	InstallPrompt InstallPhase = iota
	InstallRunning
	InstallFailed
	InstallDone
)

// InstallEntry records the outcome for one command during install.
type InstallEntry struct {
	Title  string
	Result string // "done" | "skipped" | "failed"
}

// Model is the application state. The screens are pure functions of this.
type Model struct {
	cat *Catalogue

	width, height int

	step Screen

	// Pick state
	pickView        PickView
	cursor          int
	menuCursor      int // remembered so back-from-stage restores the cursor
	picked          map[string]bool
	currentStageID  string
	currentGroupID  string
	priorView       PickView // PickMenu or PickStage (for search return)
	searchQuery     string
	searchResults   []searchHit
	installedStages map[string]bool

	// Install state (re-used for the final Run)
	installPhase InstallPhase
	installIndex int
	installLog   []InstallEntry
	installPicks []*Cmd
	installRunID int // bumped to cancel stale async ticks

	// Write animation
	writeStep  int
	writeBytes float64

	// Run state
	runPhase InstallPhase
	runIndex int
	runLog   []InstallEntry
	runRunID int

	// Reset bookkeeping
	startedAt time.Time
}

func newModel(cat *Catalogue) Model {
	return Model{
		cat:             cat,
		step:            StepSplash,
		pickView:        PickMenu,
		picked:          make(map[string]bool),
		installedStages: make(map[string]bool),
		width:           80,
		height:          24,
		startedAt:       time.Now(),
	}
}

// reset returns the model to its splash state, clearing picks.
func (m Model) reset() Model {
	n := newModel(m.cat)
	n.width = m.width
	n.height = m.height
	return n
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		// Globals: r/R = reset, ctrl+c = hard quit.
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "r", "R":
			// Only at top-level views — inside install/run we treat r as system reset too.
			return m.reset(), nil
		}
		return m.handleKey(msg)
	case installTickMsg:
		return m.advanceInstall(msg)
	case writeTickMsg:
		return m.advanceWrite(msg)
	case runTickMsg:
		return m.advanceRun(msg)
	}
	return m, nil
}

func (m Model) View() string {
	switch m.step {
	case StepSplash:
		return m.viewSplash()
	case StepPick:
		switch m.pickView {
		case PickMenu:
			return m.viewMenu()
		case PickStage:
			return m.viewStage()
		case PickSearch:
			return m.viewSearch()
		case PickInstalling:
			return m.viewInstall()
		}
	case StepReview:
		return m.viewReview()
	case StepWrite:
		return m.viewWrite()
	case StepRun:
		return m.viewRun()
	}
	return "unknown state"
}

// handleKey dispatches based on step + view.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.step {
	case StepSplash:
		return m.keySplash(msg)
	case StepPick:
		switch m.pickView {
		case PickMenu:
			return m.keyMenu(msg)
		case PickStage:
			return m.keyStage(msg)
		case PickSearch:
			return m.keySearch(msg)
		case PickInstalling:
			return m.keyInstall(msg)
		}
	case StepReview:
		return m.keyReview(msg)
	case StepWrite:
		return m.keyWrite(msg)
	case StepRun:
		return m.keyRun(msg)
	}
	return m, nil
}

// pickedCount returns the number of picked items in a given scope.
// stage == nil → total picks across all stages.
// groupID == ""  → all items in stage.
func (m Model) pickedCount(stage *Stage, groupID string) int {
	if stage == nil {
		return len(m.picked)
	}
	items := m.cat.CommandsInGroup(stage, groupID)
	n := 0
	for _, c := range items {
		if m.picked[c.ID] {
			n++
		}
	}
	return n
}

// stagePickedCount returns picks in a stage.
func (m Model) stagePickedCount(st *Stage) int { return m.pickedCount(st, "") }

// totalPicked returns picks across all stages.
func (m Model) totalPicked() int { return len(m.picked) }

// estMinutes returns the rough install-time estimate from picked.length.
func (m Model) estMinutes() int {
	n := len(m.picked)
	if n == 0 {
		return 0
	}
	v := int(float64(n) * 1.4)
	if v < 2 {
		v = 2
	}
	return v
}

// togglePick flips a single command's selection.
func (m *Model) togglePick(id string) {
	if m.picked[id] {
		delete(m.picked, id)
	} else {
		m.picked[id] = true
	}
}

func main() {
	// With a verb, magus is a reconciler driven from the command line; with no
	// arguments it is the TUI. Both paths end up converging the same manifest —
	// the TUI is a manifest builder, not a second implementation (§5).
	if len(os.Args) > 1 {
		os.Exit(runCLI(os.Args[1:]))
	}

	cat, err := loadCatalogue()
	if err != nil {
		fmt.Fprintf(os.Stderr, "magus: %v\n", err)
		os.Exit(1)
	}
	p := tea.NewProgram(newModel(cat), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "magus: %v\n", err)
		os.Exit(1)
	}
}
