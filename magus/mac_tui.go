package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func stripTerminal(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\t' {
			return -1
		}
		if r == 127 {
			return -1
		}
		return r
	}, ansi.Strip(s))
}

type macKeys struct{ bindings []key.Binding }

func (k macKeys) ShortHelp() []key.Binding  { return k.bindings }
func (k macKeys) FullHelp() [][]key.Binding { return [][]key.Binding{k.bindings} }
func hints(pairs ...string) macKeys {
	var k macKeys
	for i := 0; i+1 < len(pairs); i += 2 {
		k.bindings = append(k.bindings, key.NewBinding(key.WithKeys(pairs[i]), key.WithHelp(pairs[i], pairs[i+1])))
	}
	return k
}

type macRow struct{ ID, Name, Summary, Source, Note string }
type macCatalogueFilter int

const (
	macCatalogueAll macCatalogueFilter = iota
	macCatalogueAvailable
	macCatalogueSelected
)

type macInventory struct {
	states    map[string]string
	brew      bool
	osVersion string
}
type macInventoryResult struct {
	inventory  macInventory
	generation asyncGeneration
}
type macBootstrap struct {
	path string
	err  error
}
type macBootstrapDone struct{ err error }
type macFinderDone struct{ err error }
type activityTickMsg time.Time

func tickActivity() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return activityTickMsg(t) })
}

type macModel struct {
	shellSettings                                     modernShell
	shellAutoPackages                                 map[string]bool
	paths                                             Paths
	path                                              string
	manifest                                          Manifest
	preview                                           bool
	timeout                                           time.Duration
	logText                                           string
	latestLog                                         string
	width, height                                     int
	screen                                            macScreen
	category, appGroup                                string
	searchReturnScreen                                macScreen
	catalogueFilter                                   macCatalogueFilter
	groupCursor, searchReturnCursor                   int
	cursor                                            int
	selected                                          map[string]bool
	search                                            textinput.Model
	searching, details, showHelp, showLogs, restoring bool
	help                                              help.Model
	spinner                                           spinner.Model
	progress                                          progress.Model
	logs                                              viewport.Model
	previewPane                                       viewport.Model
	previewContent                                    string
	browserHeight                                     int
	noticeGeneration, inventoryGeneration             asyncGeneration
	selfUpdateGeneration                              asyncGeneration
	magUpdateAvailable                                bool
	magUpdateLatest                                   string
	inventoryCancel                                   context.CancelFunc
	selfUpdateCancel                                  context.CancelFunc
	sessionGeneration                                 asyncGeneration
	updateReview                                      macUpdateReview
	inventory                                         macInventory
	inspecting                                        bool
	notice                                            string
	session                                           *macSession
	outcomes                                          []macOutcome
	active                                            int
	failed, stopping                                  bool
	needsFinder                                       bool
	started                                           time.Time
	bootstrapUnlock                                   func()
	bootstrapCancel                                   context.CancelFunc
	bootstrapFile                                     string
}

func newMacModel(paths Paths, path string, m Manifest, preview bool, timeout time.Duration) *macModel {
	input := textinput.New()
	input.Prompt = "/ "
	input.Placeholder = "Search apps, tools and settings"
	input.CharLimit = 100
	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = macAccent
	model := &macModel{paths: paths, path: path, manifest: m, preview: preview, timeout: timeout, width: 80, height: 24, screen: macScreenMenu, selected: map[string]bool{}, search: input, spinner: spin, help: help.New(), progress: progress.New(progress.WithColors(lipgloss.Color("#b5a0ff"), lipgloss.Color("#7dd3fc")), progress.WithoutPercentage()), logs: viewport.New(viewport.WithWidth(72), viewport.WithHeight(8)), inspecting: true, inventory: macInventory{states: map[string]string{}}}
	model.previewPane = viewport.New(viewport.WithWidth(32), viewport.WithHeight(10))
	model.browserHeight = 10
	for _, id := range append(append([]string{}, m.Mac.Packages...), m.Mac.Settings...) {
		model.selected[id] = true
	}
	for _, id := range m.Mac.AppConfigs {
		model.selected["config:"+id] = true
	}
	if m.Mac.Terminal != "" {
		model.selected[terminalID(m.Mac.Terminal)] = true
	}
	model.shellSettings.Commands = []string{"ls", "tree", "top", "cd", "fzf", "delta"}
	if m.Mac.ModernShell != nil {
		model.selected["shell:configure"] = true
		model.shellSettings = *m.Mac.ModernShell
		model.shellSettings.Commands = append([]string{}, m.Mac.ModernShell.Commands...)
	}
	model.syncShellPackages()
	return model
}
func (m *macModel) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.inspect(), m.checkMagusUpdate(), m.spinner.Tick)
}
func (m *macModel) rows() []macRow {
	if m.screen == macScreenAppConfigs {
		return m.appConfigRows()
	}
	if m.screen == macScreenRaycast {
		return raycastRows()
	}
	if m.screen == macScreenShell {
		return m.shellRows()
	}
	if m.screen == macScreenTerminal {
		return m.terminalRows()
	}
	if m.screen == macScreenMenu {
		return []macRow{
			{ID: "apps", Name: "Apps", Summary: "Find your essentials by category.", Note: "Browsers · Developer tools · AI · Productivity · Media · Communication"},
			{ID: "tools", Name: "Command-line tools", Summary: "Developer tools and utilities you run from the command line."},
			{ID: "fonts", Name: "Fonts", Summary: "Six handpicked fonts for writing, design and coding.", Note: "Inter · Source Serif 4 · Newsreader · Fraunces · Space Grotesk · JetBrains Mono"},
			{ID: "settings", Name: "Mac settings", Summary: "Six Finder preferences. Original values are saved before changes."},
			{ID: "app-configs", Name: "App setups", Summary: "Ghostty, Zed, Firefox, Modern CLI and Raycast presets."},
			{ID: "review", Name: fmt.Sprintf("Review & install (%d)", len(m.selected)), Summary: "See your complete basket before anything changes."},
			{ID: "updates", Name: "Update all", Summary: "Update eligible Homebrew apps and terminal tools.", Note: "Includes packages installed outside Magus. Review the scope before continuing."},
			{ID: "self-update", Name: "Update Magus", Summary: m.magUpdateSummary(), Note: "Replaces the current executable atomically. Restart Magus afterwards to use the new version."},
			{ID: "terminal", Name: "Terminal setup", Summary: "Ghostty themes, fonts and configurable modern commands."},
		}
	}
	if m.screen == macScreenCategories {
		return m.categoryRows()
	}
	if m.screen == macScreenPresets {
		var rows []macRow
		for i, p := range macPresets {
			rows = append(rows, macRow{ID: fmt.Sprint(i), Name: p.Name, Summary: p.Description, Note: "Presets only add selections. Your existing picks are kept."})
		}
		return rows
	}
	var rows []macRow
	for _, p := range macPackages {
		if !m.searching && m.screen != macScreenReview && ((m.category == "apps" && (p.Kind != "cask" || strings.HasPrefix(p.ID, "font-"))) || (m.category == "fonts" && !strings.HasPrefix(p.ID, "font-")) || (m.category == "tools" && p.Kind != "formula") || m.category == "settings") {
			continue
		}
		if !m.searching && m.screen == macScreenBrowse && m.category == "apps" && m.appGroup != "" {
			group, ok := appCategory(m.appGroup)
			if !ok || !oneOf(p.ID, group.Packages) {
				continue
			}
		}
		if m.screen == macScreenReview && !m.selected[p.ID] {
			continue
		}
		source := "Homebrew " + p.Kind + " / " + p.ID
		if group, ok := packageCategory(p.ID); ok {
			source = group.Name + "\n" + source
		}
		rows = append(rows, macRow{p.ID, p.Name, p.Summary, source, p.Note})
	}
	if m.searching || m.category == "settings" || m.screen == macScreenReview {
		for _, s := range macSettings {
			if m.screen == macScreenReview && !m.selected[s.ID] {
				continue
			}
			rows = append(rows, macRow{s.ID, s.Name, s.Summary, "macOS preference / " + s.Domain + " / " + s.Key, "Original value saved. Refresh Finder after applying. Behaviour needs visual verification on this macOS release."})
		}
	}
	if m.screen == macScreenReview && m.selected["shell:configure"] {
		rows = append(rows, macRow{ID: "shell:configure", Name: (shellStep{m.shellSettings}).Describe(), Summary: "Only the Magus block in .zshrc is changed. Existing settings are preserved.", Note: strings.Join(m.shellSettings.Commands, ", ")})
	}
	if m.screen == macScreenReview {
		for _, row := range append(m.terminalRows(), m.appConfigRows()...) {
			if m.selected[row.ID] {
				rows = append(rows, row)
			}
		}
	}
	if m.searching {
		rows = fuzzyMacRows(rows, m.search.Value())
	}
	if m.screen == macScreenBrowse {
		rows = m.filterCatalogueRows(rows)
	}
	if m.screen == macScreenBrowse && m.category == "settings" && !m.searching && m.catalogueFilter == macCatalogueAll {
		rows = append(rows, macRow{ID: "restore", Name: "Restore Magus settings", Summary: "Restore the original values saved by Magus. Settings changed outside Magus are left alone."})
	}
	return rows
}

func (m *macModel) catalogueFilterLabel() string {
	switch m.catalogueFilter {
	case macCatalogueAvailable:
		return "Not installed"
	case macCatalogueSelected:
		return "Selected only"
	default:
		return "All items"
	}
}

func (m *macModel) filterCatalogueRows(rows []macRow) []macRow {
	if m.catalogueFilter == macCatalogueAll {
		return rows
	}
	filtered := make([]macRow, 0, len(rows))
	for _, row := range rows {
		if m.catalogueFilter == macCatalogueSelected && m.selected[row.ID] {
			filtered = append(filtered, row)
		}
		if m.catalogueFilter == macCatalogueAvailable && m.needsSelection(row.ID) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func (m *macModel) cycleCatalogueFilter() {
	m.catalogueFilter = (m.catalogueFilter + 1) % 3
	m.cursor = 0
}
func (m *macModel) selectionManifest() Manifest {
	n := newMacManifest()
	for _, p := range appConfigPresets {
		if m.selected["config:"+p.ID] {
			n.Mac.AppConfigs = append(n.Mac.AppConfigs, p.ID)
		}
	}
	if m.selected["shell:configure"] {
		settings := m.shellSettings
		n.Mac.ModernShell = &settings
	}
	for _, theme := range terminalThemes {
		if m.selected[terminalID(theme)] {
			n.Mac.Terminal = theme
		}
	}
	for _, p := range macPackages {
		if m.selected[p.ID] {
			n.Mac.Packages = append(n.Mac.Packages, p.ID)
		}
	}
	for _, s := range macSettings {
		if m.selected[s.ID] {
			n.Mac.Settings = append(n.Mac.Settings, s.ID)
		}
	}
	return n
}
func (m *macModel) toggle(id string) {
	if m.selected["config:"+id] {
		return
	}
	if m.selected["shell:configure"] && oneOf(id, m.shellSettings.packages()) {
		return
	}

	if m.selected[id] {
		delete(m.selected, id)
		if id == "shell:configure" {
			m.syncShellPackages()
		}
	} else if m.needsSelection(id) {
		m.selected[id] = true
	}
}

func (m *macModel) needsSelection(id string) bool {
	return !oneOf(m.inventory.states[id], []string{"installed", "already set", "outside Homebrew"})
}

func (m *macModel) pruneCompletedSelections() {
	for id := range m.selected {
		if !m.needsSelection(id) {
			delete(m.selected, id)
		}
	}
}
func (m *macModel) start() tea.Cmd {
	m.manifest = m.selectionManifest()
	m.screen = macScreenInstall
	m.started = time.Now()
	m.failed = false
	m.needsFinder = false
	m.stopping = false
	m.outcomes = nil
	m.active = 0
	m.notice = ""
	m.latestLog = ""
	m.setLogs("")
	m.session = startMacSession(m.paths, m.path, m.manifest, m.restoring, m.preview, m.timeout)
	m.session.generation = m.sessionGeneration.next()
	return tea.Batch(waitMacEvent(m.session), tickActivity(), m.progress.SetPercent(0))
}
func waitMacEvent(s *macSession) tea.Cmd {
	return func() tea.Msg {
		e, ok := <-s.events
		if !ok {
			return macEvent{Kind: macEventClosed, generation: s.generation}
		}
		e.generation = s.generation
		return e
	}
}

// Keep the complete bounded log independently from the visible viewport.
func (m *macModel) logsContent() string { return m.logText }
func (m *macModel) setLogs(s string) {
	follow := m.logs.AtBottom() || m.logText == "" || !m.showLogs
	offset := m.logs.YOffset()
	m.logText = s
	m.logs.SetContent(s)
	if follow {
		m.logs.GotoBottom()
	} else {
		m.logs.SetYOffset(offset)
	}
}

func (m *macModel) activeName() string {
	if m.active >= 0 && m.active < len(m.outcomes) {
		return m.outcomes[m.active].Name
	}
	return m.notice
}

func (m *macModel) outcomeCounts() map[string]int {
	counts := map[string]int{"installed": 0, "already present": 0, "skipped": 0, "failed": 0}
	for _, outcome := range m.outcomes {
		if _, ok := counts[outcome.Status]; ok {
			counts[outcome.Status]++
		}
	}
	return counts
}

func (m *macModel) failureDiagnostic() string {
	if m.active < 0 || m.active >= len(m.outcomes) {
		return ""
	}
	o := m.outcomes[m.active]
	if o.Status != "failed" {
		return ""
	}
	return fmt.Sprintf("Failed: %s\n%s\n\nRetry with r, skip with s, or inspect full logs with l.\nDiagnostic: Magus %s | %s | %s", o.Name, o.Detail, buildVersion, m.inventory.osVersion, o.ID)
}

func (m *macModel) summaryNextAction() string {
	counts := m.outcomeCounts()
	switch {
	case counts["failed"] > 0:
		return "Next: inspect logs with l, then reopen Magus to retry failed items."
	case m.needsFinder && !m.preview:
		return "Next: press f to refresh Finder, or return to the menu."
	case m.preview:
		return "Preview complete. Return to the menu to review and apply these choices."
	default:
		return "Next: return to the menu, or run magus doctor whenever you want to check this setup."
	}
}
func runMacTUI(paths Paths, path string, m Manifest, preview bool, timeout time.Duration) error {
	model := newMacModel(paths, path, m, preview, timeout)
	defer func() {
		if model.updateReview.cancel != nil {
			model.updateReview.cancel()
		}
		if model.inventoryCancel != nil {
			model.inventoryCancel()
		}
	}()
	options := []tea.ProgramOption{}
	if terminalProfile == termenv.Ascii {
		options = append(options, tea.WithColorProfile(colorprofile.NoTTY))
	}
	p := tea.NewProgram(model, options...)
	_, err := p.Run()
	if model.session != nil {
		model.session.cancel()
		for range model.session.events {
		}
	}
	model.endBootstrap()
	return err
}
