package main

import (
	"fmt"
	"strings"
)

// location captures all state needed to return to the exact list position.
// History remains typed, so categories and filters cannot be confused with a
// screen identifier.
func (m *macModel) location() macNavigationPoint {
	return macNavigationPoint{
		screen:          m.screen,
		category:        m.category,
		appGroup:        m.appGroup,
		cursor:          m.cursor,
		groupCursor:     m.groupCursor,
		catalogueFilter: m.catalogueFilter,
	}
}

func (m *macModel) restoreLocation(location macNavigationPoint) {
	m.screen = location.screen
	m.category = location.category
	m.appGroup = location.appGroup
	m.cursor = location.cursor
	m.groupCursor = location.groupCursor
	m.catalogueFilter = location.catalogueFilter
	m.details = false
}

func (m *macModel) openScreen(screen macScreen) {
	m.navigate(screen, m.category, m.appGroup, 0)
	m.details = false
}

func (m *macModel) back() bool {
	return m.goBack()
}

type macBasketCounts struct{ Apps, Tools, Fonts, Settings, Setups int }

func (m *macModel) basketCounts() macBasketCounts {
	var counts macBasketCounts
	for id := range m.selected {
		if p, ok := macPackage(id); ok {
			switch {
			case p.Kind == "formula":
				counts.Tools++
			case strings.HasPrefix(p.ID, "font-"):
				counts.Fonts++
			default:
				counts.Apps++
			}
			continue
		}
		if _, ok := macSetting(id); ok {
			counts.Settings++
		} else {
			counts.Setups++
		}
	}
	return counts
}

func (m *macModel) basketSummary() string {
	c := m.basketCounts()
	parts := []string{fmt.Sprintf("Basket %d", len(m.selected))}
	for _, item := range []struct {
		count int
		name  string
	}{{c.Apps, "app"}, {c.Tools, "tool"}, {c.Fonts, "font"}, {c.Settings, "setting"}, {c.Setups, "setup"}} {
		if item.count > 0 {
			name := item.name
			if item.count != 1 {
				name += "s"
			}
			parts = append(parts, fmt.Sprintf("%d %s", item.count, name))
		}
	}
	return strings.Join(parts, " · ")
}

func (m *macModel) searchScopeLabel() string {
	origin := navigationLabel(m.searchReturn)
	if origin == "" {
		origin = "Home"
	}
	return "All catalogue · from " + origin
}
