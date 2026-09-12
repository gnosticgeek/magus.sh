package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type macUpgrade struct{ ID, Kind, Installed, Available string }
type macUpdateReview struct {
	items          []macUpgrade
	table          table.Model
	loading, ready bool
	err            error
	generation     asyncGeneration
	cancel         context.CancelFunc
}
type macUpdatesChecked struct {
	items      []macUpgrade
	err        error
	generation asyncGeneration
}
type macMetadataRefreshed struct{ err error }

func parseMacOutdated(data string) ([]macUpgrade, error) {
	type entry struct {
		Name      string   `json:"name"`
		Installed []string `json:"installed_versions"`
		Current   string   `json:"current_version"`
		Pinned    bool     `json:"pinned"`
	}
	var payload struct {
		Formulae []entry `json:"formulae"`
		Casks    []entry `json:"casks"`
	}
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return nil, fmt.Errorf("cannot read Homebrew update list: %w", err)
	}
	if payload.Formulae == nil || payload.Casks == nil {
		return nil, fmt.Errorf("incomplete Homebrew update list")
	}
	valid := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9@+._/-]*$`)
	items := []macUpgrade{}
	seen := map[string]bool{}
	for _, group := range []struct {
		kind    string
		entries []entry
	}{{"formula", payload.Formulae}, {"cask", payload.Casks}} {
		for _, e := range group.entries {
			if e.Pinned {
				continue
			}
			if !valid.MatchString(e.Name) || len(e.Installed) == 0 || e.Current == "" || strings.Contains(e.Name, "..") {
				return nil, fmt.Errorf("invalid Homebrew update entry")
			}
			id := group.kind + ":" + e.Name
			if seen[id] {
				return nil, fmt.Errorf("duplicate Homebrew update entry")
			}
			seen[id] = true
			items = append(items, macUpgrade{e.Name, group.kind, strings.Join(e.Installed, ", "), e.Current})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Kind+items[i].ID < items[j].Kind+items[j].ID })
	return items, nil
}

func readMacUpdates(ctx context.Context, brew string) ([]macUpgrade, error) {
	if brew == "" {
		return nil, fmt.Errorf("Homebrew is not installed")
	}
	output, err := runMacCommand(ctx, nil, brew, "outdated", "--json=v2")
	if err != nil {
		return nil, err
	}
	return parseMacOutdated(output)
}

func sameMacUpdates(a, b []macUpgrade) bool { return reflect.DeepEqual(a, b) }

func (m *macModel) acceptUpdateReview(v macUpdatesChecked) {
	m.updateReview.loading = false
	m.updateReview.err = v.err
	m.updateReview.ready = v.err == nil
	m.updateReview.items = v.items
	rows := make([]table.Row, 0, len(v.items))
	for _, x := range v.items {
		rows = append(rows, table.Row{cleanLog(x.ID), cleanLog(x.Installed), cleanLog(x.Available)})
	}
	m.updateReview.table = table.New(table.WithColumns([]table.Column{{Title: "App / tool", Width: 24}, {Title: "Installed", Width: 16}, {Title: "Available", Width: 16}}), table.WithRows(rows), table.WithFocused(true), table.WithHeight(6))
	styles := table.DefaultStyles()
	styles.Header = macAccent.Padding(0, 1)
	styles.Selected = macAccent
	m.updateReview.table.SetStyles(styles)
}

func (m *macModel) checkUpdates(refresh bool) tea.Cmd {
	if refresh {
		if m.preview {
			m.notice = "Preview only — metadata refresh is disabled."
			return nil
		}
		brew := findBrew()
		if brew == "" {
			m.notice = "Homebrew is not installed."
			return nil
		}
		unlock, err := acquireMacLock(m.paths)
		if err != nil {
			m.notice = err.Error()
			return nil
		}
		m.updateReview.ready = false
		return tea.Exec(&macUpdateCommand{brew: brew, timeout: 5 * time.Minute, refresh: true}, func(err error) tea.Msg { unlock(); return macMetadataRefreshed{err} })
	}
	if m.updateReview.cancel != nil {
		m.updateReview.cancel()
	}
	generation := m.updateReview.generation.next()
	m.updateReview.loading, m.updateReview.ready = true, false
	m.updateReview.err = nil
	m.updateReview.items = nil
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	m.updateReview.cancel = cancel
	brew := findBrew()
	return func() tea.Msg {
		defer cancel()
		items, err := readMacUpdates(ctx, brew)
		return macUpdatesChecked{items, err, generation}
	}
}

func (m *macModel) resizeUpdateTable(width, height int) {
	nameWidth := max(10, width/3)
	versionWidth := max(8, (width-nameWidth-6)/2)
	m.updateReview.table.SetColumns([]table.Column{{Title: "App / tool", Width: nameWidth}, {Title: "Installed", Width: versionWidth}, {Title: "Available", Width: versionWidth}})
	m.updateReview.table.SetWidth(width)
	m.updateReview.table.SetHeight(max(2, height))
}

func (m *macModel) updateReviewView(width, height int) string {
	if m.updateReview.loading {
		return m.spinner.View() + " Checking installed versions…"
	}
	if m.updateReview.err != nil {
		return "Could not check updates.\n\n" + lipgloss.NewStyle().Width(width).Render(cleanLog(m.updateReview.err.Error())) + "\n\nPress r to refresh and retry; Escape returns."
	}
	if !m.updateReview.ready {
		return "Open Update all to inspect available versions."
	}
	if len(m.updateReview.items) == 0 {
		return "No eligible updates in local Homebrew metadata.\n\nPress r to refresh metadata and check again.\nPinned and self-updating apps follow Homebrew exclusions."
	}
	m.resizeUpdateTable(width, max(2, height-9))
	row := m.updateReview.table.Cursor()
	rows := make([]table.Row, 0, len(m.updateReview.items))
	for i, x := range m.updateReview.items {
		mark := "  "
		if i == row {
			mark = "> "
		}
		rows = append(rows, table.Row{mark + cleanLog(x.ID), cleanLog(x.Installed), cleanLog(x.Available)})
	}
	m.updateReview.table.SetRows(rows)
	detail := ""
	if row >= 0 && row < len(m.updateReview.items) {
		x := m.updateReview.items[row]
		detail = x.ID + " (" + x.Kind + ")\n" + cleanLog(x.Installed) + " → " + cleanLog(x.Available)
	}
	return fmt.Sprintf("Review updates · %d/%d · local Homebrew metadata\n", row+1, len(m.updateReview.items)) + m.updateReview.table.View() + "\n" + lipgloss.NewStyle().Width(width).Render(detail) + "\n\nClose affected apps. Dependencies may also change.\nEnter opens confirmation; r refreshes metadata."
}
