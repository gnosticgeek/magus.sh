package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
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
	screen, category                                  string
	appGroup, searchReturnScreen                      string
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
	browserHeight, noticeGeneration                   int
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
	model := &macModel{paths: paths, path: path, manifest: m, preview: preview, timeout: timeout, width: 80, height: 24, screen: "menu", selected: map[string]bool{}, search: input, spinner: spin, help: help.New(), progress: progress.New(progress.WithColors(lipgloss.Color("#b5a0ff"), lipgloss.Color("#7dd3fc")), progress.WithoutPercentage()), logs: viewport.New(viewport.WithWidth(72), viewport.WithHeight(8)), inspecting: true, inventory: macInventory{states: map[string]string{}}}
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
	return tea.Batch(tea.RequestBackgroundColor, m.inspect(), m.spinner.Tick)
}
func (m *macModel) inspect() tea.Cmd {
	paths := m.paths
	return func() tea.Msg {
		inv := macInventory{states: map[string]string{}, brew: findBrew() != "", osVersion: runtime.GOOS + " / " + runtime.GOARCH}
		if runtime.GOOS != "darwin" {
			return inv
		}
		c := &Context{Paths: paths, Timeout: 30 * time.Second, Report: &Reporter{Out: io.Discard}}
		if v, err := c.macCommand("/usr/bin/sw_vers", "-productVersion"); err == nil {
			inv.osVersion = "macOS " + strings.TrimSpace(v) + " / " + runtime.GOARCH
		}
		// Snapshot each package kind once for browsing. Execution always probes anew.
		outputs := map[string]string{}
		errs := map[string]error{}
		if inv.brew {
			for _, kind := range []string{"cask", "formula"} {
				outputs[kind], errs[kind] = c.macCommand(c.brewPath(), "list", "--"+kind, "-1")
			}
		}
		c.Execute = func(ctx context.Context, name string, args ...string) (string, error) {
			if len(args) == 3 && args[0] == "list" {
				kind := strings.TrimPrefix(args[1], "--")
				return outputs[kind], errs[kind]
			}
			return runMacCommand(ctx, nil, name, args...)
		}
		for _, p := range macPackages {
			label := "needs Homebrew"
			if inv.brew {
				state, err := (brewStep{p}).Check(c)
				label = "not installed"
				if err != nil {
					label = "inspection failed"
				} else if state == StateOK {
					label = "installed"
				} else if state == StateNotApplicable {
					label = "outside Homebrew"
				}
			}
			inv.states[p.ID] = label
		}
		for _, s := range macSettings {
			state, err := (preferenceStep{s}).Check(c)
			label := "not applied"
			if err != nil {
				label = "inspection failed"
			} else if state == StateOK {
				label = "already set"
			}
			inv.states[s.ID] = label
		}
		return inv
	}
}
func (m *macModel) rows() []macRow {
	if m.screen == "app-configs" {
		return m.appConfigRows()
	}
	if m.screen == "raycast" {
		return raycastRows()
	}
	if m.screen == "shell" {
		return m.shellRows()
	}
	if m.screen == "terminal" {
		return m.terminalRows()
	}
	if m.screen == "menu" {
		return []macRow{
			{ID: "apps", Name: "Apps", Summary: "Find your essentials by category.", Note: "Browsers · Developer tools · AI · Productivity · Media · Communication"},
			{ID: "tools", Name: "Command-line tools", Summary: "Developer tools and utilities you run from the command line."},
			{ID: "fonts", Name: "Fonts", Summary: "Six handpicked fonts for writing, design and coding.", Note: "Inter · Source Serif 4 · Newsreader · Fraunces · Space Grotesk · JetBrains Mono"},
			{ID: "settings", Name: "Mac settings", Summary: "Six Finder preferences. Original values are saved before changes."},
			{ID: "app-configs", Name: "App setups", Summary: "Ghostty, Zed, Firefox, Modern CLI and Raycast presets."},
			{ID: "presets", Name: "Presets", Summary: "Thoughtful starting selections. Add a preset, then make it yours."},
			{ID: "review", Name: fmt.Sprintf("Review & install (%d)", len(m.selected)), Summary: "See your complete basket before anything changes."},
			{ID: "updates", Name: "Update all", Summary: "Update eligible Homebrew apps and terminal tools.", Note: "Includes packages installed outside Magus. Review the scope before continuing."},
			{ID: "terminal", Name: "Terminal setup", Summary: "Ghostty themes, fonts and configurable modern commands."},
		}
	}
	if m.screen == "categories" {
		return m.categoryRows()
	}
	if m.screen == "presets" {
		var rows []macRow
		for i, p := range macPresets {
			rows = append(rows, macRow{ID: fmt.Sprint(i), Name: p.Name, Summary: p.Description, Note: "Presets only add selections. Your existing picks are kept."})
		}
		return rows
	}
	var rows []macRow
	for _, p := range macPackages {
		if !m.searching && m.screen != "review" && ((m.category == "apps" && (p.Kind != "cask" || strings.HasPrefix(p.ID, "font-"))) || (m.category == "fonts" && !strings.HasPrefix(p.ID, "font-")) || (m.category == "tools" && p.Kind != "formula") || m.category == "settings") {
			continue
		}
		if !m.searching && m.screen == "browse" && m.category == "apps" && m.appGroup != "" {
			group, ok := appCategory(m.appGroup)
			if !ok || !oneOf(p.ID, group.Packages) {
				continue
			}
		}
		if m.screen == "review" && !m.selected[p.ID] {
			continue
		}
		source := "Homebrew " + p.Kind + " / " + p.ID
		if group, ok := packageCategory(p.ID); ok {
			source = group.Name + "\n" + source
		}
		rows = append(rows, macRow{p.ID, p.Name, p.Summary, source, p.Note})
	}
	if m.searching || m.category == "settings" || m.screen == "review" {
		for _, s := range macSettings {
			if m.screen == "review" && !m.selected[s.ID] {
				continue
			}
			rows = append(rows, macRow{s.ID, s.Name, s.Summary, "macOS preference / " + s.Domain + " / " + s.Key, "Original value saved. Refresh Finder after applying. Behaviour needs visual verification on this macOS release."})
		}
	}
	if m.screen == "review" && m.selected["shell:configure"] {
		rows = append(rows, macRow{ID: "shell:configure", Name: (shellStep{m.shellSettings}).Describe(), Summary: "Only the Magus block in .zshrc is changed. Existing settings are preserved.", Note: strings.Join(m.shellSettings.Commands, ", ")})
	}
	if m.screen == "review" {
		for _, row := range append(m.terminalRows(), m.appConfigRows()...) {
			if m.selected[row.ID] {
				rows = append(rows, row)
			}
		}
	}
	if m.searching {
		rows = fuzzyMacRows(rows, m.search.Value())
	}
	if m.screen == "browse" {
		rows = m.filterCatalogueRows(rows)
	}
	if m.screen == "browse" && m.category == "settings" && !m.searching && m.catalogueFilter == macCatalogueAll {
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
	m.screen = "install"
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
	return tea.Batch(waitMacEvent(m.session), tickActivity(), m.progress.SetPercent(0))
}
func waitMacEvent(s *macSession) tea.Cmd {
	return func() tea.Msg {
		e, ok := <-s.events
		if !ok {
			return macEvent{Kind: "closed"}
		}
		return e
	}
}
func (m *macModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case setupOpened:
		if v.err != nil {
			return m, m.flash(v.err.Error())
		}
		return m, nil
	case tea.BackgroundColorMsg:
		lightBackground.Store(!v.IsDark())
		return m, nil
	case macUpdatesChecked:
		if v.generation == m.updateReview.generation && m.screen == "updates" {
			m.acceptUpdateReview(v)
		}
	case macMetadataRefreshed:
		if v.err != nil {
			m.updateReview.err = v.err
			m.updateReview.loading = false
			return m, nil
		}
		return m, m.checkUpdates(false)
	case macNoticeExpired:
		if v.generation == m.noticeGeneration && m.notice == v.text {
			m.notice = ""
		}
	case progress.FrameMsg:
		model, cmd := m.progress.Update(v)
		m.progress = model
		return m, cmd
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
		m.help.SetWidth(max(20, v.Width-6))
		m.search.SetWidth(max(10, v.Width-10))
		m.logs.SetWidth(max(10, v.Width-6))
		m.logs.SetHeight(max(3, v.Height-13))
		m.progress.SetWidth(max(10, v.Width-8))
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(v)
		return m, cmd
	case activityTickMsg:
		if m.screen == "install" && !m.failed {
			return m, tickActivity()
		}
	case macInventory:
		m.inventory = v
		m.inspecting = false
		m.pruneCompletedSelections()
	case macEvent:
		switch v.Kind {
		case "plan":
			m.outcomes = v.Outcomes
		case "active":
			m.active = v.Index
			m.failed = false
			m.notice = v.Text
		case "log":
			m.latestLog = v.Text
			content := m.logsContent() + "\n" + v.Text
			if len(content) > 24000 {
				content = content[len(content)-24000:]
			}
			m.setLogs(content)
		case "result":
			if strings.HasPrefix(v.Outcome.ID, "setting:") && (v.Outcome.Status == "installed" || v.Outcome.Status == "restored") {
				m.needsFinder = true
			}
			if v.Index < len(m.outcomes) {
				m.outcomes[v.Index] = v.Outcome
				if v.Outcome.Detail != "" {
					m.setLogs(m.logsContent() + "\n" + v.Outcome.Name + ": " + v.Outcome.Detail)
				}
				if p, ok := macPackage(strings.TrimPrefix(v.Outcome.ID, "package:")); ok && p.Note != "" {
					m.setLogs(m.logsContent() + "\n" + p.Name + ": " + p.Note)
				}
			}
		case "failure":
			m.setLogs(m.logsContent() + "\n" + v.Text)
			m.failed = true
			m.notice = v.Text
		case "fatal":
			m.notice = v.Text
			m.screen = "summary"
		case "finished":
			m.outcomes = v.Outcomes
			m.cursor = 0
			m.screen = "summary"
			m.notice = ""
			m.inspecting = true
		case "closed":
			m.session = nil
			return m, m.inspect()
		}
		if m.session != nil {
			if v.Kind == "result" || v.Kind == "finished" {
				done := 0
				for _, o := range m.outcomes {
					if o.Status != "unfinished" {
						done++
					}
				}
				return m, tea.Batch(waitMacEvent(m.session), m.progress.SetPercent(float64(done)/float64(max(1, len(m.outcomes)))))
			}
			return m, waitMacEvent(m.session)
		}
	case macBootstrap:
		if v.err != nil {
			m.notice = "Homebrew setup: " + v.err.Error()
			m.endBootstrap()
			m.screen = "review"
			return m, nil
		}
		m.bootstrapFile = v.path
		cmd := exec.Command("/bin/bash", v.path)
		return m, tea.ExecProcess(cmd, func(err error) tea.Msg { return macBootstrapDone{err} })
	case macBootstrapDone:
		m.endBootstrap()
		m.screen = "review"
		m.notice = "Homebrew setup finished. Rechecking prerequisites…"
		if v.err != nil {
			m.notice = "Setup needs attention: " + v.err.Error() + ". Complete the official Homebrew instructions, then reopen Magus."
		}
		m.inspecting = true
		return m, m.inspect()
	case macFinderDone:
		if v.err != nil {
			m.notice = v.err.Error()
		} else {
			m.notice = "Finder refreshed."
		}
	case macUpdatesDone:
		m.screen, m.cursor = "menu", 6
		m.notice = "Homebrew updates completed."
		if v.err != nil {
			m.notice = "Updates stopped or failed; completed updates are kept. Open Update all to retry. " + v.err.Error()
		}
		m.inspecting = true
		return m, m.inspect()
	case tea.PasteMsg:
		if m.searching {
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(v)
			m.cursor = 0
			return m, cmd
		}
	case tea.KeyPressMsg:
		k := v.String()
		if k == "ctrl+c" {
			if m.session != nil {
				m.session.cancel()
				m.stopping = true
				m.notice = "Stopping; waiting for the active process to exit…"
				return m, nil
			}
			return m, tea.Quit
		}
		if m.screen == "bootstrap" {
			return m, nil
		}
		if m.showHelp {
			if k == "?" || k == "esc" {
				m.showHelp = false
			}
			return m, nil
		}
		if !m.searching && k == "?" {
			m.showHelp = true
			return m, nil
		}
		if m.screen == "updates" || m.screen == "update-confirm" {
			switch k {
			case "esc":
				if m.screen == "update-confirm" {
					m.screen = "updates"
					return m, nil
				}
				if m.updateReview.cancel != nil {
					m.updateReview.cancel()
				}
				m.screen, m.cursor = "menu", 6
			case "enter":
				if m.screen == "update-confirm" {
					return m, m.updateAll()
				}
				if m.updateReview.ready && len(m.updateReview.items) > 0 {
					m.screen = "update-confirm"
				}
			case "r":
				if m.screen == "updates" && !m.updateReview.loading {
					return m, m.checkUpdates(true)
				}
			case "q":
				return m, tea.Quit
			default:
				if m.screen == "updates" {
					var cmd tea.Cmd
					m.updateReview.table, cmd = m.updateReview.table.Update(v)
					return m, cmd
				}
			}
			return m, nil
		}
		if m.screen == "install" {
			switch k {
			case "l":
				m.showLogs = !m.showLogs
			case "s":
				if m.failed {
					m.session.decisions <- "skip"
					m.failed = false
					return m, tickActivity()
				}
			case "r":
				if m.failed {
					m.session.decisions <- "retry"
					m.failed = false
					return m, tickActivity()
				}
			case "q", "esc":
				m.session.cancel()
				m.stopping = true
				m.notice = "Stopping; completed installations are kept."
			}
			if m.showLogs {
				var cmd tea.Cmd
				m.logs, cmd = m.logs.Update(v)
				return m, cmd
			}
			return m, nil
		}
		if m.details {
			if k == "esc" || k == "enter" || k == "tab" {
				m.details = false
				return m, nil
			}
			var cmd tea.Cmd
			m.previewPane, cmd = m.previewPane.Update(v)
			return m, cmd
		}
		if m.searching {
			switch k {
			case "esc":
				m.searching = false
				m.search.Blur()
				m.search.SetValue("")
				m.screen = m.searchReturnScreen
				m.cursor = m.searchReturnCursor
				return m, nil
			case "up", "down", "pgup", "pgdown", "enter", "space", "tab", "ctrl+s": // navigation and selection below
			default:
				var cmd tea.Cmd
				m.search, cmd = m.search.Update(v)
				m.cursor = 0
				return m, cmd
			}
		}
		if m.screen == "summary" {
			if k == "up" {
				m.cursor = max(0, m.cursor-1)
			}
			if k == "down" {
				m.cursor = min(max(0, len(m.outcomes)-1), m.cursor+1)
			}
			if k == "q" {
				return m, tea.Quit
			}
			if k == "f" && !m.preview && m.needsFinder {
				return m, func() tea.Msg {
					_, err := runMacCommand(context.Background(), nil, "/usr/bin/killall", "Finder")
					return macFinderDone{err}
				}
			}
			if k == "enter" || k == "esc" {
				m.screen = "menu"
				m.cursor = 0
				m.restoring = false
				m.notice = ""
				if saved, err := LoadManifest(m.path); err == nil && !m.preview {
					m.manifest = saved
					m.shellAutoPackages = nil
					m.shellSettings.Enabled = false
					if saved.Mac.ModernShell != nil {
						m.shellSettings = *saved.Mac.ModernShell
						m.shellSettings.Commands = append([]string{}, saved.Mac.ModernShell.Commands...)
					}
					m.selected = map[string]bool{}
					if saved.Mac.ModernShell != nil {
						m.selected["shell:configure"] = true
					}
					for _, id := range saved.Mac.AppConfigs {
						m.selected["config:"+id] = true
					}
					if saved.Mac.Terminal != "" {
						m.selected[terminalID(saved.Mac.Terminal)] = true
					}
					for _, id := range append(saved.Mac.Packages, saved.Mac.Settings...) {
						m.selected[id] = true
					}
					m.syncShellPackages()
				}
			}
			if k == "l" {
				m.showLogs = !m.showLogs
			}
			if m.showLogs {
				var cmd tea.Cmd
				m.logs, cmd = m.logs.Update(v)
				return m, cmd
			}
			return m, nil
		}
		if m.screen == "terminal-restore" {
			if k == "esc" {
				m.screen = "terminal"
				return m, nil
			}
			if k == "enter" {
				c := &Context{Paths: m.paths, DryRun: m.preview}
				if !m.preview {
					unlock, err := acquireMacLock(m.paths)
					if err != nil {
						return m, m.flash(err.Error())
					}
					defer unlock()
				}
				err := (terminalStep{}).Remove(c)
				if err != nil {
					return m, m.flash(err.Error())
				}
				m.screen = "terminal"
				if m.preview {
					return m, m.flash("Preview: would restore Ghostty settings. No files changed.")
				}
				for _, theme := range terminalThemes {
					delete(m.selected, terminalID(theme))
				}
				if saved, err := LoadManifest(m.path); err == nil {
					saved.Mac.Terminal = ""
					if err := saved.Save(m.path); err != nil {
						return m, m.flash("Settings restored, but could not update saved selection: " + err.Error())
					}
				}
				return m, m.flash("Previous Ghostty settings restored. Reload Ghostty with Cmd+Shift+,.")
			}
			return m, nil
		}
		if m.screen == "restore" {
			if k == "esc" {
				m.screen = "browse"
				m.restoring = false
				return m, nil
			}
			if k == "enter" {
				return m, m.start()
			}
			return m, nil
		}
		switch k {
		case "q":
			if !m.searching {
				return m, tea.Quit
			}
		case "/":
			if m.screen != "review" {
				m.searchReturnScreen, m.searchReturnCursor = m.screen, m.cursor
				m.searching = true
				m.screen = "browse"
				m.cursor = 0
				return m, m.search.Focus()
			}
		case "esc":
			if m.screen == "browse" && m.category == "apps" {
				m.screen, m.cursor = "categories", m.groupCursor
			} else {
				m.screen, m.cursor = "menu", 0
			}
			m.notice = ""
		case "f":
			if m.screen == "browse" {
				m.cycleCatalogueFilter()
				return m, m.flash("Filter: " + m.catalogueFilterLabel())
			}
		case "ctrl+s":
			return m, m.selectAllResults()
		case "up", "down", "left", "right", "pgup", "pgdown", "home", "end":
			if columns, _, _ := m.browserLayout(); columns > 1 {
				m.moveGrid(k, columns)
				return m, nil
			}
			browser := m.browserList(m.rows(), max(20, m.width-6), m.browserHeight)
			browser, cmd := browser.Update(v)
			m.cursor = browser.Index()
			return m, cmd
		case "tab":
			if oneOf(m.screen, []string{"menu", "browse", "categories", "review", "presets", "terminal", "shell", "app-configs", "raycast"}) {
				m.details = true
			}
		case "b":
			if m.screen == "review" && !m.preview && !m.inventory.brew {
				return m, m.bootstrap()
			}
		case "enter", "space":
			if m.screen == "review" && k == "enter" {
				if len(m.selected) == 0 {
					m.notice = "Choose something first."
					return m, nil
				}
				if m.inspecting {
					m.notice = "Wait for the prerequisite check."
					return m, nil
				}
				if len(m.selectionManifest().Mac.Packages) > 0 && !m.inventory.brew && !m.preview {
					m.notice = "Homebrew is needed. Press b for guided setup."
					return m, nil
				}
				return m, m.start()
			}
			rows := m.rows()
			if len(rows) == 0 {
				return m, nil
			}
			m.cursor = min(m.cursor, len(rows)-1)
			row := rows[m.cursor]
			if m.screen == "menu" {
				m.cursor = 0
				if row.ID == "review" || row.ID == "presets" || row.ID == "updates" || row.ID == "terminal" || row.ID == "app-configs" {
					m.screen = row.ID
					if row.ID == "updates" {
						return m, m.checkUpdates(false)
					}
				} else {
					m.category = row.ID
					m.screen = "browse"
					if row.ID == "apps" {
						m.screen = "categories"
						m.appGroup = ""
					}
				}
			} else if m.screen == "app-configs" || m.screen == "raycast" {
				if strings.HasPrefix(row.ID, "export:") {
					if m.preview {
						return m, m.flash("Preview: would export this setup under the Magus config directory.")
					}
					path, err := exportAppConfig(m.paths, strings.TrimPrefix(row.ID, "export:"))
					if err != nil {
						return m, m.flash(err.Error())
					}
					return m, m.flash("Exported " + path)
				}
				if strings.HasPrefix(row.ID, "https://") {
					if m.preview {
						return m, m.flash("Preview: would open " + row.ID)
					}
					return m, openSetupURL(row.ID)
				}
				if strings.HasPrefix(row.ID, "config:") {
					m.toggle(row.ID)
					id := strings.TrimPrefix(row.ID, "config:")
					if m.selected[row.ID] && m.needsSelection(id) {
						m.selected[id] = true
					}
					return m, nil
				}
				if row.ID == "raycast-install" {
					if m.needsSelection("raycast") {
						m.selected["raycast"] = true
					}
					return m, m.flash("Raycast added to the review basket.")
				}
				if row.ID == "restore" {
					m.restoring = true
				}
				m.screen, m.cursor = row.ID, 0
				return m, nil
			} else if m.screen == "shell" {
				switch row.ID {
				case "shell-enable":
					m.shellSettings.Enabled = !m.shellSettings.Enabled
				case "shell-undo":
					m.shellSettings.Enabled = false
					m.screen, m.cursor = "review", 0
				case "shell-review":
					m.screen, m.cursor = "review", 0
				default:
					id := strings.TrimPrefix(row.ID, "shell-option:")
					if oneOf(id, m.shellSettings.Commands) {
						var ids []string
						for _, v := range m.shellSettings.Commands {
							if v != id {
								ids = append(ids, v)
							}
						}
						m.shellSettings.Commands = ids
					} else {
						m.shellSettings.Commands = append(m.shellSettings.Commands, id)
					}
				}
				m.selected["shell:configure"] = true
				m.syncShellPackages()
				return m, nil
			} else if m.screen == "terminal" {
				if row.ID == "shell" {
					m.screen, m.cursor = "shell", 0
					return m, nil
				}
				if row.ID == "terminal-review" {
					m.screen, m.cursor = "review", 0
					return m, nil
				}
				if row.ID == "terminal-restore" {
					m.screen = "terminal-restore"
					return m, nil
				}
				if row.ID == "terminal-export" {
					theme := m.selectionManifest().Mac.Terminal
					if theme == "" {
						theme = "Catppuccin"
					}
					path := m.paths.Config + "/exports/ghostty/config.ghostty"
					if m.preview {
						return m, m.flash("Preview: would export " + path)
					}
					if err := regularTerminalPath(path); err != nil {
						return m, m.flash(err.Error())
					}
					if _, err := os.Stat(path); err == nil {
						return m, m.flash("Export already exists: " + path + ". Move it before exporting again.")
					} else if !os.IsNotExist(err) {
						return m, m.flash(err.Error())
					}
					if err := writeFileAtomic(path, []byte(terminalConfig(theme)), 0600); err != nil {
						return m, m.flash(err.Error())
					}
					return m, m.flash("Exported " + path)
				}
				for _, theme := range terminalThemes {
					delete(m.selected, terminalID(theme))
				}
				m.selected[row.ID] = true
				for _, id := range []string{"ghostty", "font-jetbrains-mono"} {
					if m.needsSelection(id) {
						m.selected[id] = true
					}
				}
				return m, m.flash("Added " + row.Name + ". Choose Review & apply setup below to apply it.")
			} else if m.screen == "categories" {
				m.groupCursor = m.cursor
				m.appGroup, m.category, m.screen, m.cursor = row.ID, "apps", "browse", 0
			} else if m.screen == "presets" {
				for _, id := range macPresets[m.cursor].IDs {
					if m.needsSelection(id) {
						m.selected[id] = true
					}
				}
				return m, m.flash("Added " + row.Name + ". Your other picks are kept.")
			} else if row.ID == "restore" {
				m.screen = "restore"
				m.restoring = true
			} else {
				if !m.needsSelection(row.ID) {
					delete(m.selected, row.ID)
					return m, m.flash(row.Name + " is already present; nothing to install.")
				}
				if m.selected["config:"+row.ID] {
					return m, m.flash("Required by the selected app preset. Remove the preset first.")
				}
				if m.selected["shell:configure"] && oneOf(row.ID, m.shellSettings.packages()) {
					return m, m.flash("Required by your modern commands. Turn off its switch in Terminal setup first.")
				}
				m.toggle(row.ID)
				action := "Removed "
				if m.selected[row.ID] {
					action = "Selected "
				}
				return m, m.flash(action + row.Name)
			}
		}
	}
	return m, nil
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
func (m *macModel) endBootstrap() {
	if m.bootstrapCancel != nil {
		m.bootstrapCancel()
		m.bootstrapCancel = nil
	}
	if m.bootstrapFile != "" {
		os.Remove(m.bootstrapFile)
		m.bootstrapFile = ""
	}
	if m.bootstrapUnlock != nil {
		m.bootstrapUnlock()
		m.bootstrapUnlock = nil
	}
}
func (m *macModel) bootstrap() tea.Cmd {
	unlock, err := acquireMacLock(m.paths)
	if err != nil {
		m.notice = err.Error()
		return nil
	}
	m.bootstrapUnlock = unlock
	parent, cancel := context.WithCancel(context.Background())
	m.bootstrapCancel = cancel
	m.screen = "bootstrap"
	return func() tea.Msg {
		f, err := os.CreateTemp("", "magus-homebrew-*.sh")
		if err != nil {
			return macBootstrap{err: err}
		}
		path := f.Name()
		f.Close()
		ctx, cancel := context.WithTimeout(parent, 2*time.Minute)
		defer cancel()
		_, err = runMacCommand(ctx, nil, "/usr/bin/curl", "--proto", "=https", "--tlsv1.2", "-fsSL", "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh", "-o", path)
		if err != nil {
			os.Remove(path)
			return macBootstrap{err: err}
		}
		return macBootstrap{path: path}
	}
}
func (m *macModel) View() tea.View {
	v := tea.NewView(m.viewContent())
	v.AltScreen = true
	v.WindowTitle = "Magus"
	return v
}

func (m *macModel) viewContent() string {
	w := max(16, m.width-6)
	header := m.headerView(w)
	h := max(4, m.height-lipgloss.Height(header)-6)
	var body string
	keys := hints("↑↓", "move", "enter", "open", "/", "search", "?", "help", "q", "quit")
	if m.showHelp {
		body = "Help / " + m.screen + "\n\n"
		for _, binding := range m.contextualKeys().bindings {
			body += binding.Help().Key + "  " + binding.Help().Desc + "\n"
		}
		keys = hints("? / esc", "close help")
	} else {
		switch m.screen {
		case "updates":
			body = m.updateReviewView(w, h)
			keys = hints("↑↓", "rows", "enter", "review", "r", "refresh", "?", "help", "esc", "back")
		case "update-confirm":
			body = fmt.Sprintf("Update %d reviewed apps and tools?\n\nClose the affected apps before continuing.\nHomebrew may also update required dependencies.\nIts live output and password prompts take over the terminal.\nCtrl+C stops; completed updates remain.\n\nOnly the reviewed package list is requested.\nNo automatic cleanup is requested.", len(m.updateReview.items))
			body = confirmationOverlay(m.updateReviewView(w, h), body, w, h)
			keys = hints("enter", "update reviewed", "esc", "back to versions")
		case "bootstrap":
			body = "Preparing the official Homebrew installer…\n\nThe terminal will be handed to Homebrew for its prompts.\nMagus will resume when it finishes."
		case "terminal-restore":
			body = "Restore previous Ghostty settings?\n\nThe saved dotfile will be restored. Ghostty and fonts stay installed.\nManually edited configurations are left unchanged."
			keys = hints("enter", "restore", "esc", "back")
		case "restore":
			body = "Restore Magus settings\n\nRestore saved preferences, Zed and Firefox files, and remove the modern shell block.\nFirefox values already loaded need separate resets in about:config.\nEdited settings are left alone. Open a new terminal afterwards.\nApplications and packages will remain installed.\n\nEnter restores / Escape returns."
			keys = hints("enter", "restore", "esc", "back")
		case "install", "summary":
			done := 0
			for _, o := range m.outcomes {
				if o.Status != "unfinished" {
					done++
				}
			}
			heading := "Setup summary"
			if m.screen == "install" {
				heading = m.spinner.View() + " Installing"
				if m.failed {
					heading = "An item needs attention"
				}
			}
			if m.screen == "install" {
				item := min(m.active+1, len(m.outcomes))
				body = heading + fmt.Sprintf("\nItem %d of %d · %d completed · elapsed %s\n", item, len(m.outcomes), done, time.Since(m.started).Round(time.Second))
				if !m.failed {
					body += m.progress.View() + "\n"
				}
				if current := m.activeName(); current != "" {
					body += sBright.Render("Now: ") + ansi.Truncate(current, max(8, w-5), "…") + "\n"
				}
				if m.latestLog != "" && !m.showLogs {
					body += sMuted.Render(ansi.Truncate(m.latestLog, w, "…")) + "\n"
				}
			} else {
				counts := m.outcomeCounts()
				body = heading + fmt.Sprintf("\n%d / %d items finished · elapsed %s\n", done, len(m.outcomes), time.Since(m.started).Round(time.Second))
				body += fmt.Sprintf("%d installed · %d already present · %d skipped · %d failed\n", counts["installed"], counts["already present"], counts["skipped"], counts["failed"])
				body += sMuted.Render(m.summaryNextAction()) + "\n"
				if len(m.outcomes) > 0 {
					body += m.progress.ViewAs(float64(done)/float64(len(m.outcomes))) + "\n"
				}
			}
			if m.screen == "install" && m.failed {
				body += "\n" + m.failureDiagnostic() + "\n"
			}
			if m.showLogs {
				body += "\n" + m.logs.View()
			} else {
				available := max(2, h-5)
				start := max(0, m.active-available+1)
				if m.screen == "summary" {
					start = max(0, m.cursor-available+1)
				}
				for i := start; i < len(m.outcomes) && i < start+available; i++ {
					o := m.outcomes[i]
					line := fmt.Sprintf("%-16s %s", o.Status, o.Name)
					if m.screen == "summary" && i == m.cursor {
						line = "> " + line
					}
					body += ansi.Truncate(line, w, "…") + "\n"
				}
			}
			keys = hints("l", "logs", "q", "stop")
			if m.failed {
				keys = hints("r", "retry", "s", "skip", "q", "stop", "l", "logs")
			}
			if m.screen == "summary" {
				keys = hints("↑↓", "scroll", "enter", "menu", "l", "logs", "q", "quit")
				if !m.preview && m.needsFinder {
					keys = hints("enter", "menu", "f", "refresh Finder", "l", "logs", "q", "quit")
				}
			}
		default:
			rows := m.rows()
			if m.cursor >= len(rows) {
				m.cursor = max(0, len(rows)-1)
			}
			prefix := ""
			if m.searching {
				prefix = m.search.View() + "\n\n"
				keys = hints("↑↓", "move", "enter/space", "select", "ctrl+s", "all/none", "esc", "close search")
			} else if m.screen == "menu" {
				if m.height >= 26 {
					prefix = macAccent.Render("Make yourself at home") + "\n"
				}
			} else if m.screen == "categories" {
				prefix = macAccent.Render("Apps") + sMuted.Render("  /  find your essentials") + "\n\n"
				keys = hints("enter", "open", "tab", "details", "esc", "back", "/", "search")
			} else if m.screen == "browse" {
				label := map[string]string{"apps": "Apps", "tools": "Command-line tools", "fonts": "Fonts", "settings": "Mac settings"}[m.category]
				prefix = macAccent.Render(label)
				if group, ok := appCategory(m.appGroup); ok && m.category == "apps" {
					prefix += sDim.Render("  /  ") + group.style().Render(group.Name)
				}
				prefix += sDim.Render("  /  "+m.catalogueFilterLabel()) + "\n\n"
				keys = hints("enter/space", "select", "f", "filter", "/", "search", "?", "help")
			}
			if m.screen == "review" {
				prefix = "Review your selection\n"
				if m.selectionManifest().Mac.Terminal != "" {
					prefix += "Ghostty dotfile will be replaced; original backed up.\n"
				}
				if len(m.selectionManifest().Mac.Settings) > 0 {
					prefix += "Settings save original values. Finder refresh is optional.\n"
				}
				if !m.inventory.brew && len(m.selectionManifest().Mac.Packages) > 0 {
					prefix += "Homebrew required · b opens official guided setup\n"
				}
				prefix += "\n"
				keys = hints("enter", "install selected", "space", "remove", "esc", "back")
			}
			if m.screen == "terminal" {
				prefix = macAccent.Render("Terminal setup") + "\n"
				keys = hints("enter", "choose", "tab", "details", "esc", "menu")
			}
			if m.screen == "app-configs" {
				prefix = "App setups · choose a preset or export\n"
			}
			if m.screen == "raycast" {
				prefix = "Raycast · guided setup\n"
			}
			if m.screen == "shell" {
				prefix = "Modern commands · Zsh\nChoose integrations, then review & apply.\n"
			}
			if m.screen == "presets" {
				prefix = "Presets\n\n"
				keys = hints("enter", "add preset", "esc", "back")
			}
			available := max(1, h-lipgloss.Height(prefix))
			columns, leftW, split := m.browserLayout()
			m.browserHeight = max(1, available-1)
			browser := m.browserList(rows, leftW, m.browserHeight)
			left := browser.View()
			if columns > 1 {
				left = m.gridView(rows, leftW, m.browserHeight, columns)
			}
			if len(rows) == 0 {
				left = "No matching items."
			}
			if m.screen != "menu" && columns == 1 {
				left += "\n" + sMuted.Render(fmt.Sprintf("%d results · page %d/%d", len(rows), browser.Paginator.Page+1, max(1, browser.Paginator.TotalPages)))
			}
			right := ""
			if len(rows) > 0 {
				r := rows[m.cursor]
				right = "## " + r.Name + "\n\n" + r.Summary
				if r.Source != "" {
					right += "\n\n**Source:** " + r.Source
				}
				if st := m.inventory.states[r.ID]; st != "" {
					right += "\n\n**Status:** " + st
				}
				if r.Note != "" {
					right += "\n\n" + r.Note
				}
			}
			if m.details && !split {
				body = m.detailsView(right, w, h)
				keys = hints("↑↓", "scroll", "tab/esc", "list")
			} else if split {
				right = m.detailsView(right, w-leftW-3, available)
				if m.details {
					keys = hints("↑↓", "scroll details", "pgup/pgdown", "page", "tab/esc", "list")
				}
				body = prefix + lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(leftW).Render(left), sDim.Render(" │ "), right)
			} else {
				body = prefix + left
			}
		}
	}
	// Clamp to actual terminal dimensions, reserving the footer at every size.
	lines := strings.Split(body, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, w, "")
	}
	body = strings.Join(lines, "\n")
	body = lipgloss.NewStyle().Height(h).Render(body)
	footer := fmt.Sprintf("%d selected", len(m.selected))
	if m.inspecting {
		footer += " · inspecting this Mac…"
	}
	if m.notice != "" {
		footer = cleanLog(m.notice)
	}
	footer = ansi.Truncate(strings.ReplaceAll(footer, "\n", " "), w, "…")
	return lipgloss.NewStyle().Padding(1, 3).Render(header + "\n\n" + body + "\n" + sMuted.Render(footer) + "\n" + m.help.View(keys))
}
func runMacTUI(paths Paths, path string, m Manifest, preview bool, timeout time.Duration) error {
	model := newMacModel(paths, path, m, preview, timeout)
	defer func() {
		if model.updateReview.cancel != nil {
			model.updateReview.cancel()
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
