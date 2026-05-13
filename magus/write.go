package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// writeTickMsg advances the script-write animation through its phases.
type writeTickMsg struct {
	step int
}

// scriptKB returns the rough KB estimate for a picked-count.
func scriptKB(n int) float64 {
	v := float64(n) * 0.18
	if v < 0.4 {
		v = 0.4
	}
	return v
}

func writeStartCmd() tea.Cmd {
	return tea.Tick(480*time.Millisecond, func(time.Time) tea.Msg { return writeTickMsg{step: 1} })
}

// advanceWrite steps through the four-stage write animation.
func (m Model) advanceWrite(msg writeTickMsg) (tea.Model, tea.Cmd) {
	if m.step != StepWrite {
		return m, nil
	}
	m.writeStep = msg.step
	switch msg.step {
	case 1:
		return m, tea.Tick(480*time.Millisecond, func(time.Time) tea.Msg { return writeTickMsg{step: 2} })
	case 2:
		return m, tea.Tick(380*time.Millisecond, func(time.Time) tea.Msg { return writeTickMsg{step: 3} })
	case 3:
		return m, tea.Tick(220*time.Millisecond, func(time.Time) tea.Msg { return writeTickMsg{step: 4} })
	case 4:
		// Done
		return m, nil
	}
	return m, nil
}

func (m Model) viewWrite() string {
	header := sBright.Render("writing magus.sh")
	var lines []string
	bytesLabel := fmt.Sprintf("%.1f KB", m.writeBytes)
	steps := []struct {
		text string
		done bool
	}{
		{"→ rendering magus.sh ................ done (" + bytesLabel + ")", m.writeStep >= 1},
		{"→ writing /tmp/magus.sh ............. done", m.writeStep >= 2},
		{"→ chmod +x .......................... done", m.writeStep >= 3},
	}
	for _, s := range steps {
		if s.done {
			lines = append(lines, sText.Render(s.text))
		} else {
			lines = append(lines, sDim.Render("→ "+strings.Repeat(".", 36)))
		}
	}

	if m.writeStep >= 4 {
		lines = append(lines, "",
			sBright.Render("saved to: ")+sAccent.Render("/tmp/magus.sh"),
			"",
			sDim.Render(fmt.Sprintf("%d commands · %s · idempotent", m.totalPicked(), bytesLabel)),
			"",
			sBright.Render("› press [enter] to continue")+sCursor.Render(" ▌"),
		)
	}

	hints := []Hint{}
	if m.writeStep >= 4 {
		hints = []Hint{
			{Key: "enter", Action: "continue", Kind: HintPrimary},
			{Key: "r", Action: "reset", Kind: HintSystem},
		}
	} else {
		hints = []Hint{{Key: "wait", Action: "writing…", Kind: HintNormal}}
	}
	return wrapScreen(m, header, strings.Join(lines, "\n"), "", statusBar(hints))
}

func (m Model) keyWrite(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.writeStep < 4 {
		return m, nil
	}
	if msg.Type == tea.KeyEnter {
		m.step = StepRun
		m.runPhase = InstallPrompt
		m.runLog = nil
		m.runIndex = 0
		return m, nil
	}
	return m, nil
}
