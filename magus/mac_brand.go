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
var magusRainbowDark = []string{"#fb7185", "#fdba74", "#fde68a", "#86efac", "#67e8f9", "#93c5fd", "#c4b5fd", "#f0abfc", "#fb7185"}
var magusRainbowLight = []string{"#be123c", "#c2410c", "#a16207", "#15803d", "#0e7490", "#1d4ed8", "#6d28d9", "#a21caf", "#be123c"}
var macAccent = lipgloss.NewStyle().Foreground(adaptiveColor{Light: "#6740b8", Dark: "#b5a0ff"}).Bold(true)

func auroraText(text string) string {
	return gradientText(text, macAuroraLight, macAuroraDark)
}

func rainbowText(text string) string {
	return rainbowTextAt(text, 0)
}

func rainbowTextAt(text string, offset int) string {
	return gradientTextAt(text, magusRainbowLight, magusRainbowDark, offset)
}

func gradientText(text string, lightStops, darkStops []string) string {
	return gradientTextAt(text, lightStops, darkStops, 0)
}

func gradientTextAt(text string, lightStops, darkStops []string, offset int) string {
	if terminalProfile == termenv.Ascii {
		return text
	}
	lines := strings.Split(text, "\n")
	stops := darkStops
	if lightBackground.Load() {
		stops = lightStops
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
				row.WriteString(lipgloss.NewStyle().Bold(true).Foreground(gradient[(x+y+offset)%len(gradient)]).Render(string(r)))
			}
			x += lipgloss.Width(string(r))
		}
		lines[y] = row.String()
	}

	return strings.Join(lines, "\n")
}

// Plain ASCII keeps the Magus sigil and wordmark readable in every terminal.
const macSigil = "/M\\"

var macWordmark = strings.Join(magusWordmark, "\n")

func (m *macModel) headerView(width int) string {
	title := "magus"
	home := m.screen == macScreenMenu && !m.showHelp
	if home {
		title = macSigil + "  MAGUS"
	}
	if m.preview {
		title += " / PREVIEW — no changes"
	}
	metadata := sMuted.Render("Magus " + buildVersion + "  ·  " + m.inventory.osVersion)
	if !home && m.breadcrumb() != "" {
		metadata += "\n" + sDim.Render(m.breadcrumb())
	}
	if home && m.height >= 22 && width >= 42 {
		return rainbowTextAt(macWordmark, m.rainbowOffset) + "\n" + metadata
	}
	return rainbowText(title) + "\n" + metadata
}
