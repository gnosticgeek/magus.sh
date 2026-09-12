package main

import (
	"context"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// updateKey contains navigation and intent handling only. It delegates
// machine work to commands and leaves asynchronous publication to Update.
func (m *macModel) updateKey(v tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := v.String()
	if k == "backspace" && !m.searching {
		k = "esc"
	}
	if k == "ctrl+c" {
		if m.session != nil {
			m.session.cancel()
			m.stopping = true
			m.notice = "Stopping; waiting for the active process to exit…"
			return m, nil
		}
		return m, tea.Quit
	}
	if m.screen == macScreenBootstrap {
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
	if m.screen == macScreenUpdates || m.screen == macScreenUpdateConfirm {
		switch k {
		case "esc":
			if m.screen == macScreenUpdateConfirm {
				m.screen = macScreenUpdates
				return m, nil
			}
			if m.updateReview.cancel != nil {
				m.updateReview.cancel()
			}
			m.screen, m.cursor = macScreenMenu, 6
		case "enter":
			if m.screen == macScreenUpdateConfirm {
				return m, m.updateAll()
			}
			if m.updateReview.ready && len(m.updateReview.items) > 0 {
				m.screen = macScreenUpdateConfirm
			}
		case "r":
			if m.screen == macScreenUpdates && !m.updateReview.loading {
				return m, m.checkUpdates(true)
			}
		case "q":
			return m, tea.Quit
		default:
			if m.screen == macScreenUpdates {
				var cmd tea.Cmd
				m.updateReview.table, cmd = m.updateReview.table.Update(v)
				return m, cmd
			}
		}
		return m, nil
	}
	if m.screen == macScreenSelfUpdate {
		switch k {
		case "esc":
			m.screen, m.cursor = macScreenMenu, 7
		case "enter":
			return m, m.updateMagus()
		case "q":
			return m, tea.Quit
		}
		return m, nil
	}
	if m.screen == macScreenInstall {
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
	if m.screen == macScreenSummary {
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
			m.screen = macScreenMenu
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
	if m.screen == macScreenTerminalRestore {
		if k == "esc" {
			m.screen = macScreenTerminal
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
			m.screen = macScreenTerminal
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
	if m.screen == macScreenRestore {
		if k == "esc" {
			m.screen = macScreenBrowse
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
		if m.screen.searchable() {
			m.searchReturnScreen, m.searchReturnCursor = m.screen, m.cursor
			m.searching = true
			m.screen = macScreenBrowse
			m.cursor = 0
			return m, m.search.Focus()
		}
	case "esc":
		if m.screen == macScreenBrowse && m.category == "apps" {
			m.screen, m.cursor = macScreenCategories, m.groupCursor
		} else {
			m.screen, m.cursor = macScreenMenu, 0
		}
		m.notice = ""
	case "f":
		if m.screen == macScreenBrowse {
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
		if m.screen.supportsDetails() {
			m.details = true
		}
	case "b":
		if m.screen == macScreenReview && !m.preview && !m.inventory.brew {
			return m, m.bootstrap()
		}
	case "enter", "space":
		if m.screen == macScreenReview && k == "enter" {
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
		if m.screen == macScreenMenu {
			m.cursor = 0
			if row.ID == "review" || row.ID == "updates" || row.ID == "self-update" || row.ID == "terminal" || row.ID == "app-configs" {
				m.screen = macScreen(row.ID)
				if row.ID == "updates" {
					return m, m.checkUpdates(false)
				}
			} else {
				m.category = row.ID
				m.screen = macScreenBrowse
				if row.ID == "apps" {
					m.screen = macScreenCategories
					m.appGroup = ""
				}
			}
		} else if m.screen == macScreenAppConfigs || m.screen == macScreenRaycast {
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
			m.screen, m.cursor = macScreen(row.ID), 0
			return m, nil
		} else if m.screen == macScreenShell {
			switch row.ID {
			case "shell-enable":
				m.shellSettings.Enabled = !m.shellSettings.Enabled
			case "shell-undo":
				m.shellSettings.Enabled = false
				m.screen, m.cursor = macScreenReview, 0
			case "shell-review":
				m.screen, m.cursor = macScreenReview, 0
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
		} else if m.screen == macScreenTerminal {
			if row.ID == "shell" {
				m.screen, m.cursor = macScreenShell, 0
				return m, nil
			}
			if row.ID == "terminal-review" {
				m.screen, m.cursor = macScreenReview, 0
				return m, nil
			}
			if row.ID == "terminal-restore" {
				m.screen = macScreenTerminalRestore
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
		} else if m.screen == macScreenCategories {
			m.groupCursor = m.cursor
			m.appGroup, m.category, m.screen, m.cursor = row.ID, "apps", macScreenBrowse, 0
		} else if m.screen == macScreenPresets {
			for _, id := range macPresets[m.cursor].IDs {
				if m.needsSelection(id) {
					m.selected[id] = true
				}
			}
			return m, m.flash("Added " + row.Name + ". Your other picks are kept.")
		} else if row.ID == "restore" {
			m.screen = macScreenRestore
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

	return m, nil
}
