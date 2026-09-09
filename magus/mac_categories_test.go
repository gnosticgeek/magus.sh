package main

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/muesli/termenv"
)

func TestEveryCaskHasExactlyOneCategory(t *testing.T) {
	seen := map[string]bool{}
	for _, group := range macAppCategories {
		if len(group.Packages) == 0 {
			t.Fatalf("empty category %s", group.ID)
		}
		for _, id := range group.Packages {
			p, ok := macPackage(id)
			if !ok || p.Kind != "cask" || seen[id] {
				t.Fatalf("invalid or repeated category member %s", id)
			}
			seen[id] = true
		}
	}
	for _, p := range macPackages {
		if p.Kind == "cask" && !strings.HasPrefix(p.ID, "font-") && !seen[p.ID] {
			t.Fatalf("uncategorised app %s", p.ID)
		}
	}
}

func TestCategoryNavigationKeepsBasketAndSearchContext(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	press(m, "enter")
	if m.screen != "categories" || len(m.rows()) != 13 {
		t.Fatal("Apps did not open categories")
	}
	press(m, "enter")
	press(m, "space")
	if !m.selected["firefox"] || len(m.rows()) != 4 {
		t.Fatal("Browsers contains wrong apps")
	}
	press(m, "esc")
	press(m, "down")
	press(m, "down")
	press(m, "enter")
	if m.appGroup != "ai" || len(m.rows()) != 8 {
		t.Fatal("AI category missing")
	}
	press(m, "space")
	press(m, "/")
	for _, r := range "productivity" {
		press(m, string(r))
	}
	if len(m.rows()) != 5 {
		t.Fatal("search does not match category names")
	}
	press(m, "esc")
	if m.screen != "browse" || m.appGroup != "ai" {
		t.Fatal("search lost its originating category")
	}
	press(m, "esc")
	if m.screen != "categories" || m.cursor != 2 {
		t.Fatal("back lost category focus")
	}
	m.screen = "review"
	if len(m.rows()) != 2 || !m.selected["firefox"] || !m.selected["ollama-app"] {
		t.Fatal("cross-category basket lost selections")
	}
	if err := m.selectionManifest().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestAuroraWordmarkColourAndPlainFallback(t *testing.T) {
	previous := terminalProfile
	defer setTerminalProfile(previous)
	setTerminalProfile(termenv.TrueColor)
	art := auroraText(macWordmark)
	colours := map[string]bool{}
	for _, code := range regexp.MustCompile(`38;2;\d+;\d+;\d+`).FindAllString(art, -1) {
		colours[code] = true
	}
	if len(colours) < 3 {
		t.Fatal("wordmark has no multicolour gradient")
	}
	if stripTerminal(art) != macWordmark {
		t.Fatal("colour styling changed the artwork")
	}
	setTerminalProfile(termenv.Ascii)
	if art := auroraText(macWordmark); art != macWordmark {
		t.Fatal("plain terminal fallback contains styling")
	}
}
