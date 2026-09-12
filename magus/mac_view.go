package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// View and viewContent are deliberately isolated from command construction and
// event handling. Rendering may normalize cursors and viewport dimensions, but
// it never starts I/O or machine mutations.
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
		body = "Help / " + string(m.screen) + "\n\n"
		for _, binding := range m.contextualKeys().bindings {
			body += binding.Help().Key + "  " + binding.Help().Desc + "\n"
		}
		keys = hints("? / esc", "close help")
	} else {
		switch m.screen {
		case macScreenUpdates:
			body = m.updateReviewView(w, h)
			keys = hints("↑↓", "rows", "enter", "review", "r", "refresh", "?", "help", "esc", "back")
		case macScreenUpdateConfirm:
			body = macWarning.Render("Attention: close affected apps before continuing.") + fmt.Sprintf("\n\nUpdate %d reviewed apps and tools?\n\nHomebrew may also update required dependencies.\nIts live output and password prompts take over the terminal.\nCtrl+C stops; completed updates remain.\n\nOnly the reviewed package list is requested.\nNo automatic cleanup is requested.", len(m.updateReview.items))
			body = confirmationOverlay(m.updateReviewView(w, h), body, w, h)
			keys = hints("enter", "update reviewed", "esc", "back to versions")
		case macScreenSelfUpdate:
			body = "Update to the latest Magus release?\n\nMagus will quit when the update completes."
			keys = hints("enter", "yes, update", "esc", "back")
		case macScreenBootstrap:
			body = "Preparing the official Homebrew installer…\n\nThe terminal will be handed to Homebrew for its prompts.\nMagus will resume when it finishes."
		case macScreenTerminalRestore:
			body = "Restore previous Ghostty settings?\n\nThe saved dotfile will be restored. Ghostty and fonts stay installed.\nManually edited configurations are left unchanged."
			keys = hints("enter", "restore", "esc", "back")
		case macScreenRestore:
			body = "Restore Magus settings\n\nRestore saved preferences, Zed and Firefox files, and remove the modern shell block.\nFirefox values already loaded need separate resets in about:config.\nEdited settings are left alone. Open a new terminal afterwards.\nApplications and packages will remain installed.\n\nEnter restores / Escape returns."
			keys = hints("enter", "restore", "esc", "back")
		case macScreenInstall, macScreenSummary:
			done := 0
			for _, o := range m.outcomes {
				if o.Status != "unfinished" {
					done++
				}
			}
			heading := "Setup summary"
			if m.screen == macScreenInstall {
				heading = m.spinner.View() + " Installing"
				if m.failed {
					heading = "An item needs attention"
				}
			}
			if m.screen == macScreenInstall {
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
				body += lipgloss.JoinHorizontal(lipgloss.Left,
					macStatusStyle(macStatusSuccess).Render(fmt.Sprintf("%d Installed", counts["installed"])),
					sMuted.Render(" · "),
					macStatusStyle(macStatusInfo).Render(fmt.Sprintf("%d Already present", counts["already present"])),
					sMuted.Render(" · "),
					macStatusStyle(macStatusWarning).Render(fmt.Sprintf("%d Skipped", counts["skipped"])),
					sMuted.Render(" · "),
					macStatusStyle(macStatusDanger).Render(fmt.Sprintf("%d Failed", counts["failed"])),
				) + "\n"
				body += sMuted.Render(m.summaryNextAction()) + "\n"
				if len(m.outcomes) > 0 {
					body += m.progress.ViewAs(float64(done)/float64(len(m.outcomes))) + "\n"
				}
			}
			if m.screen == macScreenInstall && m.failed {
				body += "\n" + macDanger.Render("Action needed") + "\n" + m.failureDiagnostic() + "\n"
			}
			if m.showLogs {
				body += "\n" + m.logs.View()
			} else {
				available := max(2, h-5)
				start := max(0, m.active-available+1)
				if m.screen == macScreenSummary {
					start = max(0, m.cursor-available+1)
				}
				for i := start; i < len(m.outcomes) && i < start+available; i++ {
					o := m.outcomes[i]
					label, status := macOutcomeStatus(o.Status)
					line := macStatusStyle(status).Render(fmt.Sprintf("%-16s", label)) + " " + o.Name
					if m.screen == macScreenSummary && i == m.cursor {
						line = "> " + line
					}
					body += ansi.Truncate(line, w, "…") + "\n"
				}
			}
			keys = hints("l", "logs", "q", "stop")
			if m.failed {
				keys = hints("r", "retry", "s", "skip", "q", "stop", "l", "logs")
			}
			if m.screen == macScreenSummary {
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
			} else if m.screen == macScreenMenu {
				if m.height >= 26 {
					prefix = macAccent.Render("Make yourself at home") + "\n"
				}
			} else if m.screen == macScreenDeveloper {
				prefix = macAccent.Render("Developer & Terminal") + sMuted.Render("  /  tools, setup and customisation") + "\n\n"
				keys = hints("enter", "open", "tab", "details", "esc", "back", "/", "search")
			} else if m.screen == macScreenCategories {
				prefix = macAccent.Render("Apps") + sMuted.Render("  /  find your essentials") + "\n\n"
				keys = hints("enter", "open", "tab", "details", "esc", "back", "/", "search")
			} else if m.screen == macScreenBrowse {
				label := map[string]string{"apps": "Apps", "tools": "Terminal tools", "fonts": "Fonts", "settings": "Mac settings"}[m.category]
				prefix = macAccent.Render(label)
				if group, ok := appCategory(m.appGroup); ok && m.category == "apps" {
					prefix += sDim.Render("  /  ") + group.style().Render(group.Name)
				}
				if m.category == "tools" && m.appGroup == macDeveloperToolsGroup {
					prefix += sDim.Render("  /  Developer environments")
				}
				prefix += sDim.Render("  /  "+m.catalogueFilterLabel()) + "\n\n"
				if m.category == "settings" {
					prefix += sMuted.Render("For broader Mac utilities: Apps > Menu Bar > Vorssaint") + "\n\n"
				}
				keys = hints("enter/space", "select", "f", "filter", "/", "search", "?", "help")
			}
			if m.screen == macScreenReview {
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
			if m.screen == macScreenTerminal {
				prefix = macAccent.Render("Terminal setup") + "\n"
				keys = hints("enter", "choose", "tab", "details", "esc", "menu")
			}
			if m.screen == macScreenAppConfigs {
				prefix = "App setups · choose a preset or export\n"
			}
			if m.screen == macScreenRaycast {
				prefix = "Raycast · guided setup\n"
			}
			if m.screen == macScreenShell {
				prefix = "Modern commands · Zsh\nChoose integrations, then review & apply.\n"
			}
			if m.screen == macScreenPresets {
				prefix = "Presets\n\n"
				keys = hints("enter", "add preset", "esc", "back")
			}
			columns, leftW, split := m.browserLayout()
			if !m.searching && !m.details && !split && len(rows) > 0 && (m.screen == macScreenMenu || m.screen == macScreenDeveloper || m.screen == macScreenBrowse) {
				focused := rows[m.cursor]
				prefix += sMuted.Render(ansi.Truncate(focused.Summary, w, "…")) + "\n\n"
			}
			available := max(1, h-lipgloss.Height(prefix))
			m.browserHeight = max(1, available-1)
			browser := m.browserList(rows, leftW, m.browserHeight)
			left := browser.View()
			if columns > 1 {
				left = m.gridView(rows, leftW, m.browserHeight, columns)
			}
			if len(rows) == 0 {
				left = "No matching items."
			}
			if m.screen != macScreenMenu && columns == 1 {
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
