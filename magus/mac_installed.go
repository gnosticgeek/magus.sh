package main

import (
	"charm.land/lipgloss/v2"
	"github.com/muesli/termenv"
)

func installedBadge(state string) string {
	label := ""
	switch state {
	case "installed":
		label = " installed"
	case "already set":
		label = " set"
	case "outside Homebrew":
		label = " external"
	default:
		return ""
	}
	mark := "✓"
	if terminalProfile == termenv.Ascii {
		mark = "+"
	}
	return lipgloss.NewStyle().Foreground(adaptiveColor{Light: "#237342", Dark: "#86d9a0"}).Render(" " + mark + label)
}
