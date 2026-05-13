package main

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) viewReview() string {
	header := sBright.Render(fmt.Sprintf("%d commands ready · est. ~%d min", m.totalPicked(), m.estMinutes()))
	if m.totalPicked() == 0 {
		header = sWarn.Render("nothing picked.")
	}

	var body strings.Builder
	if m.totalPicked() == 0 {
		body.WriteString(sMuted.Render("press [←] to go back, or [r] to restart."))
	} else {
		// Group picks by stage, in stage order.
		for _, st := range m.cat.Stages {
			var lines []string
			for _, c := range st.Items {
				if m.picked[c.ID] {
					lines = append(lines, c.Title)
				}
			}
			if len(lines) == 0 {
				continue
			}
			sort.Strings(lines) // stable, deterministic for review
			body.WriteString("  " + sBright.Render(fmt.Sprintf("%s %s", st.Num, st.Short)) + "\n")
			for _, t := range lines {
				body.WriteString("     " + sAccent.Render("✓ ") + sText.Render(t) + "\n")
			}
			body.WriteByte('\n')
		}
	}

	hints := []Hint{
		{Key: "y", Action: "write to disk", Kind: HintPrimary},
		{Key: "e", Action: "edit picks", Kind: HintNormal},
		{Key: "q", Action: "quit", Kind: HintSystem},
	}
	footer := sDim.Render("―")
	return wrapScreen(m, header, body.String(), footer, statusBar(hints))
}

func (m Model) keyReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.totalPicked() == 0 {
			return m, nil
		}
		m.step = StepWrite
		m.writeStep = 0
		m.writeBytes = scriptKB(m.totalPicked())
		return m, tea.Cmd(writeStartCmd())
	case "e", "E":
		m.step = StepPick
		m.pickView = PickMenu
		m.cursor = m.menuCursor
		return m, nil
	case "q", "Q":
		return m, tea.Quit
	case "left":
		m.step = StepPick
		m.pickView = PickMenu
		return m, nil
	}
	return m, nil
}
