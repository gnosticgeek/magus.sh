export const THEMES = [
	{ id: 'tokyo-night', name: 'Tokyo Night', tone: 'dark' },
	{ id: 'grimoire', name: 'Grimoire', tone: 'light' },
	{ id: 'konsole', name: 'Konsole', tone: 'dark' },
] as const;

export type Theme = (typeof THEMES)[number]['id'];
export type Tone = (typeof THEMES)[number]['tone'];

export const THEME_IDS = THEMES.map((t) => t.id) as readonly Theme[];

export const DEFAULT_THEME: Theme = 'tokyo-night';

export const STORAGE_KEY_THEME = 'magus:theme';

export const isTheme = (s: string | null): s is Theme =>
	s !== null && (THEME_IDS as readonly string[]).includes(s);
