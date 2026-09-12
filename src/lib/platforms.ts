export const PLATFORMS = [
	{ id: 'steam', name: 'Steam devices', detail: 'Steam Deck & Steam Machine', status: 'Available in alpha', available: true, href: '/steam', icon: 'lucide:gamepad-2', description: 'From handheld to living room. A guided setup for your SteamOS device.' },
	{ id: 'linux', name: 'Linux', detail: 'A foundation of Flatpaks', status: 'Coming soon', available: false, href: '/linux', icon: 'lucide:terminal', description: 'A considered collection of apps, with a Flatpak-focused setup for Linux.' },
	{ id: 'mac', name: 'Mac', detail: 'Apps, settings & portable setups', status: 'Available in alpha', available: true, href: '/mac', icon: 'simple-icons:apple', description: 'A native terminal flow for Homebrew apps, reversible preferences, Ghostty, Zed, Firefox, Zsh and Raycast.' },
] as const;

export type Platform = (typeof PLATFORMS)[number];
export const INSTALL_COMMAND = 'curl -fsSL https://magus.sh/install | sh';
export const LAUNCH_COMMAND = 'magus run';
