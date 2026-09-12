package main

import "github.com/muesli/termenv"

func installedBadge(state string) string {
	label, status, ok := macInventoryStatus(state)
	if !ok {
		return ""
	}
	mark := "✓"
	if terminalProfile == termenv.Ascii {
		mark = "+"
	}
	return macStatusStyle(status).Render(" " + mark + " " + label)
}
