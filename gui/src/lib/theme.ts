/**
 * Theming.
 *
 * One palette drives everything, which is the §4-step-5 claim made literal: the
 * same four names the wizard offers appear here, and switching one sets a
 * single attribute on <html>. Every colour in the app resolves through a CSS
 * custom property, so no component knows a hex value.
 *
 * These palettes are the same ones the Plasma colour scheme and kitty config
 * are generated from. The eventual single source is a palette file the Go side
 * also reads; until that exists, these values must match `magus/mockups/*.html`.
 */

export type ThemeId = 'tokyo-night' | 'gruvbox' | 'catppuccin' | 'nord';

export interface Theme {
	id: ThemeId;
	name: string;
	/** Four representative swatches, for the picker. */
	chips: string[];
}

export const THEMES: Theme[] = [
	{ id: 'tokyo-night', name: 'Tokyo Night', chips: ['#1a1b26', '#7aa2f7', '#bb9af7', '#9ece6a'] },
	{ id: 'gruvbox', name: 'Gruvbox', chips: ['#282828', '#fabd2f', '#fe8019', '#b8bb26'] },
	{ id: 'catppuccin', name: 'Catppuccin', chips: ['#1e1e2e', '#cba6f7', '#89b4fa', '#a6e3a1'] },
	{ id: 'nord', name: 'Nord', chips: ['#2e3440', '#88c0d0', '#81a1c1', '#a3be8c'] },
];

const STORAGE_KEY = 'magus.theme';

export function applyTheme(id: ThemeId): void {
	if (typeof document === 'undefined') return;
	document.documentElement.dataset.theme = id;
	try {
		localStorage.setItem(STORAGE_KEY, id);
	} catch {
		// Private browsing, or storage disabled. The theme still applies for
		// this session; only the memory of it is lost.
	}
}

export function loadTheme(): ThemeId {
	if (typeof localStorage === 'undefined') return 'tokyo-night';
	const saved = localStorage.getItem(STORAGE_KEY);
	return THEMES.some((t) => t.id === saved) ? (saved as ThemeId) : 'tokyo-night';
}
