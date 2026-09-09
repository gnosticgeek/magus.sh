export const PLATFORMS = [
	{ id: 'steam', name: 'Steam devices', detail: 'Steam Deck & Steam Machine', status: 'Available in alpha', available: true, href: '/steam', icon: 'lucide:gamepad-2', description: 'From handheld to living room. A guided setup for your SteamOS device.' },
	{ id: 'linux', name: 'Linux', detail: 'A foundation of Flatpaks', status: 'Coming soon', available: false, href: '/linux', icon: 'lucide:terminal', description: 'A considered collection of apps, with a Flatpak-focused setup for Linux.' },
	{ id: 'mac', name: 'Mac', detail: 'A fresh start for macOS', status: 'Coming soon', available: false, href: '/mac', icon: 'simple-icons:apple', description: 'The same thoughtful approach, with tools and defaults chosen for your Mac.' },
] as const;

export type Platform = (typeof PLATFORMS)[number];
export const INSTALL_COMMAND = 'curl -fsSL https://magus.sh/install | sh';
export const LAUNCH_COMMAND = 'magus run';
