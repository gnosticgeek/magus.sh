package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// runTickMsg advances the final-run install simulation.
type runTickMsg struct {
	runID int
}

// orderedPicks returns picked commands in stage → order sequence.
func (m Model) orderedPicks() []*Cmd {
	out := make([]*Cmd, 0, len(m.picked))
	for _, st := range m.cat.Stages {
		for _, c := range st.Items {
			if m.picked[c.ID] {
				out = append(out, c)
			}
		}
	}
	return out
}

func (m Model) viewRun() string {
	picks := m.orderedPicks()
	header := sBright.Render("script ready.")

	var body strings.Builder

	switch m.runPhase {
	case InstallPrompt:
		body.WriteString(sText.Render("run now?  ") +
			sAccent.Render("[y]") + sText.Render(" yes, install     ") +
			sAccent.Render("[n]") + sText.Render(" save & paste later"))
		body.WriteString("\n\n")
		body.WriteString(sMuted.Render("or later in Konsole:") + "\n")
		body.WriteString("  " + sDim.Render("$ ") + sText.Render("bash /tmp/magus.sh") + "\n\n")
		body.WriteString(sDim.Render("── preview ────────────────────────────────────") + "\n")
		body.WriteString(sDim.Render("#!/usr/bin/env bash") + "\n")
		body.WriteString(sDim.Render(fmt.Sprintf("# magus.sh — %d commands", len(picks))) + "\n\n")
		shown := 0
		for _, c := range picks {
			if shown >= 4 {
				body.WriteString(sDim.Render(fmt.Sprintf("… +%d more", len(picks)-shown)) + "\n")
				break
			}
			body.WriteString(sDim.Render("# "+c.Title) + "\n")
			if len(c.Run) > 0 {
				body.WriteString(sText.Render(truncate(firstLine(c.Run[0]), 60)) + "\n")
			}
			shown++
		}
	case InstallRunning, InstallFailed, InstallDone:
		total := len(picks)
		body.WriteString(sBright.Render(fmt.Sprintf("installing %d/%d  ", m.runIndex, total)) +
			progressBar(m.runIndex, total, 22) + "\n\n")
		from := len(m.runLog) - 3
		if from < 0 {
			from = 0
		}
		for _, e := range m.runLog[from:] {
			mark := sAccent.Render("✓ ")
			if e.Result == "skipped" {
				mark = sDim.Render("— ")
			}
			body.WriteString("  " + mark + sText.Render(e.Title) + "\n")
		}
		if m.runPhase == InstallRunning && m.runIndex < total {
			c := picks[m.runIndex]
			body.WriteString("  " + sCursor.Render("› ") + sBright.Render(c.Title) + sCursor.Render(" ▌") + "\n")
		}
		if m.runPhase == InstallFailed && m.runIndex < total {
			c := picks[m.runIndex]
			body.WriteString("  " + sWarn.Render("✗ ") + sBright.Render(c.Title) + "\n\n")
			body.WriteString("  " + sWarn.Render("error · simulated failure"))
			body.WriteString(" " + sDim.Render("(simulated)") + "\n")
		}
		if m.runPhase == InstallDone {
			done := 0
			skipped := 0
			for _, e := range m.runLog {
				switch e.Result {
				case "done":
					done++
				case "skipped":
					skipped++
				}
			}
			body.WriteString("\n" + sAccent.Render("✓ ") + sBright.Render("all done.") + "\n")
			body.WriteString(sText.Render(fmt.Sprintf("%d installed · %d skipped", done, skipped)) + "\n")
		}
	}

	var hints []Hint
	switch m.runPhase {
	case InstallPrompt:
		hints = []Hint{
			{Key: "y", Action: "run now", Kind: HintPrimary},
			{Key: "n", Action: "save & exit", Kind: HintNormal},
			{Key: "r", Action: "reset", Kind: HintSystem},
		}
	case InstallFailed:
		hints = []Hint{
			{Key: "y", Action: "retry", Kind: HintPrimary},
			{Key: "s", Action: "skip", Kind: HintNormal},
			{Key: "q", Action: "abort", Kind: HintSystem},
		}
	case InstallDone:
		hints = []Hint{
			{Key: "enter", Action: "finish", Kind: HintPrimary},
			{Key: "r", Action: "reset", Kind: HintSystem},
		}
	default:
		hints = []Hint{{Key: "wait", Action: "running…", Kind: HintNormal}}
	}

	return wrapScreen(m, header, body.String(), "", statusBar(hints))
}

func (m Model) keyRun(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.runPhase {
	case InstallPrompt:
		switch msg.String() {
		case "y", "Y":
			m.runPhase = InstallRunning
			m.runIndex = 0
			m.runLog = nil
			m.runRunID++
			id := m.runRunID
			return m, tea.Tick(installStepDelay, func(time.Time) tea.Msg { return runTickMsg{runID: id} })
		case "n", "N":
			return m, tea.Quit
		}
	case InstallFailed:
		switch msg.String() {
		case "y", "Y":
			m.runPhase = InstallRunning
			id := m.runRunID
			return m, tea.Tick(installStepDelay, func(time.Time) tea.Msg { return runTickMsg{runID: id} })
		case "s", "S":
			picks := m.orderedPicks()
			if m.runIndex < len(picks) {
				m.runLog = append(m.runLog, InstallEntry{Title: picks[m.runIndex].Title, Result: "skipped"})
				m.runIndex++
			}
			m.runPhase = InstallRunning
			id := m.runRunID
			return m, tea.Tick(installStepDelay, func(time.Time) tea.Msg { return runTickMsg{runID: id} })
		case "q", "Q":
			return m, tea.Quit
		}
	case InstallDone:
		if msg.Type == tea.KeyEnter {
			return m, tea.Quit
		}
	}
	return m, nil
}

// advanceRun is the run-mode counterpart to advanceInstall.
func (m Model) advanceRun(msg runTickMsg) (tea.Model, tea.Cmd) {
	if msg.runID != m.runRunID || m.runPhase != InstallRunning {
		return m, nil
	}
	picks := m.orderedPicks()
	// simulate failure halfway through if there's enough material to be interesting
	if m.runIndex == len(picks)/2 && len(picks) >= 6 && !hasFailed(m.runLog) {
		m.runPhase = InstallFailed
		return m, nil
	}
	c := picks[m.runIndex]
	m.runLog = append(m.runLog, InstallEntry{Title: c.Title, Result: "done"})
	m.runIndex++
	if m.runIndex >= len(picks) {
		m.runPhase = InstallDone
		return m, nil
	}
	id := m.runRunID
	return m, tea.Tick(installStepDelay, func(time.Time) tea.Msg { return runTickMsg{runID: id} })
}

func hasFailed(log []InstallEntry) bool {
	for _, e := range log {
		if e.Result == "failed" {
			return true
		}
	}
	return false
}
