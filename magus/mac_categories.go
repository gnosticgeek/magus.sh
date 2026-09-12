package main

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
)

type macAppCategory struct {
	ID, Name, Summary, Light, Dark string
	Packages                       []string
}

// Every desktop app has one home; search and the basket span all categories.
var macAppCategories = buildMacAppCategories()

func buildMacAppCategories() []macAppCategory {
	groups := []macAppCategory{
		{"browsers", "Browsers", "Choose your window onto the web.", "#086a9a", "#7dd3fc", []string{"firefox", "brave-browser", "google-chrome"}},
		{"development", "Developer tools", "Editors and terminals for building things.", "#6740b8", "#b5a0ff", []string{"ghostty", "iterm2", "visual-studio-code", "zed"}},
		{"ai", "AI & local models", "Explore models on your own Mac. Model downloads are separate.", "#a52b9d", "#f0abfc", []string{"ollama-app", "lm-studio"}},
		{"productivity", "Productivity", "Notes, window management, launchers and everyday utilities.", "#347866", "#99dec6", []string{"obsidian", "rectangle", "raycast", "the-unarchiver"}},
		{"media", "Media", "A comfortable home for music and video.", "#a46428", "#eac080", []string{"vlc", "iina"}},
		{"communication", "Communication", "Private conversations and shared communities.", "#9b4770", "#f1aacb", []string{"signal", "discord"}},
	}
	additions := map[string][]string{
		"browsers":      {"aside"},
		"development":   {},
		"ai":            {"chatgpt", "claude", "fluidvoice", "grok-bot", "hermes-desktop", "unsloth"},
		"productivity":  {"notion"},
		"media":         {"downie", "handbrake-app", "jellyfin-media-player", "moonfin"},
		"communication": {},
	}
	for i := range groups {
		groups[i].Packages = append(groups[i].Packages, additions[groups[i].ID]...)
	}
	groups[2].Name = "AI & LLMs"
	groups[2].Summary = "Assistants, local models and voice tools."
	groups[4].Name = "Video & Media"
	groups = append(groups,
		macAppCategory{"audio", "Audio & Music", "Explore audio & music apps.", "#347866", "#99dec6", []string{"audacity", "spotify"}},
		macAppCategory{"design", "Design & Graphics", "Explore design & graphics apps.", "#347866", "#99dec6", []string{"affinity", "bambu-studio", "shottr", "upscayl"}},
		macAppCategory{"cloud", "Cloud & Storage", "Explore cloud & storage apps.", "#347866", "#99dec6", []string{"proton-drive"}},
		macAppCategory{"security", "Security & Privacy", "Explore security & privacy apps.", "#347866", "#99dec6", []string{"protonvpn"}},
		macAppCategory{"games", "Games", "Explore games apps.", "#347866", "#99dec6", []string{"es-de", "retroarch-metal"}},
		macAppCategory{"menubar", "Menu Bar", "Monitors, menu organisers and everyday controls.", "#347866", "#99dec6", []string{"stats", "jordanbaird-ice", "mos", "hiddenbar", "swiftbar", "thaw", "codexbar", "vorssaint", "aldente"}},
		macAppCategory{"utilities", "Utilities", "Explore utilities apps.", "#347866", "#99dec6", []string{"balenaetcher", "caskhub", "dockflow", "keka", "monocle-app"}},
	)
	return groups
}

func appCategory(id string) (macAppCategory, bool) {
	for _, group := range macAppCategories {
		if group.ID == id {
			return group, true
		}
	}
	return macAppCategory{}, false
}

func packageCategory(id string) (macAppCategory, bool) {
	for _, group := range macAppCategories {
		if oneOf(id, group.Packages) {
			return group, true
		}
	}
	return macAppCategory{}, false
}

func (g macAppCategory) style() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(adaptiveColor{Light: g.Light, Dark: g.Dark})
}

func (m *macModel) categoryRows() []macRow {
	var rows []macRow
	for _, group := range macAppCategories {
		picked := 0
		var names []string
		for _, id := range group.Packages {
			if m.selected[id] {
				picked++
			}
			if p, ok := macPackage(id); ok {
				names = append(names, p.Name)
			}
		}
		rows = append(rows, macRow{ID: group.ID, Name: group.Name, Summary: group.Summary,
			Source: fmt.Sprintf("%d apps  /  %d selected", len(group.Packages), picked), Note: strings.Join(names, " · ")})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return strings.ToLower(rows[i].Name) < strings.ToLower(rows[j].Name)
	})
	return rows
}

func (m *macModel) rowStyle(id string) lipgloss.Style {
	if m.screen == macScreenCategories {
		if g, ok := appCategory(id); ok {
			return g.style()
		}
	}
	if g, ok := packageCategory(id); ok {
		return g.style()
	}
	return macAccent
}
