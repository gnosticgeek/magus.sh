package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type searchHit struct {
	Stage *Stage
	Cmd   *Cmd
}

// runSearch matches a substring across commands.
// If stageID is set, results are scoped to that stage.
func (m Model) runSearch(query, stageID string) []searchHit {
	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]searchHit, 0, 16)
	for _, st := range m.cat.Stages {
		if stageID != "" && st.ID != stageID {
			continue
		}
		for _, c := range st.Items {
			if q == "" || strings.Contains(strings.ToLower(c.Title), q) ||
				strings.Contains(strings.ToLower(c.Summary), q) {
				out = append(out, searchHit{Stage: st, Cmd: c})
				if len(out) >= 32 {
					return out
				}
			}
		}
	}
	return out
}

func (m Model) viewSearch() string {
	header := sBright.Render("searching: ") + sAccent.Render(m.searchQuery) + sCursor.Render("▌")
	body := strings.Builder{}
	if len(m.searchResults) == 0 {
		body.WriteString(sWarn.Render(fmt.Sprintf("  no spells match %q", m.searchQuery)))
	} else {
		shown := 0
		for i, hit := range m.searchResults {
			if shown >= 12 {
				body.WriteString(sDim.Render(fmt.Sprintf("  … +%d more", len(m.searchResults)-shown)))
				body.WriteByte('\n')
				break
			}
			focused := i == m.cursor
			cursor := "  "
			if focused {
				cursor = sCursor.Render("❯ ")
			}
			mark := sMuted.Render("◯")
			if m.picked[hit.Cmd.ID] {
				mark = sCheck.Render("◉")
			}
			scope := sDim.Render(fmt.Sprintf("[%s %s]", hit.Stage.Num, strings.ToLower(hit.Stage.Short)))
			var title string
			if focused {
				title = sBright.Render(hit.Cmd.Title)
			} else if m.picked[hit.Cmd.ID] {
				title = sText.Render(hit.Cmd.Title)
			} else {
				title = sMuted.Render(hit.Cmd.Title)
			}
			body.WriteString(cursor)
			body.WriteString(mark)
			body.WriteByte(' ')
			body.WriteString(scope)
			body.WriteByte(' ')
			body.WriteString(title)
			body.WriteByte('\n')
			shown++
		}
	}

	footer := sDim.Render(fmt.Sprintf("%d results · %d picked total", len(m.searchResults), m.totalPicked()))
	hints := []Hint{
		{Key: "enter", Action: "toggle", Kind: HintPrimary},
		{Key: "space", Action: "toggle", Kind: HintNormal},
		{Key: "↑↓", Action: "move", Kind: HintNormal},
		{Key: "esc", Action: "back", Kind: HintNormal},
		{Key: "r", Action: "reset", Kind: HintSystem},
	}
	return wrapScreen(m, header, body.String(), footer, statusBar(hints))
}

func (m Model) keySearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.pickView = m.priorView
		// restore reasonable cursor
		m.cursor = 0
		return m, nil
	case tea.KeyBackspace:
		if len(m.searchQuery) > 0 {
			runes := []rune(m.searchQuery)
			m.searchQuery = string(runes[:len(runes)-1])
		}
		stageScope := ""
		if m.priorView == PickStage {
			stageScope = m.currentStageID
		}
		m.searchResults = m.runSearch(m.searchQuery, stageScope)
		if m.cursor >= len(m.searchResults) {
			m.cursor = 0
		}
		return m, nil
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.cursor < len(m.searchResults)-1 && m.cursor < 11 {
			m.cursor++
		}
		return m, nil
	case tea.KeyEnter, tea.KeySpace:
		if m.cursor < len(m.searchResults) {
			m.togglePick(m.searchResults[m.cursor].Cmd.ID)
		}
		return m, nil
	case tea.KeyRunes:
		r := msg.Runes
		if len(r) == 1 {
			ch := r[0]
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
				(ch >= '0' && ch <= '9') || ch == '-' || ch == ' ' {
				m.searchQuery += string(ch)
				stageScope := ""
				if m.priorView == PickStage {
					stageScope = m.currentStageID
				}
				m.searchResults = m.runSearch(m.searchQuery, stageScope)
				m.cursor = 0
			}
		}
	}
	return m, nil
}
