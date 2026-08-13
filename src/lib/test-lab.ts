/* Scratch bench for the /test page.
   Candidate commands that have NOT been proven on Steam Machine hardware yet.
   Nothing here is part of the `commands` content collection on purpose — this
   is where a command lives until it earns a card on /setup. */

export type LabCommand = {
	run: string;
	note?: string;
};

export type LabTest = {
	id: string;
	title: string;
	/* what we don't know yet — the actual experiment */
	question: string;
	/* what a pass looks like on the machine */
	expect: string;
	commands: LabCommand[];
	danger?: 'low' | 'medium' | 'high';
	needsDesktop?: boolean;
	undo?: string;
	ref?: { name: string; url: string };
};

export type LabBatch = {
	id: string;
	name: string;
	tagline: string;
	icon: string;
	tests: LabTest[];
};

export const LAB_BATCHES: LabBatch[] = [
	{
		id: 'probe',
		name: 'Probe',
		tagline: 'Establish what the box actually is before trusting any result below.',
		icon: 'lucide:scan-line',
		tests: [
			{
				id: 'probe-os',
				title: 'SteamOS build & kernel',
				question:
					'Which SteamOS release and kernel does the Steam Machine ship, and does it differ from Deck 3.x?',
				expect:
					'VERSION_ID and BUILD_ID from os-release; record them — every other result on this page is only valid for this build.',
				commands: [
					{ run: 'cat /etc/os-release' },
					{ run: 'uname -srm' },
					{
						run: 'hostnamectl | sed -n "1,12p"',
						note: 'Chassis / hardware vendor line tells us how the machine self-identifies.',
					},
				],
			},
			{
				id: 'probe-session',
				title: 'Session & compositor',
				question:
					'Is the desktop session the same Plasma/gamescope pairing as the Deck, and is it Wayland-only?',
				expect:
					'XDG_CURRENT_DESKTOP=KDE, XDG_SESSION_TYPE=wayland, and a running gamescope process in Game Mode.',
				commands: [
					{ run: 'echo "desktop=$XDG_CURRENT_DESKTOP session=$XDG_SESSION_TYPE"' },
					{ run: 'pgrep -a gamescope | head -3' },
					{ run: 'plasmashell --version' },
				],
			},
			{
				id: 'probe-readonly',
				title: 'Read-only rootfs',
				question:
					'Is / still immutable? Anything that writes outside $HOME needs this answered first.',
				expect:
					'"enabled" — themes must then live under ~/ (which survives updates) and never in /usr.',
				commands: [
					{ run: 'sudo steamos-readonly status' },
					{
						run: 'findmnt -no FSTYPE,OPTIONS / /home',
						note: 'Confirms the btrfs/ext4 split and which mount is writable.',
					},
				],
				danger: 'low',
			},
		],
	},
	{
		id: 'theming',
		name: 'Theming',
		tagline:
			'Does the Deck theming stack — CSS Loader, Plasma, boot video — transfer to the Steam Machine untouched?',
		icon: 'lucide:palette',
		tests: [
			{
				id: 'theme-decky-present',
				title: 'Decky Loader service',
				question:
					'Does Decky Loader install and stay running on this build? CSS Loader is a Decky plugin, so everything downstream depends on it.',
				expect:
					'plugin_loader.service active (running), and ~/homebrew/plugins exists.',
				commands: [
					{ run: 'systemctl status plugin_loader --no-pager | sed -n "1,6p"' },
					{ run: 'ls -1 ~/homebrew/plugins 2>/dev/null || echo "no decky plugins dir"' },
				],
				ref: {
					name: 'SteamDeckHomebrew/decky-loader',
					url: 'https://github.com/SteamDeckHomebrew/decky-loader',
				},
			},
			{
				id: 'theme-cssloader',
				title: 'CSS Loader installer',
				question:
					"Does the upstream installer's Deck assumptions (paths, service restart) hold on Steam Machine?",
				expect:
					'Installer exits 0, plugin appears in the Decky menu in Game Mode, no restart loop in the service log.',
				commands: [
					{
						run: 'curl -L https://github.com/suchmememanyskill/SDH-CssLoader/raw/main/install_release.sh | sh',
						note: 'External script — read it once before piping if you have not already.',
					},
					{ run: 'journalctl -u plugin_loader -n 20 --no-pager' },
				],
				danger: 'low',
				undo:
					'rm -rf ~/homebrew/plugins/SDH-CssLoader && sudo systemctl restart plugin_loader.service',
				ref: {
					name: 'suchmememanyskill/SDH-CssLoader',
					url: 'https://github.com/suchmememanyskill/SDH-CssLoader',
				},
			},
			{
				id: 'theme-manifest',
				title: 'Hand-rolled theme + manifest version',
				question:
					'Which CSS Loader manifest_version does this build accept? Writing a minimal theme by hand isolates that from any store download.',
				expect:
					'"magus test" shows up in the CSS Loader plugin list and the Quick Access panel picks up the accent colour.',
				commands: [
					{
						run: `mkdir -p ~/homebrew/themes/magus-test && cat > ~/homebrew/themes/magus-test/theme.json <<'JSON'
{
  "name": "magus test",
  "version": "v1.0",
  "author": "magus.sh",
  "manifest_version": 8,
  "inject": { "shared.css": ["QuickAccess", "SP", "Desktop"] }
}
JSON`,
						note: 'Bump manifest_version and re-run if the plugin rejects it — that number is the actual thing under test.',
					},
					{
						run: `cat > ~/homebrew/themes/magus-test/shared.css <<'CSS'
:root { --magus-probe: #7aa2f7; }
body { outline: 2px solid var(--magus-probe); outline-offset: -2px; }
CSS`,
						note: 'A visible outline is a blunt but unmistakable "the CSS landed" signal.',
					},
					{ run: 'sudo systemctl restart plugin_loader.service' },
				],
				danger: 'low',
				undo: 'rm -rf ~/homebrew/themes/magus-test && sudo systemctl restart plugin_loader.service',
			},
			{
				id: 'theme-plasma-colors',
				title: 'Plasma colour scheme',
				question:
					'Do the plasma-apply-* CLI tools exist on this image, or has Valve trimmed them?',
				expect:
					'A scheme list prints, and applying one repaints the desktop session immediately.',
				commands: [
					{ run: 'plasma-apply-colorscheme --list-schemes' },
					{ run: 'plasma-apply-colorscheme BreezeDark' },
				],
				needsDesktop: true,
				undo: 'plasma-apply-colorscheme BreezeDark',
			},
			{
				id: 'theme-plasma-global',
				title: 'Global (look-and-feel) theme',
				question:
					'Does lookandfeeltool survive here, and does applying a global theme break the Return-to-Game-Mode shortcut?',
				expect:
					'Theme list prints; after applying, the Return to Gaming Mode desktop icon still works.',
				commands: [
					{ run: 'lookandfeeltool -l' },
					{ run: 'lookandfeeltool -a org.kde.breezedark.desktop' },
				],
				needsDesktop: true,
				danger: 'medium',
				undo: 'lookandfeeltool -a org.kde.breezedark.desktop',
			},
			{
				id: 'theme-cursor-icons',
				title: 'Cursor, icon & desktop theme',
				question:
					'Are the Steam-branded cursor and icon themes present, and do Wayland clients pick up a change without a relog?',
				expect: 'Each --list-themes prints; the cursor changes without logging out.',
				commands: [
					{ run: 'plasma-apply-cursortheme --list-themes' },
					{ run: 'plasma-apply-desktoptheme --list-themes' },
					{ run: 'ls -1 /usr/share/icons ~/.local/share/icons 2>/dev/null' },
				],
				needsDesktop: true,
			},
			{
				id: 'theme-flatpak-gtk',
				title: 'Flatpak apps following the system theme',
				question:
					'Do Flatpaks read the host GTK theme once the config dirs are exposed, or do they stay Adwaita-light?',
				expect:
					'Overrides apply cleanly and a relaunched Flatpak (Brave, VLC) matches the desktop.',
				commands: [
					{
						run: 'flatpak override --user --filesystem=xdg-config/gtk-3.0:ro --filesystem=xdg-config/gtk-4.0:ro',
					},
					{ run: 'gsettings get org.gnome.desktop.interface color-scheme' },
					{ run: 'flatpak override --user --show' },
				],
				needsDesktop: true,
				undo: 'flatpak override --user --reset',
			},
			{
				id: 'theme-boot-video',
				title: 'Boot video override',
				question:
					'Does the Steam Machine honour the Deck uioverrides/movies path for the startup animation?',
				expect:
					'Directory exists (or can be created) and a webm dropped in as deck_startup.webm plays on next boot.',
				commands: [
					{ run: 'mkdir -p ~/.steam/root/config/uioverrides/movies' },
					{ run: 'ls -la ~/.steam/root/config/uioverrides/movies' },
					{
						run: 'ls -1 ~/.steam/root/steamui/movies 2>/dev/null | head',
						note: 'Stock animations — filenames here tell us what an override has to be called on this build.',
					},
				],
				danger: 'low',
				undo: 'rm -f ~/.steam/root/config/uioverrides/movies/deck_startup.webm',
			},
			{
				id: 'theme-fonts',
				title: 'Custom fonts',
				question:
					'Do user fonts in ~/.local/share/fonts get picked up by both Plasma and the Steam client?',
				expect:
					'fc-cache sees the new font and it is selectable in Plasma settings; note whether Game Mode also picks it up.',
				commands: [
					{ run: 'mkdir -p ~/.local/share/fonts && fc-cache -fv ~/.local/share/fonts | tail -3' },
					{ run: 'fc-list : family | sort -u | wc -l' },
				],
				undo: 'rm -rf ~/.local/share/fonts && fc-cache -f',
			},
			{
				id: 'theme-persistence',
				title: 'Survives a SteamOS update',
				question:
					'Which of the above live on the writable /home partition and therefore survive an OS update?',
				expect:
					'Everything under ~/homebrew, ~/.local, ~/.steam resolves inside /home — anything that does not will be wiped.',
				commands: [
					{
						run: 'for p in ~/homebrew ~/.local/share/fonts ~/.steam/root/config/uioverrides ~/.config/kdeglobals; do printf "%-46s %s\\n" "$p" "$(df --output=target "$p" 2>/dev/null | tail -1)"; done',
					},
				],
			},
		],
	},
];

export const LAB_TEST_COUNT = LAB_BATCHES.reduce((n, b) => n + b.tests.length, 0);

export const STORAGE_KEY_LAB = 'magus:test-lab';
