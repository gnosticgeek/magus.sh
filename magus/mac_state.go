package main

// macScreen names the finite set of top-level Mac TUI views. Keeping navigation
// APIs typed prevents screen values from being confused with category, package,
// and row identifiers as the catalogue evolves.
type macScreen string

const (
	macScreenMenu            macScreen = "menu"
	macScreenDeveloper       macScreen = "developer"
	macScreenAgents          macScreen = "agents"
	macScreenSkills          macScreen = "skills"
	macScreenCategories      macScreen = "categories"
	macScreenBrowse          macScreen = "browse"
	macScreenBasket          macScreen = "basket"
	macScreenReview          macScreen = "review"
	macScreenPresets         macScreen = "presets"
	macScreenUpdates         macScreen = "updates"
	macScreenUpdateConfirm   macScreen = "update-confirm"
	macScreenSelfUpdate      macScreen = "self-update"
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

func (s macScreen) searchable() bool { return s != macScreenReview && s != macScreenBasket }

func (s macScreen) supportsDetails() bool {
	switch s {
	case macScreenMenu, macScreenDeveloper, macScreenAgents, macScreenSkills, macScreenBrowse, macScreenCategories, macScreenBasket, macScreenReview,
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

// macPreviewResult is the publication boundary for any preview that later
// needs I/O. Results carry both the focused target and a generation so moving
// focus can invalidate work that is otherwise perfectly valid but obsolete.
type macPreviewResult struct {
	target     string
	content    string
	generation asyncGeneration
}

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
