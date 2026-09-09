package main

import (
	"charm.land/lipgloss/v2"
	"github.com/muesli/termenv"
	"sync/atomic"
)

// Bubble Tea owns terminal queries; styles only resolve the last reported theme.
var lightBackground atomic.Bool
var terminalProfile = termenv.TrueColor

func setTerminalProfile(p termenv.Profile) { terminalProfile = p }

type adaptiveColor struct{ Light, Dark string }

func (c adaptiveColor) RGBA() (uint32, uint32, uint32, uint32) {
	if lightBackground.Load() {
		return lipgloss.Color(c.Light).RGBA()
	}
	return lipgloss.Color(c.Dark).RGBA()
}
