package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

// MAGUS wordmark — block letters, kept tight so it fits inside 80 cols
// next to the info box on the right.
var magusWordmark = []string{
	"███╗   ███╗  █████╗   ██████╗  ██╗   ██╗ ███████╗",
	"████╗ ████║ ██╔══██╗ ██╔════╝  ██║   ██║ ██╔════╝",
	"██╔████╔██║ ███████║ ██║  ███╗ ██║   ██║ ███████╗",
	"██║╚██╔╝██║ ██╔══██║ ██║   ██║ ██║   ██║ ╚════██║",
	"██║ ╚═╝ ██║ ██║  ██║ ╚██████╔╝ ╚██████╔╝ ███████║",
	"╚═╝     ╚═╝ ╚═╝  ╚═╝  ╚═════╝   ╚═════╝  ╚══════╝",
}

func (m Model) viewSplash() string {
	// Info box on the right
	info := []string{
		sBright.Render("magus.sh v0.1"),
		sMuted.Render("spells     ") + sText.Render(fmt.Sprintf("%d", m.cat.TotalCommands())),
		sMuted.Render("stages     ") + sText.Render(fmt.Sprintf("%d", len(m.cat.Stages))),
		sMuted.Render("runtime    ") + sText.Render("~10 min"),
	}

	left := lipgloss.NewStyle().Foreground(colorAccent).Render(strings.Join(magusWordmark, "\n"))
	right := lipgloss.NewStyle().PaddingTop(1).PaddingLeft(2).Render(strings.Join(info, "\n"))

	header := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	tagline := sMuted.Italic(true).Render("your device, transmuted in ten minutes")

	// Stage legend pulls sigils from the catalogue so we never drift.
	legendParts := make([]string, 0, len(m.cat.Stages)*2)
	for i, st := range m.cat.Stages {
		if i > 0 {
			legendParts = append(legendParts, sDim.Render(" · "))
		}
		legendParts = append(legendParts,
			sAccent.Render(st.Sigil)+" "+sText.Render(strings.ToLower(st.Short)))
	}
	legend := sDim.Render("── ") + lipgloss.JoinHorizontal(lipgloss.Center, legendParts...) + sDim.Render(" ──")

	bullets := []string{
		sAccent.Render("✓") + sText.Render(" repeatable — safe to run again"),
		sAccent.Render("✓") + sText.Render(" no telemetry · one script · paste, pick, run"),
	}

	cta := sBright.Render("press ") + sAccent.Render("[enter]") + sBright.Render(" to begin the ceremony")

	body := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		tagline,
		"",
		legend,
		"",
		bullets[0],
		bullets[1],
		"",
		cta,
	)

	// Status bar
	hints := []Hint{
		{Key: "enter", Action: "begin", Kind: HintPrimary},
		{Key: "ctrl+c", Action: "quit", Kind: HintSystem},
	}
	bar := statusBar(hints)

	return frame(m, body, bar)
}

func (m Model) keySplash(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.step = StepPick
		m.pickView = PickMenu
		m.cursor = 0
		return m, nil
	}
	return m, nil
}

// frame wraps the body and status bar to fill the terminal — centres the body
// vertically when there's room, and pins the status bar to the bottom row.
func frame(m Model, body, bar string) string {
	innerW := m.width
	if innerW <= 0 {
		innerW = 80
	}
	innerH := m.height
	if innerH <= 0 {
		innerH = 24
	}

	bodyLines := strings.Count(body, "\n") + 1
	barLines := strings.Count(bar, "\n") + 1
	pad := innerH - bodyLines - barLines - 1
	if pad < 0 {
		pad = 0
	}

	return lipgloss.NewStyle().
		Width(innerW).
		Render(body) + "\n" +
		strings.Repeat("\n", pad) +
		lipgloss.NewStyle().Width(innerW).Render(bar)
}
