package main

import "fmt"

// MacPackage carries identity independently from its delivery provider. The
// descriptions and preset approach are adapted from Machinist's AthanorKit.
type MacPackage struct {
	ID, Name, Summary, Kind, AppBundle, Note string
}

var macPackages = []MacPackage{
	{"thefuck", "thefuck", "Suggest corrections for mistyped shell commands.", "formula", "", "Enable in App setups > Modern CLI."},
	{"dust", "dust", "Find large folders with a visual disk-usage tree.", "formula", "", "Run dust; du keeps its original behaviour."},
	{"duf", "duf", "Readable disk-space tables.", "formula", "", "Run duf; df keeps its original behaviour."},
	{"atuin", "Atuin", "Searchable shell history with optional sync.", "formula", "", "Enable in Terminal setup > Modern commands. Account and sync setup are separate."},
	{"tlrc", "tldr", "Practical command examples from tldr pages.", "formula", "", "Official Rust client. Run tldr; man remains available."},
	{"font-inter", "Inter", "Polished, readable sans serif for interfaces, presentations and everyday documents.", "cask", "", "Choose the font in your app after installation. You may need to reopen the app."},
	{"font-source-serif-4", "Source Serif 4", "Quietly elegant serif for reports, essays and long documents.", "cask", "", "Choose the font in your app after installation. You may need to reopen the app."},
	{"font-newsreader", "Newsreader", "Literary editorial serif with graceful italics for articles and newsletters.", "cask", "", "Choose the font in your app after installation. You may need to reopen the app."},
	{"font-fraunces", "Fraunces", "Warm, expressive serif for memorable headings, branding and invitations.", "cask", "", "Choose the font in your app after installation. You may need to reopen the app."},
	{"font-space-grotesk", "Space Grotesk", "Distinctive geometric sans serif for headings, portfolios and modern branding.", "cask", "", "Choose the font in your app after installation. You may need to reopen the app."},
	{"font-jetbrains-mono", "JetBrains Mono", "Clear, comfortable monospace for coding and terminals.", "cask", "", "Choose the font in your app after installation. You may need to reopen the app."},

	{"stats", "Stats", "System monitor for the menu bar.", "cask", "Stats.app", "Requires macOS 12 or later."},
	{"mos", "Mos", "Smooth scrolling with independent mouse scroll direction.", "cask", "Mos.app", "Conflicts with the Mos beta cask."},
	{"hiddenbar", "Hidden Bar", "Hide and reveal menu bar items.", "cask", "Hidden Bar.app", ""},
	{"swiftbar", "SwiftBar", "Customise your menu bar with script-based plugins.", "cask", "SwiftBar.app", "Requires macOS 12 or later. Choose and configure plugins after installation."},
	{"thaw", "Thaw", "Organise and manage your menu bar items.", "cask", "Thaw.app", "Requires macOS 26 or later."},
	{"codexbar", "CodexBar", "Menu bar usage monitor for Codex and Claude.", "cask", "CodexBar.app", "Requires macOS 14 or later. Configure your providers after opening the app."},
	{"aldente", "AlDente", "Limit your MacBook's maximum charging percentage.", "cask", "AlDente.app", "Requires macOS 12 or later. Configure charging limits after installation."},
	{"firefox", "Firefox", "The independent, privacy-minded browser.", "cask", "Firefox.app", ""},
	{"brave-browser", "Brave", "Privacy-first Chromium browser.", "cask", "Brave Browser.app", ""},
	{"google-chrome", "Google Chrome", "The web developer's default.", "cask", "Google Chrome.app", ""},
	{"ghostty", "Ghostty", "Fast, native, GPU-accelerated terminal.", "cask", "Ghostty.app", ""},
	{"iterm2", "iTerm2", "The classic macOS terminal replacement.", "cask", "iTerm.app", ""},
	{"visual-studio-code", "Visual Studio Code", "The ubiquitous, extensible editor.", "cask", "Visual Studio Code.app", ""},
	{"zed", "Zed", "High-performance, multiplayer editor.", "cask", "Zed.app", ""},
	{"obsidian", "Obsidian", "Local-first Markdown notes & knowledge base.", "cask", "Obsidian.app", ""},
	{"vlc", "VLC", "Plays absolutely everything.", "cask", "VLC.app", ""},
	{"iina", "IINA", "Modern, open-source media player for macOS.", "cask", "IINA.app", ""},
	{"rectangle", "Rectangle", "Move & resize windows with shortcuts.", "cask", "Rectangle.app", "Enable Accessibility access when opening Rectangle."},
	{"raycast", "Raycast", "Blazing-fast launcher & command palette.", "cask", "Raycast.app", ""},
	{"the-unarchiver", "The Unarchiver", "Extract just about any archive.", "cask", "The Unarchiver.app", ""},
	{"signal", "Signal", "End-to-end encrypted private messenger.", "cask", "Signal.app", "Link your phone after installation."},
	{"discord", "Discord", "Voice, video & text for communities.", "cask", "Discord.app", ""},
	{"ollama-app", "Ollama", "Run and chat with language models on your Mac.", "cask", "Ollama.app", "Requires macOS 14 or later. Download models separately after opening the app; they can use several GB."},
	{"lm-studio", "LM Studio", "Discover, download and run local models with a GUI.", "cask", "LM Studio.app", "The Homebrew cask requires Apple Silicon and macOS 12 or later. Models are downloaded separately and can use several GB."},
	{"affinity", "Affinity", "Image editing and design software", "cask", "Affinity.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"aside", "Aside", "Web browser with built-in AI assistant", "cask", "Aside.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"audacity", "Audacity", "Multi-track audio editor and recorder", "cask", "Audacity 4.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"balenaetcher", "Etcher", "Tool to flash OS images to SD cards & USB drives", "cask", "balenaEtcher.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"bambu-studio", "Bambu Studio", "3D model slicing software for 3D printers, maintained by Bambu Lab", "cask", "BambuStudio.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"caskhub", "CaskHub", "Native GUI for Homebrew casks", "cask", "CaskHub.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"chatgpt", "ChatGPT", "OpenAI's official ChatGPT desktop app", "cask", "ChatGPT.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"claude", "Claude", "Anthropic's official Claude AI desktop app", "cask", "Claude.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"dockflow", "DockFlow", "Manage Dock presets and switch between them instantly", "cask", "DockFlow.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"downie", "Downie", "Downloads videos from different websites", "cask", "Downie 4.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"es-de", "ES-DE", "Frontend for browsing and launching games from your multi-platform collection", "cask", "ES-DE.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"fluidvoice", "FluidVoice", "Offline voice-to-text dictation app with AI enhancement", "cask", "FluidVoice.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"grok-bot", "Grok Bot", "AI teammates that work across your apps and tools", "cask", "Grok Bot.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"handbrake-app", "HandBrake", "Open-source video transcoder", "cask", "HandBrake.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"hermes-desktop", "Hermes Agent Desktop", "Open-source desktop AI agent", "cask", "Hermes.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"jellyfin-media-player", "Jellyfin Media Player", "Jellyfin desktop client", "cask", "Jellyfin Media Player.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"jordanbaird-ice", "Ice", "Menu bar manager", "cask", "Ice.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"keka", "Keka", "File archiver", "cask", "Keka.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"monocle-app", "Monocle", "Window dimming utility", "cask", "Monocle.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"moonfin", "Moonfin", "Media streaming client for Jellyfin and Emby", "cask", "Moonfin.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"notion", "Notion", "App to write, plan, collaborate, and get organised", "cask", "Notion.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"proton-drive", "Proton Drive", "Client for Proton Drive", "cask", "Proton Drive.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"protonvpn", "ProtonVPN", "VPN client focusing on security", "cask", "ProtonVPN.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"retroarch-metal", "RetroArch", "Frontend for emulators, game engines and media players (Metal graphics API)", "cask", "RetroArch.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"shottr", "Shottr", "Screenshot measurement and annotation tool", "cask", "Shottr.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"spotify", "Spotify", "Music streaming service", "cask", "Spotify.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"unsloth", "Unsloth Desktop", "Desktop application for Unsloth Studio", "cask", "Unsloth.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"upscayl", "Upscayl", "AI image upscaler", "cask", "Upscayl.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"vorssaint", "Vorssaint", "Menu bar toolkit with keep-awake, system monitor and volume mixer", "cask", "Vorssaint.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"git", "Git", "Distributed version control.", "formula", "", ""},
	{"jq", "jq", "Command-line JSON processor.", "formula", "", ""},
	{"yq", "yq", "Portable YAML / JSON / XML processor.", "formula", "", ""},
	{"ripgrep", "ripgrep", "Blazing-fast recursive search (rg).", "formula", "", ""},
	{"fd", "fd", "Simple, fast alternative to find.", "formula", "", ""},
	{"fzf", "fzf", "Command-line fuzzy finder.", "formula", "", "Enable shortcuts in Terminal setup > Modern commands."},
	{"bat", "bat", "cat clone with syntax highlighting.", "formula", "", ""},
	{"eza", "eza", "Modern, colourful replacement for ls.", "formula", "", ""},
	{"btop", "btop", "Resource monitor with a gorgeous TUI.", "formula", "", ""},
	{"fastfetch", "fastfetch", "Fast system information tool.", "formula", "", ""},
	{"zoxide", "zoxide", "Smarter cd that learns your habits.", "formula", "", "Enable cd integration in Terminal setup > Modern commands."},
	{"direnv", "direnv", "Load & unload env vars per directory.", "formula", "", "Optional: add eval \"$(direnv hook zsh)\" to your shell configuration yourself."},
	{"git-delta", "delta", "Beautiful syntax-highlighted git diffs.", "formula", "", "Enable the Git pager in Terminal setup > Modern commands."},
	{"lazygit", "lazygit", "A terminal UI for git commands.", "formula", "", ""},
	{"tmux", "tmux", "Terminal multiplexer for persistent sessions.", "formula", "", ""},
	{"mole", "Mole", "Mac maintenance utility for app removal, cleanup and disk analysis.", "formula", "", "Run mole after installation to choose a maintenance action. Installing it does not run cleanup or remove apps."},
}

type MacPreset struct {
	Name, Description string
	IDs               []string
}

var macPresets = []MacPreset{
	{"Everyday", "Window management, archives and media. Adds three apps.", []string{"rectangle", "the-unarchiver", "iina"}},
	{"Developer", "A terminal, editor and useful command-line tools.", []string{"ghostty", "visual-studio-code", "git", "jq", "ripgrep", "fd", "fzf", "bat", "git-delta", "lazygit", "eza", "zoxide", "btop", "dust", "duf", "atuin", "tlrc"}},
}

type MacSelection struct {
	AppConfigs  []string     `toml:"app_configs,omitempty"`
	ModernShell *modernShell `toml:"modern_shell,omitempty"`
	Terminal    string       `toml:"terminal,omitempty"`
	Packages    []string     `toml:"packages"`
	Settings    []string     `toml:"settings"`
}

func newMacManifest() Manifest {
	return Manifest{Magus: MagusSection{Version: Version, Platform: "darwin", Device: "mac"}}
}
func macPackage(id string) (MacPackage, bool) {
	for _, p := range macPackages {
		if p.ID == id {
			return p, true
		}
	}
	return MacPackage{}, false
}
func validateMacManifest(m Manifest) error {
	if m.Magus.Version != Version {
		return fmt.Errorf("unsupported Mac manifest schema %q", m.Magus.Version)
	}
	if m.Magus.Platform != "darwin" || m.Magus.Device != "mac" {
		return fmt.Errorf("Mac manifest requires platform darwin and device mac")
	}
	if m.Choices.Terminal != "" || m.Choices.Browser != "" || len(m.Choices.Bundles) > 0 || m.Choices.Theme != "" || m.Optimisations != (OptimisationsSection{}) {
		return fmt.Errorf("Mac manifest contains Steam/Linux choices")
	}
	if m.Mac.Terminal != "" && !oneOf(m.Mac.Terminal, terminalThemes) {
		return fmt.Errorf("unknown terminal profile %q", m.Mac.Terminal)
	}
	if m.Mac.ModernShell != nil {
		if err := m.Mac.ModernShell.validate(); err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	for _, id := range m.Mac.AppConfigs {
		if _, ok := appConfig(id); !ok {
			return fmt.Errorf("unknown app configuration %q", id)
		}
		if seen[id] {
			return fmt.Errorf("duplicate app configuration %q", id)
		}
		seen[id] = true
	}
	seen = map[string]bool{}
	for _, id := range m.Mac.Packages {
		if _, ok := macPackage(id); !ok {
			return fmt.Errorf("unknown Mac package %q", id)
		}
		if seen[id] {
			return fmt.Errorf("duplicate package %q", id)
		}
		seen[id] = true
	}
	for _, id := range m.Mac.Settings {
		if _, ok := macSetting(id); !ok {
			return fmt.Errorf("unknown Mac setting %q", id)
		}
		if seen[id] {
			return fmt.Errorf("duplicate setting %q", id)
		}
		seen[id] = true
	}
	return nil
}
func manifestPlatformCheck(m Manifest, platform string) error {
	target := m.Magus.Platform
	if target == "" {
		target = "linux"
	}
	if target != platform {
		return fmt.Errorf("this %s manifest cannot run on %s; use a separate manifest for this machine", target, platform)
	}
	if !oneOf(m.Magus.Version, []string{"0.1.0", "0.2.0", "0.3.0", Version}) {
		return fmt.Errorf("unsupported manifest schema %q", m.Magus.Version)
	}
	return nil
}
func macSteps(m Manifest) []Step {
	var steps []Step
	packages := append([]string{}, m.Mac.Packages...)
	if m.Mac.ModernShell != nil {
		packages = append(packages, m.Mac.ModernShell.packages()...)
	}
	for _, id := range m.Mac.AppConfigs {
		packages = append(packages, id)
	}
	for _, p := range macPackages {
		if oneOf(p.ID, packages) {
			steps = append(steps, brewStep{Package: p})
		}
	}
	for _, s := range macSettings {
		if oneOf(s.ID, m.Mac.Settings) {
			steps = append(steps, preferenceStep{Setting: s})
		}
	}
	if m.Mac.Terminal != "" {
		steps = append(steps, terminalStep{m.Mac.Terminal})
	}
	if m.Mac.ModernShell != nil {
		steps = append(steps, shellStep{*m.Mac.ModernShell})
	}
	for _, id := range m.Mac.AppConfigs {
		steps = append(steps, appConfigStep{IDValue: id})
	}
	return steps
}
