package main

import "charm.land/lipgloss/v2"

// A compact terminal keeps the ordinary confirmation page. On a larger screen,
// layers retain the reviewed versions behind the confirmation dialog.
func confirmationOverlay(background, message string, width, height int) string {
	box := lipgloss.NewStyle().Width(min(64, width-4)).Padding(1, 2).
		Border(lipgloss.RoundedBorder()).BorderForeground(colorAccent).
		Background(adaptiveColor{Light: "#ffffff", Dark: "#201d29"}).Render(message)
	bw, bh := lipgloss.Size(box)
	if bw > width || bh > height {
		return message
	}
	return lipgloss.NewCanvas(width, height).
		Compose(lipgloss.NewLayer(background)).
		Compose(lipgloss.NewLayer(box).X((width - bw) / 2).Y((height - bh) / 2)).Render()
}
