package main

import "charm.land/lipgloss/v2"

// macStatus is the small, shared vocabulary for state shown throughout the Mac
// TUI. Rendering code uses these labels and styles instead of assigning colours
// ad hoc per screen.
type macStatus int

const (
	macStatusNeutral macStatus = iota
	macStatusInfo
	macStatusSuccess
	macStatusWarning
	macStatusDanger
	macStatusMuted
)

var (
	macInfo    = lipgloss.NewStyle().Foreground(adaptiveColor{Light: "#086a9a", Dark: "#7dd3fc"})
	macSuccess = lipgloss.NewStyle().Foreground(adaptiveColor{Light: "#237342", Dark: "#86d9a0"})
	macWarning = lipgloss.NewStyle().Foreground(adaptiveColor{Light: "#9a5b00", Dark: "#fbbf24"})
	macDanger  = lipgloss.NewStyle().Foreground(adaptiveColor{Light: "#a11b42", Dark: "#fb7185"})
)

func macStatusStyle(status macStatus) lipgloss.Style {
	switch status {
	case macStatusInfo:
		return macInfo
	case macStatusSuccess:
		return macSuccess
	case macStatusWarning:
		return macWarning
	case macStatusDanger:
		return macDanger
	case macStatusMuted:
		return sMuted
	default:
		return sText
	}
}

func macInventoryStatus(state string) (string, macStatus, bool) {
	switch state {
	case "installed":
		return "Installed", macStatusSuccess, true
	case "already set":
		return "Configured", macStatusSuccess, true
	case "outside Homebrew":
		return "External", macStatusInfo, true
	case "needs Homebrew":
		return "Needs Homebrew", macStatusWarning, true
	case "inspection failed":
		return "Check failed", macStatusDanger, true
	default:
		return "", macStatusNeutral, false
	}
}

func macOutcomeStatus(state string) (string, macStatus) {
	switch state {
	case "installed":
		return "Installed", macStatusSuccess
	case "already present":
		return "Already present", macStatusInfo
	case "skipped":
		return "Skipped", macStatusWarning
	case "failed":
		return "Failed", macStatusDanger
	default:
		return "Pending", macStatusMuted
	}
}

func macStatusBadge(label string, status macStatus) string {
	return macStatusStyle(status).Render(" " + label)
}
