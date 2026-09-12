package main

// macScreen names the finite set of top-level Mac TUI views. Keeping navigation
// APIs typed prevents screen values from being confused with category, package,
// and row identifiers as the catalogue evolves.
type macScreen string

const (
	macScreenMenu            macScreen = "menu"
	macScreenCategories      macScreen = "categories"
	macScreenBrowse          macScreen = "browse"
	macScreenReview          macScreen = "review"
	macScreenPresets         macScreen = "presets"
	macScreenUpdates         macScreen = "updates"
	macScreenUpdateConfirm   macScreen = "update-confirm"
	macScreenInstall         macScreen = "install"
	macScreenSummary         macScreen = "summary"
	macScreenBootstrap       macScreen = "bootstrap"
	macScreenRestore         macScreen = "restore"
	macScreenTerminal        macScreen = "terminal"
	macScreenTerminalRestore macScreen = "terminal-restore"
	macScreenShell           macScreen = "shell"
	macScreenAppConfigs      macScreen = "app-configs"
	macScreenRaycast         macScreen = "raycast"
)

func (s macScreen) searchable() bool { return s != macScreenReview }

func (s macScreen) supportsDetails() bool {
	switch s {
	case macScreenMenu, macScreenBrowse, macScreenCategories, macScreenReview,
		macScreenPresets, macScreenTerminal, macScreenShell, macScreenAppConfigs,
		macScreenRaycast:
		return true
	default:
		return false
	}
}

// macEventKind describes messages emitted by an installation session. The
// session is the sole producer and the TUI update loop is the sole consumer.
type macEventKind string

const (
	macEventPlan     macEventKind = "plan"
	macEventActive   macEventKind = "active"
	macEventLog      macEventKind = "log"
	macEventResult   macEventKind = "result"
	macEventFailure  macEventKind = "failure"
	macEventFatal    macEventKind = "fatal"
	macEventFinished macEventKind = "finished"
	macEventClosed   macEventKind = "closed"
)
