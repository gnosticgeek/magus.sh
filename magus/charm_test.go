package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestSearchAcceptsTextAndPasteWithoutReset(t *testing.T) {
	cat, err := loadCatalogue()
	if err != nil {
		t.Fatal(err)
	}
	model := newModel(cat)
	model.step = StepPick
	next, _ := model.startSearch(PickMenu)
	model = next.(Model)
	next, _ = model.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	model = next.(Model)
	if model.searchQuery != "r" || model.pickView != PickSearch || model.step != StepPick {
		t.Fatal("search reset while typing r")
	}
	next, _ = model.Update(tea.PasteMsg{Content: "etro games"})
	model = next.(Model)
	if model.searchQuery != "retro games" {
		t.Fatalf("paste lost: %q", model.searchQuery)
	}
	next, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	model = next.(Model)
	if model.searchQuery != "retro game" {
		t.Fatal("backspace failed")
	}
}

func TestTUIViewsShowBuildVersion(t *testing.T) {
	previous := buildVersion
	buildVersion = "v0.4.0-test"
	t.Cleanup(func() { buildVersion = previous })

	cat, err := loadCatalogue()
	if err != nil {
		t.Fatal(err)
	}
	if got := ansi.Strip(newModel(cat).viewSplash()); !strings.Contains(got, "magus.sh v0.4.0-test") {
		t.Fatal("Steam TUI splash does not show the build version")
	}

	mac := newMacModel(Paths{}, "", newMacManifest(), true, 0)
	if got := ansi.Strip(mac.headerView(80)); !strings.Contains(got, "Magus v0.4.0-test") {
		t.Fatal("Mac TUI header does not show the build version")
	}
}

func TestSearchCanReachEveryResult(t *testing.T) {
	cat, err := loadCatalogue()
	if err != nil {
		t.Fatal(err)
	}
	model := newModel(cat)
	model.step = StepPick
	next, _ := model.startSearch(PickMenu)
	model = next.(Model)
	if len(model.searchResults) <= 32 {
		t.Fatal("catalogue unexpectedly small")
	}
	for i := 0; i < len(model.searchResults)-1; i++ {
		next, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		model = next.(Model)
	}
	if model.cursor != len(model.searchResults)-1 {
		t.Fatal("results are unreachable")
	}
	last := model.searchResults[model.cursor].Cmd
	if !strings.Contains(ansi.Strip(model.viewSearch()), last.Title) {
		t.Fatal("focused result is not visible")
	}
	next, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model = next.(Model)
	if !model.picked[last.ID] {
		t.Fatal("last result cannot be selected")
	}
}
func TestTerminalTextFitsVisibleWidth(t *testing.T) {
	for _, width := range []int{1, 2, 8, 24} {
		for _, text := range []string{"你好世界 long", "👩‍💻 developer", lipgloss.NewStyle().Bold(true).Render("你好 styled")} {
			if got := truncate(text, width); lipgloss.Width(got) > width {
				t.Fatalf("truncate %q exceeds %d", got, width)
			}
		}
		rendered := renderMarkdown("## Help\n\n**Changes** and [source](https://example.com/a-very-long-path).\n\n```sh\necho very-long-command-here\n```", width)
		for _, line := range strings.Split(rendered, "\n") {
			if lipgloss.Width(line) > width {
				t.Fatalf("markdown exceeds %d", width)
			}
		}
	}
}
func TestAccessibleSetupReviewAndConsent(t *testing.T) {
	for _, answer := range []string{"y", "n"} {
		t.Run(answer, func(t *testing.T) {
			var out bytes.Buffer
			m, confirmed, err := runSetupForms(Device{Kind: DeviceMachine}, true, strings.NewReader("\n\n0\n\n\n"+answer+"\n"), &out)
			if err != nil {
				t.Fatal(err)
			}
			if confirmed != (answer == "y") {
				t.Fatal("confirmation mismatch")
			}
			if err := m.Validate(); err != nil {
				t.Fatal(err)
			}
			text := out.String()
			for _, want := range []string{"Recorded but not applied", "manifest.toml", "Write this manifest"} {
				if !strings.Contains(text, want) {
					t.Fatalf("missing %q", want)
				}
			}
			if strings.Contains(text, "\x1b") {
				t.Fatal("accessible output contains terminal escapes")
			}
		})
	}
}
func TestAccessibleSetupClosedInputNeverConfirms(t *testing.T) {
	_, confirmed, err := runSetupForms(Device{Kind: DeviceDeck}, true, strings.NewReader(""), io.Discard)
	if confirmed || err == nil {
		t.Fatalf("closed input accepted: %v %v", confirmed, err)
	}
}
func TestConfirmationOverlayFitsAndRetainsText(t *testing.T) {
	for _, size := range [][2]int{{40, 10}, {80, 24}, {120, 32}} {
		message := fmt.Sprintf("Update %d reviewed apps?", 3)
		result := confirmationOverlay("Reviewed versions", message, size[0], size[1])
		if !strings.Contains(ansi.Strip(result), message) {
			t.Fatal("confirmation lost")
		}
		if lipgloss.Width(result) > size[0] || lipgloss.Height(result) > size[1] {
			t.Fatal("overlay exceeds viewport")
		}
	}
}
