export const STAGES = [
	{
		id: 'setup',
		num: '01',
		label: 'one',
		short: 'Setup',
		title: 'Setup',
		tagline: 'Sudo and Decky basics.',
		stageLabel: 'stage_01',
		rise: 2,
	},
	{
		id: 'install',
		num: '02',
		label: 'two',
		short: 'Apps',
		title: 'Apps & Tools',
		tagline: 'Launchers, browsers, streaming, utilities.',
		stageLabel: 'stage_02',
		rise: 3,
	},
	{
		id: 'optimise',
		num: '03',
		label: 'three',
		short: 'Optimise',
		title: 'Optimise',
		tagline: 'Performance and Plasma tweaks.',
		stageLabel: 'stage_03',
		rise: 4,
	},
	{
		id: 'customise',
		num: '04',
		label: 'four',
		short: 'Customise',
		title: 'Customise',
		tagline: 'Plugins, art, polish.',
		stageLabel: 'stage_04',
		rise: 5,
	},
	{
		id: 'gaming',
		num: '05',
		label: 'five',
		short: 'Gaming',
		title: 'Gaming',
		tagline: 'Launchers, emulators, streams. Source ports welcome.',
		stageLabel: 'stage_05',
		rise: 5,
		variant: 'sigil',
	},
] as const;

export type Stage = (typeof STAGES)[number]['id'];

export const STAGE_IDS = STAGES.map((s) => s.id) as readonly Stage[];

export const STAGE_LABELS: Record<Stage, string> = Object.fromEntries(
	STAGES.map((s) => [s.id, s.short]),
) as Record<Stage, string>;

export const STAGE_SIGILS: Record<Stage, { char: string; name: string }> = {
	setup: { char: '\u{1F703}', name: 'Earth' },
	install: { char: '\u{1F701}', name: 'Air' },
	optimise: { char: '\u{1F702}', name: 'Fire' },
	customise: { char: '\u{1F704}', name: 'Water' },
	gaming: { char: '☉', name: 'Sun' },
};

export const isStage = (s: string): s is Stage =>
	(STAGE_IDS as readonly string[]).includes(s);

export type CommandPattern = string;

export const matchPattern = (cmdId: string, pat: CommandPattern): boolean =>
	cmdId.includes(pat);

export type Group = {
	id: string;
	name: string;
	patterns: CommandPattern[];
};

export const STAGE_GROUPS: Partial<Record<Stage, Group[]>> = {
	install: [
		{
			id: 'capture',
			name: 'Capture & Chat',
			patterns: ['discover-overlay', '-obs'],
		},
		{
			id: 'system',
			name: 'System Tools',
			patterns: [
				'flatseal',
				'warehouse',
				'mission-center',
				'distrobox',
				'protontricks',
			],
		},
		{
			id: 'comms',
			name: 'Browsers & Comms',
			patterns: ['brave', 'firefox', 'vesktop', 'bitwarden'],
		},
		{
			id: 'media',
			name: 'Media',
			patterns: ['-vlc', 'spotify'],
		},
	],
	gaming: [
		{
			id: 'retro',
			name: 'Retro & Emulation',
			patterns: [
				'retroarch',
				'pcsx2',
				'rpcs3',
				'steam-rom-manager',
				'shadps4',
				'dolphin',
				'duckstation',
				'ppsspp',
				'emudeck',
				'retrodeck',
				'bsnes',
				'desmume',
				'melonds',
				'cemu',
				'xemu',
				'mame',
				'flycast',
				'mega-bezel',
				'duimon',
			],
		},
		{
			id: 'launchers',
			name: 'Launchers & Compat',
			patterns: [
				'protonup',
				'wine-cellar',
				'bottles',
				'cartridges',
				'heroic',
				'lutris',
				'waydroid',
				'nonsteamlaunchers',
				'minigalaxy',
				'prism-launcher',
				'itch',
			],
		},
		{
			id: 'streaming',
			name: 'Streaming & Remote Play',
			patterns: [
				'geforce-now',
				'chiaki',
				'moonlight',
				'sunshine',
				'localsend',
				'steam-link',
				'greenlight',
				'parsec',
			],
		},
		{
			id: 'tools',
			name: 'Tools & Overlays',
			patterns: [
				'goverlay',
				'ludusavi',
				'boilr',
				'syncthing',
				'shortix',
				'shader-cache-killer',
			],
		},
		{
			id: 'ports',
			name: 'Source Ports',
			patterns: ['gzdoom', 'openmw', 'devilutionx'],
		},
	],
};

export type Preset = {
	id: string;
	name: string;
	tagline: string;
	patterns: CommandPattern[];
};

export const PRESETS: Preset[] = [
	{
		id: 'magnum-opus',
		name: 'Magnum Opus',
		tagline: 'the full proven kit',
		patterns: [
			'set-sudo-password',
			'install-dependencies',
			'protonup-qt',
			'heroic',
			'install-cryoutilities',
			'wifi-powersave',
			'tablet-mode',
			'decky-install',
			'decky-cssloader',
			'decky-steamgriddb',
			'flatseal',
			'brave',
		],
	},
	{
		id: 'retro-operator',
		name: 'Retro Operator',
		tagline: 'shaders & bezels',
		patterns: [
			'set-sudo-password',
			'install-dependencies',
			'retroarch',
			'dolphin',
			'duckstation',
			'retrodeck',
			'duimon-mega-bezel',
			'protonup-qt',
		],
	},
	{
		id: 'hush-mode',
		name: 'Hush Mode',
		tagline: 'a calm Deck',
		patterns: [
			'set-sudo-password',
			'wifi-powersave',
			'tablet-mode',
			'flatseal',
			'brave',
		],
	},
];
