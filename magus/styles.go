package main

import "github.com/charmbracelet/lipgloss"

// Palette — mirrors the t-* CSS variables used by the Astro prototype.
// Hex values picked to land on a dark terminal background.
var (
	colorBright = lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#f4f1ea"} // focused / primary text
	colorText   = lipgloss.AdaptiveColor{Light: "#2a2a2a", Dark: "#d6cfc1"} // normal / picked
	colorMuted  = lipgloss.AdaptiveColor{Light: "#555555", Dark: "#857f72"} // unpicked / summary
	colorDim    = lipgloss.AdaptiveColor{Light: "#888888", Dark: "#5a5448"} // separators / hints
	colorAccent = lipgloss.AdaptiveColor{Light: "#b45309", Dark: "#f59e0b"} // cursor / checkbox / focus
	colorWarn   = lipgloss.AdaptiveColor{Light: "#9f1239", Dark: "#fb7185"} // errors / empty state
)

// Reusable styles.
var (
	sBright = lipgloss.NewStyle().Foreground(colorBright).Bold(true)
	sText   = lipgloss.NewStyle().Foreground(colorText)
	sMuted  = lipgloss.NewStyle().Foreground(colorMuted)
	sDim    = lipgloss.NewStyle().Foreground(colorDim)
	sAccent = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	sWarn   = lipgloss.NewStyle().Foreground(colorWarn)

	sCursor = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	sCheck  = lipgloss.NewStyle().Foreground(colorAccent)

	sHeaderRule = lipgloss.NewStyle().Foreground(colorDim)
)

// pad pads s to n runes (visible width). Trims if longer.
func pad(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + repeat(" ", n-w)
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}

// rule renders a dim horizontal divider of width n.
func rule(n int) string {
	return sDim.Render(repeat("─", n))
}

// HintKind ranks status-bar keys by importance.
type HintKind int

const (
	HintNormal HintKind = iota
	HintPrimary
	HintSystem
)

// Hint is one key-action pair shown in the status bar.
type Hint struct {
	Key    string
	Action string
	Kind   HintKind
}

// statusBar renders hints with key-kind styling.
func statusBar(hints []Hint) string {
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		var kbd, action string
		switch h.Kind {
		case HintPrimary:
			kbd = lipgloss.NewStyle().
				Foreground(colorBright).
				Background(colorAccent).
				Padding(0, 1).
				Render(h.Key)
			action = sText.Render(h.Action)
		case HintSystem:
			kbd = lipgloss.NewStyle().
				Foreground(colorDim).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorDim).
				Padding(0, 0).
				Render(" " + h.Key + " ")
			action = sDim.Render(h.Action)
		default:
			kbd = lipgloss.NewStyle().
				Foreground(colorText).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorDim).
				Padding(0, 0).
				Render(" " + h.Key + " ")
			action = sText.Render(h.Action)
		}
		parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Center, kbd, " ", action))
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, joinSpaced(parts, "   ")...)
}

func joinSpaced(parts []string, sep string) []string {
	if len(parts) == 0 {
		return parts
	}
	out := make([]string, 0, 2*len(parts)-1)
	for i, p := range parts {
		if i > 0 {
			out = append(out, sMuted.Render(sep))
		}
		out = append(out, p)
	}
	return out
}
