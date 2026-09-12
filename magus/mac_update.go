package main

import (
	"os/exec"
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// Update is the Mac TUI's state-transition boundary. Commands perform work and
// return typed messages; only this method publishes their results into model
// state. Generation checks reject messages from superseded asynchronous runs.
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
		if m.updateReview.generation.current(v.generation) && m.screen == macScreenUpdates {
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
		if m.noticeGeneration.current(v.generation) && m.notice == v.text {
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
		if m.screen == macScreenInstall && !m.failed {
			return m, tickActivity()
		}
	case macInventory:
		// Direct inventory messages remain useful to focused unit tests. Runtime
		// inspections use macInventoryResult so obsolete probes can be rejected.
		m.inventory = v
		m.inspecting = false
		m.pruneCompletedSelections()
	case macInventoryResult:
		if !m.inventoryGeneration.current(v.generation) {
			return m, nil
		}
		m.inventory = v.inventory
		m.inspecting = false
		m.pruneCompletedSelections()
	case macSelfUpdateChecked:
		if m.selfUpdateGeneration.current(v.generation) {
			m.magUpdateAvailable = v.available
			m.magUpdateLatest = v.latest
		}
	case macEvent:
		if !m.sessionGeneration.current(v.generation) {
			return m, nil
		}
		switch v.Kind {
		case macEventPlan:
			m.outcomes = v.Outcomes
		case macEventActive:
			m.active = v.Index
			m.failed = false
			m.notice = v.Text
		case macEventLog:
			m.latestLog = v.Text
			content := m.logsContent() + "\n" + v.Text
			if len(content) > 24000 {
				content = content[len(content)-24000:]
			}
			m.setLogs(content)
		case macEventResult:
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
		case macEventFailure:
			m.setLogs(m.logsContent() + "\n" + v.Text)
			m.failed = true
			m.notice = v.Text
		case macEventFatal:
			m.notice = v.Text
			m.screen = macScreenSummary
		case macEventFinished:
			m.outcomes = v.Outcomes
			m.cursor = 0
			m.screen = macScreenSummary
			m.notice = ""
			m.inspecting = true
		case macEventClosed:
			m.session = nil
			return m, m.inspect()
		}
		if m.session != nil {
			if v.Kind == macEventResult || v.Kind == macEventFinished {
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
			m.screen = macScreenReview
			return m, nil
		}
		m.bootstrapFile = v.path
		cmd := exec.Command("/bin/bash", v.path)
		return m, tea.ExecProcess(cmd, func(err error) tea.Msg { return macBootstrapDone{err} })
	case macBootstrapDone:
		m.endBootstrap()
		m.screen = macScreenReview
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
		m.screen, m.cursor = macScreenMenu, 6
		m.notice = "Homebrew updates completed."
		if v.err != nil {
			m.notice = "Updates stopped or failed; completed updates are kept. Open Update all to retry. " + v.err.Error()
		}
		m.inspecting = true
		return m, m.inspect()
	case macSelfUpdateDone:
		m.screen, m.cursor = macScreenMenu, 7
		if v.err != nil {
			m.notice = "Magus update failed; the existing executable is unchanged. " + v.err.Error()
		} else {
			m.notice = "Magus updated successfully. Restart it to use the new version."
		}
	case tea.PasteMsg:
		if m.searching {
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(v)
			m.cursor = 0
			return m, cmd
		}
	case tea.KeyPressMsg:
		return m.updateKey(v)
	}
	return m, nil
}
