export const STAGES = [
	{ id: 'setup', num: '01', label: 'one', short: 'Setup' },
	{ id: 'install', num: '02', label: 'two', short: 'Install' },
	{ id: 'optimise', num: '03', label: 'three', short: 'Optimise' },
	{ id: 'customise', num: '04', label: 'four', short: 'Customise' },
	{ id: 'retro', num: '05', label: 'five', short: 'Retro' },
] as const;

export type Stage = (typeof STAGES)[number]['id'];

export const STAGE_IDS = STAGES.map((s) => s.id) as readonly Stage[];

export const STAGE_LABELS: Record<Stage, string> = Object.fromEntries(
	STAGES.map((s) => [s.id, s.short]),
) as Record<Stage, string>;

export const isStage = (s: string): s is Stage =>
	(STAGE_IDS as readonly string[]).includes(s);
