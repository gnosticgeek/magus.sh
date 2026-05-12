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
		id: 'retro',
		num: '05',
		label: 'five',
		short: 'Retro',
		title: 'Retro',
		tagline: 'Cartridge dust, phosphor glow. PS1, GameCube, Wii.',
		stageLabel: 'stage_05',
		rise: 5,
		variant: 'retro',
	},
] as const;

export type Stage = (typeof STAGES)[number]['id'];

export const STAGE_IDS = STAGES.map((s) => s.id) as readonly Stage[];

export const STAGE_LABELS: Record<Stage, string> = Object.fromEntries(
	STAGES.map((s) => [s.id, s.short]),
) as Record<Stage, string>;

export const isStage = (s: string): s is Stage =>
	(STAGE_IDS as readonly string[]).includes(s);
