package main

import "fmt"

// MacPackage carries identity independently from its delivery provider. Summary
// is the concise catalogue description rendered in the Mac TUI.
type MacPackage struct {
	ID, Name, Summary, Kind, AppBundle, Note string
}

var macPackages = []MacPackage{
	{"thefuck", "thefuck", "Suggest corrections for mistyped shell commands.", "formula", "", "Enable in App setups > Modern commands."},
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
	{"font-atkinson-hyperlegible-next", "Atkinson Hyperlegible Next", "Highly legible sans serif designed to make similar characters easier to distinguish.", "cask", "", "Choose the font in your app after installation. You may need to reopen the app."},
	{"font-cascadia-code", "Cascadia Code", "Friendly monospace with programming ligatures for terminals and code editors.", "cask", "", "Choose the font in your app after installation. You may need to reopen the app."},

	{"stats", "Stats", "System monitor for the menu bar.", "cask", "Stats.app", "Requires macOS 12 or later."},
	{"mos", "Mos", "Smooth scrolling with independent mouse scroll direction.", "cask", "Mos.app", "Conflicts with the Mos beta cask."},
	{"hiddenbar", "Hidden Bar", "Hide and reveal menu bar items.", "cask", "Hidden Bar.app", ""},
	{"swiftbar", "SwiftBar", "Customise your menu bar with script-based plugins.", "cask", "SwiftBar.app", "Requires macOS 12 or later. Choose and configure plugins after installation."},
	{"thaw", "Thaw", "Free, open-source menu bar manager with hidden sections, search, profiles and appearance controls.", "cask", "Thaw.app", "Requires macOS 26 or later. Accessibility permission enables item movement; Screen Recording is optional and used for live previews."},
	{"codexbar", "CodexBar", "Menu bar usage monitor for Codex and Claude.", "cask", "CodexBar.app", "Requires macOS 14 or later. Configure your providers after opening the app."},
	{"aldente", "AlDente", "Limit your MacBook's maximum charging percentage.", "cask", "AlDente.app", "Requires macOS 12 or later. Configure charging limits after installation."},
	{"monitorcontrol", "MonitorControl", "Control external-display brightness and volume from the menu bar.", "cask", "MonitorControl.app", "Display support varies; DDC-capable external monitors work best. Grant Accessibility access for keyboard controls if requested."},
	{"firefox", "Firefox", "The independent, privacy-minded browser.", "cask", "Firefox.app", ""},
	{"brave-browser", "Brave", "Privacy-first Chromium browser.", "cask", "Brave Browser.app", ""},
	{"google-chrome", "Google Chrome", "The web developer's default.", "cask", "Google Chrome.app", ""},
	{"librewolf", "LibreWolf", "Firefox-based browser with privacy-focused defaults and no telemetry.", "cask", "LibreWolf.app", "Browser extensions and sync are configured separately."},
	{"ghostty", "Ghostty", "Fast, native, GPU-accelerated terminal.", "cask", "Ghostty.app", ""},
	{"iterm2", "iTerm2", "The classic macOS terminal replacement.", "cask", "iTerm.app", ""},
	{"visual-studio-code", "Visual Studio Code", "The ubiquitous, extensible editor.", "cask", "Visual Studio Code.app", ""},
	{"vscodium", "VSCodium", "Community-built VS Code binaries without Microsoft branding or telemetry.", "cask", "VSCodium.app", "Uses Open VSX by default, so some Microsoft Marketplace extensions may be unavailable."},
	{"coteditor", "CotEditor", "Fast, native plain-text and source-code editor.", "cask", "CotEditor.app", "Includes the cot command-line tool. Choose syntax and text-encoding preferences after installation."},
	{"utm", "UTM", "Run virtual machines on macOS using QEMU and Apple virtualization.", "cask", "UTM.app", "Guest operating-system images are separate downloads and virtual machines can use substantial disk space."},
	{"vimr", "VimR", "Open-source native macOS interface for Neovim.", "cask", "VimR.app", "Requires macOS 14 or later. VimR includes its own Neovim runtime; the separate neovim formula is optional."},
	{"zed", "Zed", "High-performance, multiplayer editor.", "cask", "Zed.app", ""},
	{"obsidian", "Obsidian", "Local-first Markdown notes & knowledge base.", "cask", "Obsidian.app", ""},
	{"joplin", "Joplin", "Open-source notes and to-do lists with optional encrypted sync.", "cask", "Joplin.app", "Requires macOS 12 or later. Sync is optional; configure a provider after installation if wanted."},
	{"libreoffice", "LibreOffice", "Open-source office suite for documents, spreadsheets and presentations.", "cask", "LibreOffice.app", "Installs the current fresh release. Microsoft Office file compatibility is good but complex layouts may differ."},
	{"maccy", "Maccy", "Fast, open-source clipboard history for the menu bar.", "cask", "Maccy.app", "Requires macOS 14 or later. Enable Accessibility access and choose clipboard retention after opening the app."},
	{"hammerspoon", "Hammerspoon", "Automate macOS with Lua scripts and keyboard shortcuts.", "cask", "Hammerspoon.app", "Requires macOS 13 or later. Automation scripts and requested macOS permissions are configured separately."},
	{"localsend", "LocalSend", "Share files privately across nearby devices without an account.", "cask", "LocalSend.app", "Install LocalSend on the other devices too. Local Network permission and the same network are required."},
	{"vlc", "VLC", "Plays absolutely everything.", "cask", "VLC.app", ""},
	{"iina", "IINA", "Modern, open-source media player for macOS.", "cask", "IINA.app", ""},
	{"drawio", "draw.io Desktop", "Open-source diagrams and flowcharts stored locally.", "cask", "draw.io.app", "Requires macOS 13 or later. Online integrations are optional; local files work without an account."},
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
	{"keka", "Keka", "File archiver", "cask", "Keka.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"monocle-app", "Monocle", "Window dimming utility", "cask", "Monocle.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"moonfin", "Moonfin", "Media streaming client for Jellyfin and Emby", "cask", "Moonfin.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"notion", "Notion", "App to write, plan, collaborate, and get organised", "cask", "Notion.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"proton-drive", "Proton Drive", "Client for Proton Drive", "cask", "Proton Drive.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"protonvpn", "ProtonVPN", "VPN client focusing on security", "cask", "ProtonVPN.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"1password", "1Password", "Polished password manager for logins, passkeys and secure documents.", "cask", "1Password.app", "Requires macOS 12 or later and a 1Password account; continued use requires a subscription."},
	{"bitwarden", "Bitwarden", "Open-source password manager with cross-device sync.", "cask", "Bitwarden.app", "Requires macOS 12 or later. Sign in or create an account after installation; paid features are optional."},
	{"keepassxc", "KeePassXC", "Offline, open-source password manager for local encrypted vaults.", "cask", "KeePassXC.app", "Requires macOS 12 or later. You are responsible for backing up and syncing your vault file."},
	{"lulu", "LuLu", "Free, open-source firewall for controlling outgoing connections.", "cask", "LuLu.app", "Approve the required Network Extension in System Settings after installation."},
	{"knockknock", "KnockKnock", "Inspect software that starts persistently on your Mac.", "cask", "KnockKnock.app", "This is an inspection tool, not an antivirus. Review results before removing or changing anything."},
	{"blockblock", "BlockBlock", "Alert when software adds persistent startup components.", "cask", "", "Uses an installer that requests administrator access and system permissions. Review alerts before allowing or blocking anything."},
	{"oversight", "OverSight", "Alert when an app activates the microphone or webcam.", "cask", "", "Requires macOS 12 or later. Uses an installer that requests administrator access and monitoring permissions."},
	{"taskexplorer", "TaskExplorer", "Inspect running processes, signatures, connections and open files.", "cask", "TaskExplorer.app", "An advanced inspection tool. Review findings before terminating processes or changing the system."},
	{"whatsyoursign", "What's Your Sign?", "Show code-signing details for files from Finder.", "cask", "", "Homebrew stages a manual installer. Enable the Finder extension after installation if macOS asks."},
	{"retroarch-metal", "RetroArch", "Frontend for emulators, game engines and media players (Metal graphics API)", "cask", "RetroArch.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"steam", "Steam", "Game store, library and community from Valve.", "cask", "Steam.app", "The Mac client is Intel-only and requires Rosetta 2 on Apple Silicon. Installing Rosetta is a separate, difficult-to-reverse choice."},
	{"heroic", "Heroic Games Launcher", "Open-source launcher for Epic, GOG and Amazon game libraries.", "cask", "Heroic.app", "Requires macOS 12 or later. Sign in to the stores you want to use after installation."},
	{"moonlight", "Moonlight", "Open-source client for streaming games from another computer.", "cask", "Moonlight.app", "Pair it with a compatible host such as Sunshine after installation."},
	{"shottr", "Shottr", "Screenshot measurement and annotation tool", "cask", "Shottr.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"spotify", "Spotify", "Music streaming service", "cask", "Spotify.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"strawberry", "Strawberry", "Open-source music player and organiser for local collections.", "cask", "Strawberry.app", "Choose your music-library folders after opening the app."},
	{"mixxx", "Mixxx", "Open-source DJ software for mixing, recording and live performance.", "cask", "Mixxx.app", "Connect audio hardware and configure your music library after installation."},
	{"lmms", "LMMS", "Open-source music production workstation for composing and arranging.", "cask", "LMMS.app", "Audio devices, plugins and project libraries are configured in the app."},
	{"element", "Element", "Open-source Matrix messenger for private and decentralised conversations.", "cask", "Element.app", "Sign in to or create a Matrix account after installation."},
	{"mattermost", "Mattermost", "Open-source team chat client for self-hosted workspaces.", "cask", "Mattermost.app", "Your organisation provides the server address and account."},
	{"thunderbird", "Thunderbird", "Open-source email and calendar client from Mozilla.", "cask", "Thunderbird.app", "Add your mail and calendar accounts after installation."},
	{"nextcloud", "Nextcloud", "Open-source client for synchronising files with a Nextcloud server.", "cask", "Nextcloud.app", "Connect your own or your organisation's Nextcloud server after installation."},
	{"syncthing", "Syncthing", "Open-source peer-to-peer file synchronisation without a central cloud.", "cask", "Syncthing.app", "Pair devices and choose shared folders after installation."},
	{"cryptomator", "Cryptomator", "Open-source encrypted vaults for files stored in the cloud.", "cask", "Cryptomator.app", "Create or open a vault and keep its recovery details safe."},
	{"unsloth", "Unsloth Desktop", "Desktop application for Unsloth Studio", "cask", "Unsloth.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"upscayl", "Upscayl", "AI image upscaler", "cask", "Upscayl.app", "Check the app's licensing and first-launch setup; some features require an account or subscription."},
	{"vorssaint", "Vorssaint", "Free, open-source menu bar toolkit for monitoring, window management and everyday Mac utilities.", "cask", "Vorssaint.app", "Apple Silicon and macOS 14 or later. No account, telemetry or subscription. Features request only the macOS permissions they need."},
	{"git", "Git", "Distributed version control.", "formula", "", ""},
	{"container", "Apple Container", "Create and run Linux containers with Apple's lightweight virtual machines.", "formula", "", "Requires Apple Silicon, macOS 26 or later and Xcode 26 or later. After installation, choose a kernel with `container system kernel set --recommended`."},
	{"node", "Node.js & npm", "JavaScript runtime with the npm package manager and npx runner.", "formula", "", "Install a project-specific version manager separately if you need to switch Node versions often."},
	{"python@3.14", "Python", "Current Python 3 runtime with pip for installing Python packages.", "formula", "", "Provides python3 and pip3. Use a virtual environment or uv for project dependencies."},
	{"uv", "uv", "Fast Python package and project manager.", "formula", "", "Use uv to create and manage project environments; it can also install Python versions."},
	{"go", "Go", "Go programming language toolchain with gofmt.", "formula", "", "Requires macOS 12 or later."},
	{"rust", "Rust", "Rust compiler, Cargo package manager and formatter.", "formula", "", "If you use rustup, follow Homebrew's guidance to avoid PATH conflicts."},
	{"docker", "Docker CLI", "Command-line client for building and running container images.", "formula", "", "Needs a compatible container runtime such as Colima or Docker Desktop to run containers."},
	{"colima", "Colima", "Lightweight local container runtime for Docker-compatible workflows.", "formula", "", "Run `colima start` after installation. It uses a Linux virtual machine and may download an image."},
	{"jq", "jq", "Terminal-based JSON processor.", "formula", "", ""},
	{"yq", "yq", "Portable YAML / JSON / XML processor.", "formula", "", ""},
	{"ripgrep", "ripgrep", "Blazing-fast recursive search (rg).", "formula", "", ""},
	{"fd", "fd", "Simple, fast alternative to find.", "formula", "", ""},
	{"fzf", "fzf", "Terminal fuzzy finder.", "formula", "", "Enable shortcuts in Terminal setup > Modern commands."},
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
	{"yt-dlp", "YT-DLP", "Download video and audio from YouTube and many other supported sites.", "formula", "", "Use only for media you are permitted to download. FFmpeg enables merging and post-processing additional formats."},
	{"ocrmypdf", "OCRmyPDF", "Add a searchable text layer to scanned PDF files.", "formula", "", "Installs Tesseract and Ghostscript as dependencies. OCR quality and language support depend on the source scan and installed Tesseract data."},
	{"ffmpeg", "FFmpeg", "Play, inspect, convert, record and stream audio and video.", "formula", "", "The standard Homebrew build supports selected codecs; ffmpeg-full is a separate formula with additional dependencies."},
	{"tesseract", "Tesseract", "Extract text from images with an open-source OCR engine.", "formula", "", "The base formula includes English, orientation and script detection, and serial-number data. Install tesseract-lang separately for other languages."},
	{"imagemagick", "ImageMagick", "Convert, resize, inspect and manipulate images from the command line.", "formula", "", "The standard formula supports common formats; imagemagick-full is available separately for additional delegates."},
	{"neovim", "Neovim", "Extensible Vim-based text editor for the terminal.", "formula", "", "Provides the nvim command. Configuration and plugins are separate."},
}

type MacPreset struct {
	Name, Description string
	IDs               []string
}

var macPresets = []MacPreset{
	{"Everyday", "Window management, archives and media. Adds three apps.", []string{"rectangle", "the-unarchiver", "iina"}},
	{"Developer", "A terminal, editor and useful terminal tools.", []string{"ghostty", "visual-studio-code", "git", "jq", "ripgrep", "fd", "fzf", "bat", "git-delta", "lazygit", "eza", "zoxide", "btop", "dust", "duf", "atuin", "tlrc"}},
}

type MacSelection struct {
	AppConfigs  []string     `toml:"app_configs,omitempty"`
	Skills      []string     `toml:"skills,omitempty"`
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
	for _, id := range m.Mac.Skills {
		if _, ok := agentSkillByID(id); !ok {
			return fmt.Errorf("unknown agent skill %q", id)
		}
		if seen[id] {
			return fmt.Errorf("duplicate agent skill %q", id)
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
	for _, id := range m.Mac.Skills {
		steps = append(steps, agentSkillStep{IDValue: id})
	}
	return steps
}
