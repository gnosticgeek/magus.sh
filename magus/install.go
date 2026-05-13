package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// installTickMsg fires once per simulated install step (~800ms apart).
type installTickMsg struct {
	runID int
}

const installStepDelay = 800 * time.Millisecond

// startStageInstall transitions to the installing view and queues the first tick.
func (m Model) startStageInstall() (tea.Model, tea.Cmd) {
	st := m.cat.StageByID(m.currentStageID)
	picks := make([]*Cmd, 0, m.stagePickedCount(st))
	for _, c := range st.Items {
		if m.picked[c.ID] {
			picks = append(picks, c)
		}
	}
	if len(picks) == 0 {
		return m, nil
	}
	m.pickView = PickInstalling
	m.installPhase = InstallRunning
	m.installIndex = 0
	m.installLog = nil
	m.installPicks = picks
	m.installRunID++
	id := m.installRunID
	return m, tea.Tick(installStepDelay, func(time.Time) tea.Msg { return installTickMsg{runID: id} })
}

// advanceInstall handles each tick of the simulated install.
func (m Model) advanceInstall(msg installTickMsg) (tea.Model, tea.Cmd) {
	if msg.runID != m.installRunID || m.installPhase != InstallRunning {
		return m, nil // stale
	}
	// Simulated failure at index 2 when there are at least 4 commands.
	if m.installIndex == 2 && len(m.installPicks) >= 4 {
		m.installPhase = InstallFailed
		return m, nil
	}
	c := m.installPicks[m.installIndex]
	m.installLog = append(m.installLog, InstallEntry{Title: c.Title, Result: "done"})
	m.installIndex++
	if m.installIndex >= len(m.installPicks) {
		m.installPhase = InstallDone
		st := m.cat.StageByID(m.currentStageID)
		if st != nil {
			m.installedStages[st.ID] = true
		}
		return m, nil
	}
	id := m.installRunID
	return m, tea.Tick(installStepDelay, func(time.Time) tea.Msg { return installTickMsg{runID: id} })
}

func (m Model) keyInstall(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.installPhase {
	case InstallFailed:
		switch msg.String() {
		case "y", "Y":
			// retry — same index, clear simulated failure
			if m.installIndex == 2 {
				m.installPicks = append(m.installPicks[:2], m.installPicks[3:]...)
				m.installPicks = append([]*Cmd{}, m.installPicks...)
			}
			m.installPhase = InstallRunning
			id := m.installRunID
			return m, tea.Tick(installStepDelay, func(time.Time) tea.Msg { return installTickMsg{runID: id} })
		case "s", "S":
			c := m.installPicks[m.installIndex]
			m.installLog = append(m.installLog, InstallEntry{Title: c.Title, Result: "skipped"})
			m.installIndex++
			m.installPhase = InstallRunning
			if m.installIndex >= len(m.installPicks) {
				m.installPhase = InstallDone
				st := m.cat.StageByID(m.currentStageID)
				if st != nil {
					m.installedStages[st.ID] = true
				}
				return m, nil
			}
			id := m.installRunID
			return m, tea.Tick(installStepDelay, func(time.Time) tea.Msg { return installTickMsg{runID: id} })
		case "q", "Q":
			// abort — back to stage view, no install marker.
			m.pickView = PickStage
			m.installRunID++ // cancel any pending ticks
			return m, nil
		}
	case InstallDone:
		switch msg.Type {
		case tea.KeyEscape:
			m.pickView = PickMenu
			m.cursor = m.menuCursor
			return m, nil
		}
	}
	return m, nil
}

func (m Model) viewInstall() string {
	st := m.cat.StageByID(m.currentStageID)
	total := len(m.installPicks)
	header := sBright.Render(fmt.Sprintf("installing %d/%d  ", m.installIndex, total)) +
		progressBar(m.installIndex, total, 22)

	var body strings.Builder
	body.WriteByte('\n')

	switch m.installPhase {
	case InstallRunning, InstallFailed:
		// completed entries (tail)
		from := len(m.installLog) - 3
		if from < 0 {
			from = 0
		}
		for _, e := range m.installLog[from:] {
			mark := sAccent.Render("✓ ")
			if e.Result == "skipped" {
				mark = sDim.Render("— ")
			}
			body.WriteString("  " + mark + sText.Render(e.Title) + "\n")
		}
		if m.installPhase == InstallRunning && m.installIndex < total {
			c := m.installPicks[m.installIndex]
			body.WriteString("  " + sCursor.Render("› ") + sBright.Render(c.Title) + sCursor.Render(" ▌") + "\n")
			remaining := total - m.installIndex - 1
			if remaining > 0 {
				body.WriteString(sDim.Render(fmt.Sprintf("    %d more queued", remaining)) + "\n")
			}
		} else if m.installPhase == InstallFailed && m.installIndex < total {
			c := m.installPicks[m.installIndex]
			body.WriteString("  " + sWarn.Render("✗ ") + sBright.Render(c.Title) + "\n")
			body.WriteByte('\n')
			body.WriteString("  " + sWarn.Render("error · connection timed out — could not reach download server"))
			body.WriteString(" " + sDim.Render("(simulated)") + "\n")
			body.WriteString("  " + sDim.Render("exit code 1") + "\n")
		}
	case InstallDone:
		done := 0
		skipped := 0
		for _, e := range m.installLog {
			switch e.Result {
			case "done":
				done++
			case "skipped":
				skipped++
			}
		}
		body.WriteString(sBright.Render(fmt.Sprintf("%s done ", st.Short)) + sAccent.Render("✓") + "\n\n")
		body.WriteString(sText.Render(fmt.Sprintf("  %d installed   %d skipped", done, skipped)) + "\n\n")
		for _, e := range m.installLog {
			switch e.Result {
			case "done":
				body.WriteString("  " + sAccent.Render("✓ ") + sText.Render(e.Title) + "\n")
			case "skipped":
				body.WriteString("  " + sDim.Render("— ") + sMuted.Render(e.Title) + "\n")
			case "failed":
				body.WriteString("  " + sWarn.Render("✗ ") + sMuted.Render(e.Title) + "\n")
			}
		}
	}

	var hints []Hint
	switch m.installPhase {
	case InstallFailed:
		hints = []Hint{
			{Key: "y", Action: "retry", Kind: HintPrimary},
			{Key: "s", Action: "skip", Kind: HintNormal},
			{Key: "q", Action: "abort", Kind: HintSystem},
		}
	case InstallDone:
		hints = []Hint{
			{Key: "esc", Action: "continue", Kind: HintPrimary},
			{Key: "r", Action: "reset", Kind: HintSystem},
		}
	default:
		hints = []Hint{
			{Key: "wait", Action: "installing…", Kind: HintNormal},
		}
	}

	footer := ""
	return wrapScreen(m, header, body.String(), footer, statusBar(hints))
}

// progressBar renders 22-char filled/unfilled bar with percent.
func progressBar(done, total, width int) string {
	if total == 0 {
		return ""
	}
	filled := done * width / total
	if filled > width {
		filled = width
	}
	bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
	pct := done * 100 / total
	return sAccent.Render(bar) + sDim.Render(fmt.Sprintf("  %d%%", pct))
}
