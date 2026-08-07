package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuRow is one row in the Pick Menu.
type MenuRow struct {
	Kind   string // "stage" | "preset" | "action" | "sep"
	Stage  *Stage
	Preset *Preset
	ID     string // for actions: "review" | "quit"
	Label  string // for actions
}

func (m Model) menuRows() []MenuRow {
	rows := make([]MenuRow, 0, len(m.cat.Stages)+len(m.cat.Presets)+4)
	for _, st := range m.cat.Stages {
		rows = append(rows, MenuRow{Kind: "stage", Stage: st})
	}
	rows = append(rows, MenuRow{Kind: "sep"})
	for _, p := range m.cat.Presets {
		rows = append(rows, MenuRow{Kind: "preset", Preset: p})
	}
	rows = append(rows, MenuRow{Kind: "sep"})
	rows = append(rows, MenuRow{Kind: "action", ID: "review", Label: fmt.Sprintf("Review my picks (%d)", m.totalPicked())})
	rows = append(rows, MenuRow{Kind: "action", ID: "quit", Label: "Quit"})
	return rows
}

// focusableIndex returns the index of the nth focusable row (skipping separators).
func focusableMenuIndex(rows []MenuRow, cursor int) int {
	count := -1
	for i, r := range rows {
		if r.Kind == "sep" {
			continue
		}
		count++
		if count == cursor {
			return i
		}
	}
	return -1
}

func focusableMenuCount(rows []MenuRow) int {
	n := 0
	for _, r := range rows {
		if r.Kind != "sep" {
			n++
		}
	}
	return n
}

func (m Model) viewMenu() string {
	rows := m.menuRows()
	focusedIdx := focusableMenuIndex(rows, m.cursor)

	leftW, rightW := splitWidths(m.width)

	// HEADER
	header := sBright.Render("? Where to next?") + "  " +
		sDim.Render("(↑ ↓ move · enter open)")

	// LEFT: rows
	leftLines := make([]string, 0, len(rows))
	for i, r := range rows {
		focused := i == focusedIdx
		leftLines = append(leftLines, renderMenuRow(m, r, focused, leftW-2))
	}
	left := strings.Join(leftLines, "\n")

	// RIGHT: preview pane
	var right string
	if focusedIdx >= 0 {
		right = renderMenuPreview(m, rows[focusedIdx], rightW-2)
	}

	// Compose split pane.
	leftCol := lipgloss.NewStyle().Width(leftW).Render(left)
	rightCol := lipgloss.NewStyle().
		Width(rightW).
		BorderStyle(lipgloss.Border{Left: "┊"}).
		BorderForeground(colorDim).
		BorderLeft(true).
		PaddingLeft(1).
		Render(right)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)

	// FOOTER
	footer := sDim.Render(fmt.Sprintf("%d of %d commands selected total",
		m.totalPicked(), m.cat.TotalCommands()))

	hints := []Hint{
		{Key: "↑↓", Action: "move", Kind: HintNormal},
		{Key: "enter", Action: "open", Kind: HintPrimary},
		{Key: "/", Action: "search", Kind: HintNormal},
		{Key: "r", Action: "reset", Kind: HintSystem},
	}
	bar := statusBar(hints)

	return wrapScreen(m, header, body, footer, bar)
}

// renderMenuRow draws one menu row.
func renderMenuRow(m Model, r MenuRow, focused bool, w int) string {
	cursor := "  "
	if focused {
		cursor = sCursor.Render("❯ ")
	}
	switch r.Kind {
	case "sep":
		return sDim.Render(strings.Repeat("─", min(w, 6)))
	case "stage":
		st := r.Stage
		picked := m.stagePickedCount(st)
		total := len(st.Items)
		title := fmt.Sprintf("%s  %s", st.Num, st.Short)
		label := pad(title, 16)
		sigil := st.Sigil
		var sigilStyled string
		if focused || picked > 0 {
			sigilStyled = sAccent.Render(sigil)
		} else {
			sigilStyled = sDim.Render(sigil)
		}
		var labelStyled, countStyled string
		if focused {
			labelStyled = sBright.Render(label)
		} else {
			labelStyled = sText.Render(label)
		}
		countText := fmt.Sprintf("%d/%d", picked, total)
		if picked > 0 {
			countStyled = sAccent.Render(countText)
		} else {
			countStyled = sDim.Render(countText)
		}
		marker := "  "
		if m.installedStages[st.ID] {
			marker = sAccent.Render("⚡")
		} else if picked == total && total > 0 {
			marker = sAccent.Render(" ✓")
		}
		return cursor + sigilStyled + " " + labelStyled + " " + countStyled + " " + marker
	case "preset":
		p := r.Preset
		matches := len(p.CommandIDs)
		icon := "✦"
		var iconStyled, nameStyled string
		if focused {
			iconStyled = sAccent.Render(icon)
			nameStyled = sBright.Render(p.Name)
		} else {
			iconStyled = sDim.Render(icon)
			nameStyled = sText.Render(p.Name)
		}
		line := fmt.Sprintf("%s %s · %s", iconStyled, nameStyled, sMuted.Render(p.Tagline))
		return cursor + line + " " + sDim.Render(fmt.Sprintf("%d", matches))
	case "action":
		var label string
		if focused {
			label = sBright.Render(r.Label)
		} else {
			label = sText.Render(r.Label)
		}
		return cursor + label
	}
	return ""
}

// renderMenuPreview returns the right-pane content for a focused menu row.
func renderMenuPreview(m Model, r MenuRow, w int) string {
	switch r.Kind {
	case "stage":
		return previewStage(m, r.Stage, w)
	case "preset":
		return previewPreset(m, r.Preset, w)
	case "action":
		return previewAction(m, r.ID, w)
	}
	return ""
}

func previewStage(m Model, st *Stage, w int) string {
	lines := []string{
		sDim.Render("── stage ──"),
		sAccent.Render(st.Sigil) + " " + sBright.Render(fmt.Sprintf("%s %s", st.Num, st.Short)),
		sMuted.Render(st.Tagline),
		"",
	}
	if len(st.Groups) > 0 {
		lines = append(lines, sText.Render("includes"))
		for _, g := range st.Groups {
			cmds := m.cat.CommandsInGroup(st, g.ID)
			picked := 0
			for _, c := range cmds {
				if m.picked[c.ID] {
					picked++
				}
			}
			lines = append(lines, "  "+sText.Render(pad(g.Name, 22))+sDim.Render(fmt.Sprintf("%d/%d", picked, len(cmds))))
		}
	} else {
		lines = append(lines, sText.Render("commands"))
		shown := 0
		for _, c := range st.Items {
			if shown >= 8 {
				lines = append(lines, sDim.Render(fmt.Sprintf("  … +%d more", len(st.Items)-shown)))
				break
			}
			mark := sDim.Render("◯")
			if m.picked[c.ID] {
				mark = sAccent.Render("◉")
			}
			lines = append(lines, "  "+mark+" "+sText.Render(truncate(c.Title, w-6)))
			shown++
		}
	}
	picked := m.stagePickedCount(st)
	total := len(st.Items)
	mins := total * 2
	lines = append(lines, "",
		sDim.Render(fmt.Sprintf("%d/%d picked · ~%d min if all", picked, total, mins)))
	return strings.Join(lines, "\n")
}

func previewPreset(m Model, p *Preset, w int) string {
	lines := []string{
		sDim.Render("── preset ──"),
		sAccent.Render("✦ ") + sBright.Render(p.Name),
		sMuted.Render(p.Tagline),
		"",
		sText.Render(fmt.Sprintf("applies %d commands", len(p.CommandIDs))),
	}
	shown := 0
	for _, id := range p.CommandIDs {
		c := m.cat.CmdByID(id)
		if c == nil {
			continue
		}
		if shown >= 12 {
			lines = append(lines, sDim.Render(fmt.Sprintf("  … +%d more", len(p.CommandIDs)-shown)))
			break
		}
		lines = append(lines, "  "+sAccent.Render("+")+" "+sText.Render(truncate(c.Title, w-4)))
		shown++
	}
	lines = append(lines, "", sDim.Render("enter to apply · won't deselect anything"))
	return strings.Join(lines, "\n")
}

func previewAction(m Model, id string, w int) string {
	switch id {
	case "review":
		lines := []string{
			sDim.Render("── review ──"),
			sBright.Render("Review my picks"),
			sMuted.Render("Cross-stage summary before writing a script."),
			"",
		}
		if m.totalPicked() == 0 {
			lines = append(lines, sWarn.Render("nothing picked yet"))
		} else {
			lines = append(lines, sText.Render(fmt.Sprintf("%d commands queued · est. ~%d min", m.totalPicked(), m.estMinutes())))
		}
		return strings.Join(lines, "\n")
	case "quit":
		return strings.Join([]string{
			sDim.Render("── quit ──"),
			sBright.Render("Quit"),
			sMuted.Render("Exits the wizard. Picks are not saved."),
			"",
			sDim.Render(fmt.Sprintf("%d commands queued so far", m.totalPicked())),
		}, "\n")
	}
	_ = w
	return ""
}

func (m Model) keyMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.menuRows()
	maxF := focusableMenuCount(rows) - 1
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "down", "j":
		if m.cursor < maxF {
			m.cursor++
		}
		return m, nil
	case "/":
		m.priorView = PickMenu
		m.pickView = PickSearch
		m.searchQuery = ""
		m.searchResults = m.runSearch("", "")
		m.cursor = 0
		return m, nil
	case "enter":
		focusedIdx := focusableMenuIndex(rows, m.cursor)
		if focusedIdx < 0 {
			return m, nil
		}
		row := rows[focusedIdx]
		switch row.Kind {
		case "stage":
			m.menuCursor = m.cursor
			m.currentStageID = row.Stage.ID
			m.currentGroupID = ""
			m.pickView = PickStage
			m.cursor = 0
			return m, nil
		case "preset":
			for _, id := range row.Preset.CommandIDs {
				m.picked[id] = true
			}
			return m, nil
		case "action":
			switch row.ID {
			case "review":
				m.step = StepReview
				return m, nil
			case "quit":
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

// splitWidths returns left/right column widths for the split layout (~52/48).
func splitWidths(total int) (int, int) {
	if total <= 0 {
		total = 80
	}
	left := total * 52 / 100
	if left < 36 {
		left = 36
	}
	right := total - left
	if right < 30 {
		right = 30
		left = total - right
	}
	return left, right
}

// wrapScreen builds the standard 5-section layout used by most screens:
//
//	header
//	<blank>
//	body
//	<blank>
//	footer
//	status bar pinned at bottom.
func wrapScreen(m Model, header, body, footer, bar string) string {
	bodyW := m.width
	if bodyW <= 0 {
		bodyW = 80
	}
	out := lipgloss.JoinVertical(lipgloss.Left,
		sBright.Render(header),
		"",
		body,
		"",
		sDim.Render(footer),
	)
	innerH := m.height
	if innerH <= 0 {
		innerH = 24
	}
	bodyLines := strings.Count(out, "\n") + 1
	barLines := strings.Count(bar, "\n") + 1
	pad := innerH - bodyLines - barLines - 1
	if pad < 0 {
		pad = 0
	}
	return lipgloss.NewStyle().Width(bodyW).Render(out) + "\n" +
		strings.Repeat("\n", pad) +
		lipgloss.NewStyle().Width(bodyW).Render(bar)
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	// Naive truncation on rune count.
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
