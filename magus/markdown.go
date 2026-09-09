package main

import (
	"charm.land/glamour/v2"
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
	"strings"
	"sync"
)

var markdownCache = struct {
	sync.Mutex
	entries map[string]string
}{entries: map[string]string{}}

// Explicit styles avoid competing terminal queries. Bound the cache because
// details are rendered again on every spinner tick, but change infrequently.
func renderMarkdown(source string, width int) string {
	width = max(1, width)
	style := "dark"
	if lightBackground.Load() {
		style = "light"
	}
	if terminalProfile == termenv.Ascii {
		style = "notty"
	}
	source = ansi.Strip(source)
	key := fmt.Sprintf("%s:%d:%s", style, width, source)
	markdownCache.Lock()
	defer markdownCache.Unlock()
	if rendered, ok := markdownCache.entries[key]; ok {
		return rendered
	}
	renderer, err := glamour.NewTermRenderer(glamour.WithStandardStyle(style), glamour.WithWordWrap(width))
	rendered := source
	if err == nil {
		if result, e := renderer.Render(source); e == nil {
			rendered = strings.Trim(result, "\n")
		}
	}
	// Code blocks and long URLs may exceed the renderer's word-wrap width.
	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "…")
	}
	rendered = strings.Join(lines, "\n")
	if terminalProfile == termenv.Ascii {
		rendered = ansi.Strip(rendered)
	}
	if len(markdownCache.entries) >= 64 {
		clear(markdownCache.entries)
	}
	markdownCache.entries[key] = rendered
	return rendered
}
