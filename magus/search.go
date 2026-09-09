package main

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

type searchHit struct {
	Stage *Stage
	Cmd   *Cmd
}

func (h searchHit) FilterValue() string { return h.Cmd.Title + " " + h.Cmd.Summary }

func (m Model) runSearch(query, stageID string) []searchHit {
	var hits []searchHit
	var targets []string
	for _, st := range m.cat.Stages {
		if stageID != "" && st.ID != stageID {
			continue
		}
		for _, c := range st.Items {
			h := searchHit{st, c}
			hits = append(hits, h)
			targets = append(targets, h.FilterValue())
		}
	}
	if strings.TrimSpace(query) == "" {
		return hits
	}
	var result []searchHit
	for _, rank := range list.DefaultFilter(query, targets) {
		result = append(result, hits[rank.Index])
	}
	return result
}

func newSearchInput() textinput.Model {
	input := textinput.New()
	input.Prompt = "Search: "
	input.Placeholder = "Apps, tools and settings"
	input.CharLimit = 100
	return input
}
func (m Model) startSearch(prior PickView) (tea.Model, tea.Cmd) {
	m.priorView = prior
	m.pickView = PickSearch
	m.searchQuery = ""
	m.searchInput = newSearchInput()
	m.searchInput.SetWidth(max(1, m.width-10))
	m.cursor = 0
	m.refreshSearch()
	return m, m.searchInput.Focus()
}
func (m *Model) refreshSearch() {
	scope := ""
	if m.priorView == PickStage {
		scope = m.currentStageID
	}
	m.searchQuery = m.searchInput.Value()
	m.searchResults = m.runSearch(m.searchQuery, scope)
	m.cursor = min(m.cursor, max(0, len(m.searchResults)-1))
}

type searchDelegate struct{ picked map[string]bool }

func (d searchDelegate) Height() int                         { return 1 }
func (d searchDelegate) Spacing() int                        { return 0 }
func (d searchDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }
func (d searchDelegate) Render(w io.Writer, l list.Model, index int, item list.Item) {
	h := item.(searchHit)
	mark := "[ ]"
	if d.picked[h.Cmd.ID] {
		mark = "[x]"
	}
	cursor := "  "
	style := sText
	if index == l.Index() {
		cursor = "> "
		style = sBright
	}
	fmt.Fprint(w, style.Render(ansi.Truncate(fmt.Sprintf("%s%s [%s] %s", cursor, mark, h.Stage.Short, h.Cmd.Title), max(1, l.Width()), "…")))
}
func (m Model) searchList() list.Model {
	items := make([]list.Item, len(m.searchResults))
	for i, h := range m.searchResults {
		items[i] = h
	}
	l := list.New(items, searchDelegate{m.picked}, max(1, m.width), max(1, m.height-10))
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.Select(m.cursor)
	return l
}
func (m Model) viewSearch() string {
	input := m.searchInput
	input.SetWidth(max(1, m.width-10))
	l := m.searchList()
	footer := fmt.Sprintf("%d results · %d picked · page %d/%d", len(m.searchResults), m.totalPicked(), l.Paginator.Page+1, max(1, l.Paginator.TotalPages))
	return wrapScreen(m, input.View(), l.View(), footer, statusBar([]Hint{{Key: "enter", Action: "toggle", Kind: HintPrimary}, {Key: "↑↓", Action: "move"}, {Key: "pgup/pgdown", Action: "page"}, {Key: "esc", Action: "back"}}))
}
func (m Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "esc":
			m.pickView = m.priorView
			m.cursor = 0
			m.searchInput.Blur()
			return m, nil
		case "enter":
			if len(m.searchResults) > 0 {
				m.togglePick(m.searchResults[m.cursor].Cmd.ID)
			}
			return m, nil
		case "up", "down", "pgup", "pgdown":
			l := m.searchList()
			l, cmd := l.Update(msg)
			m.cursor = l.Index()
			return m, cmd
		}
	}
	before := m.searchInput.Value()
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	if m.searchInput.Value() != before {
		m.cursor = 0
		m.refreshSearch()
	}
	return m, cmd
}
func (m Model) keySearch(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) { return m.updateSearch(msg) }
