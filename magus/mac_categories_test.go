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

func TestMacPackageSummariesAreConcise(t *testing.T) {
	for _, p := range macPackages {
		summary := strings.TrimSpace(p.Summary)
		if summary == "" {
			t.Fatalf("%s has no catalogue description", p.ID)
		}
		if strings.Contains(summary, "\n") || len([]rune(summary)) > 180 {
			t.Fatalf("%s has an overlong catalogue description: %q", p.ID, summary)
		}
		if strings.Count(summary, ".")+strings.Count(summary, "!")+strings.Count(summary, "?") > 2 {
			t.Fatalf("%s has more than two sentences: %q", p.ID, summary)
		}
	}
}

func TestAppleContainerEligibilityIsVisibleBeforeSelection(t *testing.T) {
	m := newMacModel(macTestContext(t).Paths, "", newMacManifest(), true, time.Second)
	m.inventory = macInventory{states: map[string]string{}, appleSilicon: false, osMajor: 15, xcodeMajor: 16}
	if reason := m.selectionBlockReason("container"); !strings.Contains(reason, "Apple Silicon") || !strings.Contains(reason, "macOS 26+") || !strings.Contains(reason, "Xcode 26+") {
		t.Fatalf("unexpected compatibility reason: %q", reason)
	}
	m.inventory.appleSilicon, m.inventory.osMajor, m.inventory.xcodeMajor = true, 26, 26
	if reason := m.selectionBlockReason("container"); reason != "" {
		t.Fatalf("supported Apple Container was blocked: %q", reason)
	}
}

func TestMacStatusVocabularyIsConsistent(t *testing.T) {
	for _, want := range []struct {
		state string
		label string
	}{
		{"installed", "Installed"},
		{"already set", "Configured"},
		{"outside Homebrew", "External"},
		{"needs Homebrew", "Needs Homebrew"},
		{"inspection failed", "Check failed"},
	} {
		label, _, ok := macInventoryStatus(want.state)
		if !ok || label != want.label {
			t.Fatalf("%q rendered as %q, %t; want %q", want.state, label, ok, want.label)
		}
	}
	for _, want := range []struct {
		state string
		label string
	}{
		{"installed", "Installed"},
		{"already present", "Already present"},
		{"skipped", "Skipped"},
		{"failed", "Failed"},
		{"unfinished", "Pending"},
	} {
		label, _ := macOutcomeStatus(want.state)
		if label != want.label {
			t.Fatalf("outcome %q rendered as %q; want %q", want.state, label, want.label)
		}
	}
}

func TestExpandedCategoriesStayCurated(t *testing.T) {
	wantMinimum := map[string]int{"games": 5, "security": 6}
	for id, minimum := range wantMinimum {
		group, ok := appCategory(id)
		if !ok {
			t.Fatalf("missing category %s", id)
		}
		if len(group.Packages) < minimum {
			t.Fatalf("category %s has %d apps; want at least %d", id, len(group.Packages), minimum)
		}
	}
}

func TestOpenSourceCaskImportIsCategorised(t *testing.T) {
	for _, id := range []string{
		"blockblock", "coteditor", "drawio", "hammerspoon", "joplin", "libreoffice", "localsend",
		"maccy", "monitorcontrol", "oversight", "taskexplorer", "utm", "vscodium", "whatsyoursign",
	} {
		p, ok := macPackage(id)
		if !ok || p.Kind != "cask" {
			t.Fatalf("open-source cask import is missing %s", id)
		}
		if _, ok := packageCategory(id); !ok {
			t.Fatalf("open-source cask import is uncategorised: %s", id)
		}
	}
}

func TestReviewedMediaAndDeveloperToolsAreAvailable(t *testing.T) {
	for _, id := range []string{"yt-dlp", "ocrmypdf", "ffmpeg", "tesseract", "imagemagick", "neovim"} {
		p, ok := macPackage(id)
		if !ok || p.Kind != "formula" {
			t.Fatalf("reviewed formula is missing %s", id)
		}
	}
	p, ok := macPackage("vimr")
	if !ok || p.Kind != "cask" {
		t.Fatal("reviewed VimR cask is missing")
	}
	group, ok := packageCategory("vimr")
	if !ok || group.ID != "development" {
		t.Fatal("VimR is not in Developer tools")
	}
	if _, ok := macPackage("topgrade"); ok {
		t.Fatal("Topgrade was included despite being excluded from the review")
	}
}

func TestThawReplacesIce(t *testing.T) {
	if _, ok := macPackage("jordanbaird-ice"); ok {
		t.Fatal("Ice remains in the catalogue after replacement")
	}
	p, ok := macPackage("thaw")
	if !ok || p.Kind != "cask" || p.AppBundle != "Thaw.app" {
		t.Fatal("Thaw replacement is missing or invalid")
	}
	group, ok := packageCategory("thaw")
	if !ok || group.ID != "menubar" {
		t.Fatal("Thaw is not in the Menu Bar category")
	}
}

func TestLanguagesMenuContainsCoreRuntimes(t *testing.T) {
	for _, id := range []string{"node", "python@3.14", "uv", "go", "rust"} {
		p, ok := macPackage(id)
		if !ok || p.Kind != "formula" || !oneOf(id, macLanguageToolIDs) {
			t.Fatalf("languages menu is missing %s", id)
		}
	}
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen = macScreenDeveloper
	found := false
	for i, row := range m.rows() {
		if row.ID == macLanguagesGroup {
			m.cursor = i
			found = true
			break
		}
	}
	if !found {
		t.Fatal("developer menu did not expose languages")
	}
	press(m, "enter")
	if m.appGroup != macLanguagesGroup || len(m.rows()) != len(macLanguageToolIDs) {
		t.Fatal("languages menu did not show its packages")
	}
}

func TestAgentsMenuSeparatesSkillsFromAppSetups(t *testing.T) {
	m := newMacModel(macTestContext(t).Paths, "", newMacManifest(), true, time.Second)
	for i, row := range m.rows() {
		if row.ID == "agents" {
			m.cursor = i
			break
		}
	}
	press(m, "enter")
	if m.screen != macScreenAgents {
		t.Fatal("AI & agents menu did not open")
	}
	for i, row := range m.rows() {
		if row.ID == "skills" {
			m.cursor = i
			break
		}
	}
	press(m, "enter")
	if m.screen != macScreenSkills {
		t.Fatal("skills menu did not open")
	}
	if len(m.rows()) != len(agentSkills)+1 {
		t.Fatal("skills menu has an unexpected catalogue")
	}
}

func TestCategoryNavigationKeepsBasketAndSearchContext(t *testing.T) {
	c := macTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	press(m, "enter")
	if m.screen != macScreenCategories || len(m.rows()) != 13 {
		t.Fatal("Apps did not open categories")
	}
	for i, row := range m.rows() {
		if row.ID == "browsers" {
			m.cursor = i
			break
		}
	}
	press(m, "enter")
	for i, row := range m.rows() {
		if row.ID == "firefox" {
			m.cursor = i
			break
		}
	}
	press(m, "space")
	if !m.selected["firefox"] || len(m.rows()) != 5 {
		t.Fatal("Browsers contains wrong apps")
	}
	press(m, "esc")
	aiCursor := 0
	for i, row := range m.rows() {
		if row.ID == "ai" {
			m.cursor, aiCursor = i, i
			break
		}
	}
	press(m, "enter")
	if m.appGroup != "ai" || len(m.rows()) != 8 {
		t.Fatal("AI category missing")
	}
	for i, row := range m.rows() {
		if row.ID == "ollama-app" {
			m.cursor = i
			break
		}
	}
	press(m, "space")
	press(m, "/")
	for _, r := range "productivity" {
		press(m, string(r))
	}
	if len(m.rows()) != 9 {
		t.Fatal("search does not match category names")
	}
	press(m, "esc")
	if m.screen != macScreenBrowse || m.appGroup != "ai" {
		t.Fatal("search lost its originating category")
	}
	press(m, "esc")
	if m.screen != macScreenCategories || m.cursor != aiCursor {
		t.Fatal("back lost category focus")
	}
	m.screen = macScreenReview
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
	m := newMacModel(macTestContext(t).Paths, "", newMacManifest(), true, time.Second)
	m.height = 20
	if !strings.Contains(stripTerminal(m.headerView(30)), macSigil) {
		t.Fatal("compact header does not use the Magus sigil")
	}
}
