package main

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/muesli/termenv"
)

// Matches src/components/AuroraText.astro, with darker tones for light terminals.
var macAuroraDark = []string{"#c4b5fd", "#f0abfc", "#a78bfa", "#7dd3fc", "#e9d5ff", "#c4b5fd"}
var macAuroraLight = []string{"#6740b8", "#a52b9d", "#7045ba", "#086a9a", "#8550ad", "#6740b8"}
var macAccent = lipgloss.NewStyle().Foreground(adaptiveColor{Light: "#6740b8", Dark: "#b5a0ff"}).Bold(true)

func auroraText(text string) string {
	if terminalProfile == termenv.Ascii {
		return text
	}
	lines := strings.Split(text, "\n")
	stops := macAuroraDark
	if lightBackground.Load() {
		stops = macAuroraLight
	}
	colours := make([]color.Color, len(stops))
	for i, stop := range stops {
		colours[i] = lipgloss.Color(stop)
	}
	width := max(1, lipgloss.Width(text))
	gradient := lipgloss.Blend1D(max(width, len(stops)), colours...)
	for y, line := range lines {
		var row strings.Builder
		x := 0
		for _, r := range line {
			if r == ' ' {
				row.WriteRune(r)
			} else {
				row.WriteString(lipgloss.NewStyle().Bold(true).Foreground(gradient[min(x+y, len(gradient)-1)]).Render(string(r)))
			}
			x += lipgloss.Width(string(r))
		}
		lines[y] = row.String()
	}

	return strings.Join(lines, "\n")
}

// Plain ASCII keeps the wordmark readable with the terminal's existing font.
const macWordmark = `    __  ___  ___    ______ __  __ _____
   /  |/  / /   |  / ____// / / // ___/
  / /|_/ / / /| | / / __ / / / / \__ \
 / /  / / / ___ |/ /_/ // /_/ / ___/ /
/_/  /_/ /_/  |_|\____/ \____/ /____/`

func (m *macModel) headerView(width int) string {
	title := "magus"
	home := m.screen == macScreenMenu && !m.showHelp
	if home {
		title = "<>  M A G U S"
	}
	if m.preview {
		title += " / PREVIEW — no changes"
	}
	metadata := sMuted.Render("Magus " + buildVersion + "  ·  " + m.inventory.osVersion)
	if home && m.height >= 24 && width >= max(60, lipgloss.Width(macWordmark)) {
		lines := strings.Split(macWordmark, "\n")
		seal := []string{"    /\\    ", "   /  \\   ", "  < /\\ >  ", "   \\  /   ", "    \\/    "}
		for i := range lines {
			lines[i] = seal[i] + "  " + lines[i]
		}
		tagline := "A little magic for your Mac."
		if m.preview {
			tagline += "  / PREVIEW — no changes"
		}
		return auroraText(strings.Join(lines, "\n")) + "\n" + sMuted.Render(tagline) + "\n" + metadata
	}
	return auroraText(title) + "\n" + metadata
}
