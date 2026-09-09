package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Keep readable columns at 80 columns; wider terminals also get a preview.
func (m *macModel) browserLayout() (columns, width int, split bool) {
	w := max(16, m.width-6)
	grid := m.screen == "browse" && (m.searching || m.category == "apps" || m.category == "tools" || m.category == "fonts")
	if grid && m.width >= 80 {
		width = w
		split = m.width >= 120
		if split {
			width = w - max(32, w/3) - 3
		}
		return max(2, width/34), width, split
	}
	if m.width >= 80 {
		return 1, (w - 3) / 2, true
	}
	return 1, w, false
}

func (m *macModel) gridView(rows []macRow, width, height, columns int) string {
	pageSize := max(1, height) * columns
	page := m.cursor / pageSize
	start := page * pageSize
	cellWidth := (width - (columns-1)*2) / columns
	browser := m.browserList(rows, cellWidth, height)
	var lines []string
	for row := 0; row < height; row++ {
		cells := []string{}
		for col := 0; col < columns; col++ {
			index := start + row*columns + col
			var cell strings.Builder
			if index < len(rows) {
				(macRowDelegate{m}).Render(&cell, browser, index, rows[index])
			}
			cells = append(cells, lipgloss.NewStyle().Width(cellWidth).Render(cell.String()))
		}
		lines = append(lines, strings.Join(cells, "  "))
	}
	return strings.Join(lines, "\n") + "\n" + sMuted.Render(fmt.Sprintf("%d results · page %d/%d · ←↑↓→ move", len(rows), page+1, max(1, (len(rows)+pageSize-1)/pageSize)))
}

func (m *macModel) moveGrid(key string, columns int) {
	count := len(m.rows())
	pageSize := max(1, m.browserHeight) * columns
	switch key {
	case "up":
		m.cursor = max(0, m.cursor-columns)
	case "down":
		m.cursor = min(count-1, m.cursor+columns)
	case "left":
		if m.cursor%columns > 0 {
			m.cursor--
		}
	case "right":
		if m.cursor%columns < columns-1 {
			m.cursor = min(count-1, m.cursor+1)
		}
	case "pgup":
		m.cursor = max(0, m.cursor-pageSize)
	case "pgdown":
		m.cursor = min(count-1, m.cursor+pageSize)
	case "home":
		m.cursor = 0
	case "end":
		m.cursor = count - 1
	}
	m.cursor = max(0, m.cursor)
}

func (m *macModel) selectAllResults() tea.Cmd {
	if m.screen != "browse" {
		return nil
	}
	rows := m.rows()
	ids := []string{}
	all := true
	for _, r := range rows {
		_, pkg := macPackage(r.ID)
		_, setting := macSetting(r.ID)
		if !pkg && !setting {
			continue
		}
		if m.selected["shell:configure"] && oneOf(r.ID, m.shellSettings.packages()) {
			continue
		}
		if !m.needsSelection(r.ID) {
			continue
		}
		ids = append(ids, r.ID)
		if !m.selected[r.ID] {
			all = false
		}
	}
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		if all {
			delete(m.selected, id)
		} else {
			m.selected[id] = true
		}
	}
	action := "Selected"
	if all {
		action = "Deselected"
	}
	return m.flash(fmt.Sprintf("%s all %d results. Other selections kept.", action, len(ids)))
}
