package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StageRow is one row in either the group list or command list.
type StageRow struct {
	Kind      string // "item" | "group" | "select-all" | "install" | "back" | "back-group" | "sep"
	Cmd       *Cmd
	Group     *Group
	Label     string
	Disabled  bool
}

// stageRows builds the row set for the current view (group list or command list).
func (m Model) stageRows() []StageRow {
	st := m.cat.StageByID(m.currentStageID)
	if st == nil {
		return nil
	}
	// Group list: when stage has groups AND we're at stage root.
	if len(st.Groups) > 0 && m.currentGroupID == "" {
		rows := make([]StageRow, 0, len(st.Groups)+4)
		for i := range st.Groups {
			g := &st.Groups[i]
			rows = append(rows, StageRow{Kind: "group", Group: g})
		}
		rows = append(rows, StageRow{Kind: "sep"})
		picked := m.stagePickedCount(st)
		if picked == 0 {
			rows = append(rows, StageRow{Kind: "install", Disabled: true, Label: "⚡ pick at least one to install"})
		} else {
			rows = append(rows, StageRow{Kind: "install", Label: fmt.Sprintf("⚡ Install %d commands", picked)})
		}
		rows = append(rows, StageRow{Kind: "back", Label: "← back to stages"})
		return rows
	}

	// Command list — either inside a group, or stage has no groups.
	items := m.cat.CommandsInGroup(st, m.currentGroupID)
	rows := make([]StageRow, 0, len(items)+5)
	for _, c := range items {
		rows = append(rows, StageRow{Kind: "item", Cmd: c})
	}
	rows = append(rows, StageRow{Kind: "sep"})
	rows = append(rows, StageRow{Kind: "select-all", Label: "select all in this stage"})

	if m.currentGroupID != "" {
		rows = append(rows, StageRow{Kind: "back-group", Label: "← back to " + st.Short})
	} else {
		picked := m.stagePickedCount(st)
		if picked == 0 {
			rows = append(rows, StageRow{Kind: "install", Disabled: true, Label: "⚡ pick at least one to install"})
		} else {
			rows = append(rows, StageRow{Kind: "install", Label: fmt.Sprintf("⚡ Install %d commands", picked)})
		}
		rows = append(rows, StageRow{Kind: "back", Label: "← back to stages"})
	}
	return rows
}

func focusableStageIndex(rows []StageRow, cursor int) int {
	count := -1
	for i, r := range rows {
		if r.Kind == "sep" || r.Disabled {
			continue
		}
		count++
		if count == cursor {
			return i
		}
	}
	return -1
}

func focusableStageCount(rows []StageRow) int {
	n := 0
	for _, r := range rows {
		if r.Kind == "sep" || r.Disabled {
			continue
		}
		n++
	}
	return n
}

func (m Model) viewStage() string {
	st := m.cat.StageByID(m.currentStageID)
	rows := m.stageRows()
	focusedIdx := focusableStageIndex(rows, m.cursor)

	leftW, rightW := splitWidths(m.width)

	// HEADER — includes group breadcrumb + tab strip when inside a group.
	var header string
	headerLine := fmt.Sprintf("? %s · %s", st.Num, st.Short)
	if m.currentGroupID != "" {
		g := m.cat.GroupByID(st, m.currentGroupID)
		headerLine += " " + sDim.Render("›") + " " + sBright.Render(g.Name)
	}
	headerHint := sDim.Render("(space toggle · enter open)")
	if len(st.Groups) > 0 && m.currentGroupID == "" {
		headerHint = sDim.Render("(↑ ↓ move · enter open)")
	}
	header = sBright.Render(headerLine) + "  " + headerHint
	if m.currentGroupID != "" {
		strip := renderGroupTabStrip(st, m.currentGroupID)
		header += "\n" + strip
	}

	// LEFT: rows
	leftLines := make([]string, 0, len(rows))
	for i, r := range rows {
		focused := i == focusedIdx
		leftLines = append(leftLines, renderStageRow(m, r, focused, leftW-2))
	}
	left := strings.Join(leftLines, "\n")

	// RIGHT: preview
	var right string
	if focusedIdx >= 0 {
		right = renderStagePreview(m, st, rows[focusedIdx], rightW-2)
	}

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
	scope := "this stage"
	scopeItems := m.cat.CommandsInGroup(st, m.currentGroupID)
	picked := 0
	for _, c := range scopeItems {
		if m.picked[c.ID] {
			picked++
		}
	}
	footer := sDim.Render(fmt.Sprintf("%d of %d selected in %s", picked, len(scopeItems), scope))

	// HINTS — adapt to view.
	var hints []Hint
	if len(st.Groups) > 0 && m.currentGroupID == "" {
		hints = []Hint{
			{Key: "↑↓", Action: "move", Kind: HintNormal},
			{Key: "enter", Action: "open", Kind: HintPrimary},
			{Key: "esc", Action: "back", Kind: HintNormal},
			{Key: "/", Action: "search", Kind: HintNormal},
			{Key: "r", Action: "reset", Kind: HintSystem},
		}
	} else {
		hints = []Hint{
			{Key: "space", Action: "toggle", Kind: HintPrimary},
			{Key: "↑↓", Action: "move", Kind: HintNormal},
			{Key: "esc", Action: "back", Kind: HintNormal},
			{Key: "/", Action: "search", Kind: HintNormal},
			{Key: "r", Action: "reset", Kind: HintSystem},
		}
	}

	return wrapScreen(m, header, body, footer, statusBar(hints))
}

// renderGroupTabStrip lays out group tabs inline with the active one bracketed.
func renderGroupTabStrip(st *Stage, activeID string) string {
	parts := make([]string, 0, len(st.Groups)*2)
	for i, g := range st.Groups {
		if i > 0 {
			parts = append(parts, sDim.Render("  "))
		}
		if g.ID == activeID {
			parts = append(parts, sAccent.Render("["+g.Name+"]"))
		} else {
			parts = append(parts, sMuted.Render(g.Name))
		}
	}
	return strings.Join(parts, "")
}

func renderStageRow(m Model, r StageRow, focused bool, w int) string {
	cursor := "  "
	if focused {
		cursor = sCursor.Render("❯ ")
	}
	switch r.Kind {
	case "sep":
		return sDim.Render(strings.Repeat("─", min(w, 6)))
	case "item":
		c := r.Cmd
		var box string
		if m.picked[c.ID] {
			box = sCheck.Render("◉")
		} else {
			box = sMuted.Render("◯")
		}
		var title string
		switch {
		case focused:
			title = sBright.Render(c.Title)
		case m.picked[c.ID]:
			title = sText.Render(c.Title)
		default:
			title = sMuted.Render(c.Title)
		}
		return cursor + box + " " + title
	case "group":
		g := r.Group
		st := m.cat.StageByID(m.currentStageID)
		cmds := m.cat.CommandsInGroup(st, g.ID)
		picked := 0
		for _, c := range cmds {
			if m.picked[c.ID] {
				picked++
			}
		}
		label := pad(g.Name, 20)
		var nameStyled, countStyled string
		if focused {
			nameStyled = sBright.Render(label)
		} else {
			nameStyled = sText.Render(label)
		}
		count := fmt.Sprintf("%d/%d", picked, len(cmds))
		if picked > 0 {
			countStyled = sAccent.Render(count)
		} else {
			countStyled = sDim.Render(count)
		}
		return cursor + nameStyled + " " + countStyled + " " + sDim.Render("›")
	case "select-all":
		var label string
		if focused {
			label = sBright.Render(r.Label)
		} else {
			label = sText.Render(r.Label)
		}
		return cursor + label
	case "install":
		var label string
		if r.Disabled {
			label = sDim.Render(r.Label)
			return "  " + label // non-focusable, no cursor
		}
		if focused {
			label = sAccent.Render(r.Label)
		} else {
			label = sText.Render(r.Label)
		}
		return cursor + label
	case "back", "back-group":
		var label string
		if focused {
			label = sBright.Render(r.Label)
		} else {
			label = sMuted.Render(r.Label)
		}
		return cursor + label
	}
	return ""
}

func renderStagePreview(m Model, st *Stage, r StageRow, w int) string {
	switch r.Kind {
	case "item":
		return previewItem(m, st, r.Cmd, w)
	case "group":
		return previewGroup(m, st, r.Group, w)
	case "install":
		return previewInstall(m, st, r.Disabled, w)
	case "select-all":
		return previewSelectAll(m, st, w)
	case "back":
		return strings.Join([]string{
			sDim.Render("── back ──"),
			sBright.Render(r.Label),
			sMuted.Render("Picks carry across — your selections are safe."),
		}, "\n")
	case "back-group":
		return strings.Join([]string{
			sDim.Render("── back ──"),
			sBright.Render(r.Label),
			sMuted.Render("Return to the group list. Your picks stay put."),
		}, "\n")
	}
	return ""
}

func previewItem(m Model, st *Stage, c *Cmd, w int) string {
	mark := sMuted.Render("◯")
	if m.picked[c.ID] {
		mark = sCheck.Render("◉")
	}
	header := sDim.Render(fmt.Sprintf("── %s ──", strings.ToLower(st.Short)))
	lines := []string{
		header,
		mark + " " + sBright.Render(c.Title),
		sMuted.Render(wrap(c.Summary, w)),
		"",
		sText.Render("will run"),
	}
	for i, run := range c.Run {
		if i >= 3 {
			lines = append(lines, sDim.Render(fmt.Sprintf("  … +%d more", len(c.Run)-3)))
			break
		}
		first := firstLine(run)
		lines = append(lines, "  "+sDim.Render("$ ")+sText.Render(truncate(first, w-4)))
	}
	if c.Upstream != nil && c.Upstream.Name != "" {
		lines = append(lines, "", sText.Render("source"), "  "+sMuted.Render(truncate(c.Upstream.Name, w-2)))
		if c.Upstream.URL != "" {
			lines = append(lines, "  "+sDim.Render(truncate(c.Upstream.URL, w-2)))
		}
	}
	lines = append(lines, "", sDim.Render("space toggles · enter also toggles"))
	return strings.Join(lines, "\n")
}

func previewGroup(m Model, st *Stage, g *Group, w int) string {
	cmds := m.cat.CommandsInGroup(st, g.ID)
	picked := 0
	for _, c := range cmds {
		if m.picked[c.ID] {
			picked++
		}
	}
	lines := []string{
		sDim.Render("── group ──"),
		sBright.Render(g.Name),
		sDim.Render(fmt.Sprintf("%d commands · %d picked", len(cmds), picked)),
		"",
	}
	shown := 0
	for _, c := range cmds {
		if shown >= 10 {
			lines = append(lines, sDim.Render(fmt.Sprintf("  … +%d more", len(cmds)-shown)))
			break
		}
		mark := sMuted.Render("◯")
		if m.picked[c.ID] {
			mark = sCheck.Render("◉")
		}
		lines = append(lines, "  "+mark+" "+sText.Render(truncate(c.Title, w-4)))
		shown++
	}
	return strings.Join(lines, "\n")
}

func previewInstall(m Model, st *Stage, disabled bool, w int) string {
	if disabled || m.stagePickedCount(st) == 0 {
		return strings.Join([]string{
			sDim.Render("── install ──"),
			sWarn.Render("nothing picked yet"),
			sMuted.Render("space toggles a command…"),
		}, "\n")
	}
	picked := m.stagePickedCount(st)
	mins := picked * 1
	if mins < 2 {
		mins = 2
	}
	lines := []string{
		sDim.Render("── install ──"),
		sAccent.Render(fmt.Sprintf("⚡ %d commands ready", picked)),
		sDim.Render(fmt.Sprintf("est. ~%d min · %s stage only", mins, strings.ToLower(st.Short))),
		"",
	}
	shown := 0
	for _, c := range st.Items {
		if !m.picked[c.ID] {
			continue
		}
		if shown >= 8 {
			lines = append(lines, sDim.Render(fmt.Sprintf("  … +%d more", picked-shown)))
			break
		}
		lines = append(lines, "  "+sAccent.Render("✓")+" "+sText.Render(truncate(c.Title, w-4)))
		shown++
	}
	return strings.Join(lines, "\n")
}

func previewSelectAll(m Model, st *Stage, w int) string {
	items := m.cat.CommandsInGroup(st, m.currentGroupID)
	allPicked := len(items) > 0
	for _, c := range items {
		if !m.picked[c.ID] {
			allPicked = false
			break
		}
	}
	label := "Select all in this stage"
	if allPicked {
		label = "Deselect all in this stage"
	}
	scope := st.Short
	if m.currentGroupID != "" {
		g := m.cat.GroupByID(st, m.currentGroupID)
		scope = g.Name
	}
	_ = w
	return strings.Join([]string{
		sDim.Render("── shortcut ──"),
		sBright.Render(label),
		sMuted.Render(fmt.Sprintf("toggles every command in %s (%d total)", scope, len(items))),
	}, "\n")
}

func (m Model) keyStage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	st := m.cat.StageByID(m.currentStageID)
	rows := m.stageRows()
	maxF := focusableStageCount(rows) - 1
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
	case "esc":
		if m.currentGroupID != "" {
			m.currentGroupID = ""
			m.cursor = 0
		} else {
			m.pickView = PickMenu
			m.cursor = m.menuCursor
		}
		return m, nil
	case "/":
		m.priorView = PickStage
		m.pickView = PickSearch
		m.searchQuery = ""
		m.searchResults = m.runSearch("", m.currentStageID)
		m.cursor = 0
		return m, nil
	case " ":
		idx := focusableStageIndex(rows, m.cursor)
		if idx < 0 {
			return m, nil
		}
		row := rows[idx]
		if row.Kind == "item" {
			m.togglePick(row.Cmd.ID)
		} else if row.Kind == "select-all" {
			m.toggleSelectAll(st)
		}
		return m, nil
	case "enter":
		idx := focusableStageIndex(rows, m.cursor)
		if idx < 0 {
			return m, nil
		}
		row := rows[idx]
		switch row.Kind {
		case "item":
			m.togglePick(row.Cmd.ID)
		case "group":
			m.currentGroupID = row.Group.ID
			m.cursor = 0
		case "select-all":
			m.toggleSelectAll(st)
		case "install":
			if row.Disabled {
				return m, nil
			}
			return m.startStageInstall()
		case "back-group":
			m.currentGroupID = ""
			m.cursor = 0
		case "back":
			m.pickView = PickMenu
			m.cursor = m.menuCursor
		}
	}
	return m, nil
}

func (m *Model) toggleSelectAll(st *Stage) {
	items := m.cat.CommandsInGroup(st, m.currentGroupID)
	allPicked := len(items) > 0
	for _, c := range items {
		if !m.picked[c.ID] {
			allPicked = false
			break
		}
	}
	if allPicked {
		for _, c := range items {
			delete(m.picked, c.ID)
		}
	} else {
		for _, c := range items {
			m.picked[c.ID] = true
		}
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func wrap(s string, w int) string {
	if w <= 0 || lipgloss.Width(s) <= w {
		return s
	}
	// Simple word-wrap at word boundaries.
	words := strings.Fields(s)
	var b strings.Builder
	line := 0
	for i, word := range words {
		wl := lipgloss.Width(word)
		if line == 0 {
			b.WriteString(word)
			line = wl
		} else if line+1+wl > w {
			b.WriteByte('\n')
			b.WriteString(word)
			line = wl
		} else {
			b.WriteByte(' ')
			b.WriteString(word)
			line += 1 + wl
		}
		_ = i
	}
	return b.String()
}
