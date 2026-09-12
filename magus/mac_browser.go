package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (r macRow) FilterValue() string { return r.Name + " " + r.ID + " " + r.Summary + " " + r.Source }

type macRowDelegate struct{ owner *macModel }

func (d macRowDelegate) Height() int                         { return 1 }
func (d macRowDelegate) Spacing() int                        { return 0 }
func (d macRowDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }
func (d macRowDelegate) Render(w io.Writer, l list.Model, index int, item list.Item) {
	m := d.owner
	r := item.(macRow)
	if m.screen == macScreenMenu {
		marker := "  "
		if index == l.Index() {
			marker = "› "
		}
		label := fmt.Sprintf("%s%d.  %s", marker, index+1, r.Name)
		style := sText
		if index == l.Index() {
			style = macAccent.Background(adaptiveColor{Light: "#eee8fa", Dark: "#302b45"})
		} else if r.ID == "review" && len(m.selected) > 0 {
			style = macAccent
		}
		if r.ID == "self-update" && m.magUpdateAvailable {
			available := lipgloss.NewStyle().Foreground(adaptiveColor{Light: "#237342", Dark: "#86d9a0"}).Bold(true).Render(" · update available")
			label = ansi.Truncate(label, max(1, l.Width()-lipgloss.Width(available)), "…")
			fmt.Fprint(w, style.Width(l.Width()-lipgloss.Width(available)).Render(label)+available)
			return
		}
		fmt.Fprint(w, style.Width(l.Width()).Render(ansi.Truncate(label, l.Width(), "…")))
		return
	}
	mark := "  "
	if index == l.Index() {
		mark = "> "
	}
	check := ""
	blocked := m.selectionBlockReason(r.ID)
	if (m.screen == macScreenBrowse || m.screen == macScreenReview) && r.ID != "restore" && m.needsSelection(r.ID) && blocked == "" {
		check = "[ ] "
		if m.selected[r.ID] {
			check = "[x] "
		}
	}
	label := mark + check + r.Name
	if m.screen == macScreenCategories {
		group, _ := appCategory(r.ID)
		picked := 0
		for _, id := range group.Packages {
			if m.selected[id] {
				picked++
			}
		}
		label = fmt.Sprintf("%s%-20s %d/%d >", mark, r.Name, picked, len(group.Packages))
	}
	badge := installedBadge(m.inventory.states[r.ID])
	if blocked != "" {
		badge = macStatusBadge("Unavailable", macStatusWarning)
	}
	label = ansi.Truncate(label, max(1, l.Width()-lipgloss.Width(badge)), "…")
	style := sText
	if index == l.Index() {
		style = m.rowStyle(r.ID)
	} else if m.screen == macScreenCategories {
		style = m.rowStyle(r.ID).Bold(false)
	} else if blocked != "" {
		style = sMuted
	}
	fmt.Fprint(w, style.Render(label)+badge)
}

func (m *macModel) browserList(rows []macRow, width, height int) list.Model {
	items := make([]list.Item, len(rows))
	for i := range rows {
		items[i] = rows[i]
	}
	l := list.New(items, macRowDelegate{m}, width, max(1, height))
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.KeyMap.CursorUp = key.NewBinding(key.WithKeys("up"))
	l.KeyMap.CursorDown = key.NewBinding(key.WithKeys("down"))
	l.KeyMap.PrevPage = key.NewBinding(key.WithKeys("pgup"))
	l.KeyMap.NextPage = key.NewBinding(key.WithKeys("pgdown"))
	l.KeyMap.GoToStart = key.NewBinding(key.WithKeys("home"))
	l.KeyMap.GoToEnd = key.NewBinding(key.WithKeys("end"))
	l.Select(min(m.cursor, max(0, len(rows)-1)))
	return l
}

func fuzzyMacRows(rows []macRow, query string) []macRow {
	if strings.TrimSpace(query) == "" {
		return rows
	}
	targets := make([]string, len(rows))
	for i, r := range rows {
		targets[i] = r.FilterValue()
	}
	ranks := list.DefaultFilter(query, targets)
	result := make([]macRow, 0, len(ranks))
	for _, rank := range ranks {
		result = append(result, rows[rank.Index])
	}
	return result
}

type macNoticeExpired struct {
	generation asyncGeneration
	text       string
}

func (m *macModel) flash(text string) tea.Cmd {
	generation := m.noticeGeneration.next()
	m.notice = text
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg { return macNoticeExpired{generation, text} })
}

func (m *macModel) detailsView(text string, width, height int) string {
	content := renderMarkdown(text, width)
	m.previewPane.SetWidth(width)
	m.previewPane.SetHeight(max(1, height-1))
	if content != m.previewContent {
		m.previewContent = content
		m.previewPane.SetContent(content)
		m.previewPane.GotoTop()
	}
	label := "Tab: focus details"
	if m.details {
		label = "DETAILS · ↑↓ scroll · Tab: list"
	}
	return macAccent.Render(ansi.Truncate(label, width, "…")) + "\n" + m.previewPane.View()
}

func (m *macModel) contextualKeys() macKeys {
	if m.details {
		return hints("↑↓", "scroll details", "pgup/pgdown", "page", "tab/esc", "back to list")
	}
	if m.screen == macScreenInstall {
		if m.failed {
			return hints("r", "retry failure", "s", "skip failure", "l", "logs", "q/esc", "stop")
		}
		return hints("l", "toggle logs", "↑↓", "scroll logs", "home/end", "log start/live output", "q/esc", "stop")
	}
	if m.screen == macScreenUpdates {
		return hints("↑↓", "choose row", "pgup/pgdown", "page", "enter", "review confirmation", "r", "refresh metadata", "esc", "back")
	}
	if m.screen == macScreenUpdateConfirm {
		return hints("enter", "update reviewed packages", "esc", "back to versions")
	}
	if m.screen == macScreenShell {
		return hints("↑↓", "move", "enter/space", "toggle", "tab", "details", "esc", "menu")
	}
	if m.screen == macScreenTerminal || m.screen == macScreenAppConfigs || m.screen == macScreenRaycast {
		return hints("↑↓", "move", "enter", "choose setup", "tab", "details", "esc", "menu")
	}
	if m.screen == macScreenMenu || m.screen == macScreenCategories {
		return hints("↑↓", "move", "pgup/pgdown", "page", "home/end", "first/last", "enter", "open", "tab", "focus details", "/", "search", "esc", "back", "q", "quit")
	}
	if m.screen == macScreenPresets {
		return hints("↑↓", "move", "enter", "add preset", "tab", "details", "esc", "back")
	}
	if m.screen == macScreenRestore {
		return hints("enter", "restore recorded settings", "esc", "back")
	}
	if m.screen == macScreenReview {
		return hints("↑↓", "move", "space", "remove from basket", "enter", "install selected", "esc", "back")
	}
	if m.screen == macScreenSummary {
		return hints("↑↓", "scroll", "enter", "menu", "l", "logs", "q", "quit")
	}
	return hints("←↑↓→", "move", "pgup/pgdown", "page", "home/end", "first/last", "enter/space", "select item", "ctrl+s", "select/deselect all results", "f", "filter catalogue", "tab", "focus details", "/", "fuzzy search", "esc", "back", "?", "help")
}
