#!/usr/bin/env node
// Walks src/content/commands/**/*.md and emits magus/commands.json.
// Source of truth for the Go TUI's command catalogue.

import { readdir, readFile, writeFile } from 'node:fs/promises';
import { join, dirname, basename } from 'node:path';
import { fileURLToPath } from 'node:url';
import matter from 'gray-matter';

const HERE = dirname(fileURLToPath(import.meta.url));
const ROOT = join(HERE, '..', '..');
const SRC = join(ROOT, 'src', 'content', 'commands');
const OUT = join(HERE, '..', 'commands.json');

const STAGE_ORDER = ['setup', 'install', 'optimise', 'customise', 'gaming'];
const STAGE_SHORT = {
	setup: 'Setup',
	install: 'Apps',
	optimise: 'Optimise',
	customise: 'Customise',
	gaming: 'Gaming',
};
const STAGE_TAGLINE = {
	setup: 'Sudo and Decky basics.',
	install: 'Launchers, browsers, streaming, utilities.',
	optimise: 'Performance and Plasma tweaks.',
	customise: 'Plugins, art, polish.',
	gaming: 'Launchers, emulators, streams. Source ports welcome.',
};
const STAGE_SIGIL = {
	setup: '🜃',
	install: '🜁',
	optimise: '🜂',
	customise: '🜄',
	gaming: '☉',
};
const STAGE_NUM = {
	setup: '01',
	install: '02',
	optimise: '03',
	customise: '04',
	gaming: '05',
};

const GROUPS = {
	install: [
		{
			id: 'capture',
			name: 'Capture & Chat',
			patterns: [/obs/, /discover-overlay/],
		},
		{
			id: 'system',
			name: 'System Tools',
			patterns: [/flatseal/, /warehouse/, /mission-center/],
		},
		{
			id: 'comms',
			name: 'Browsers & Comms',
			patterns: [/brave/, /firefox/, /bitwarden/, /vesktop/],
		},
		{ id: 'media', name: 'Media', patterns: [/vlc/, /spotify/] },
	],
	gaming: [
		{
			id: 'retro',
			name: 'Retro & Emulation',
			patterns: [
				/retroarch/,
				/dolphin/,
				/duckstation/,
				/emudeck/,
				/retrodeck/,
				/mega-bezel/,
				/duimon/,
			],
		},
		{
			id: 'launchers',
			name: 'Launchers & Compat',
			patterns: [
				/protonup/,
				/wine-cellar/,
				/bottles/,
				/cartridges/,
				/heroic/,
				/lutris/,
				/waydroid/,
			],
		},
		{
			id: 'streaming',
			name: 'Streaming & Remote Play',
			patterns: [/geforce/, /chiaki/, /moonlight/, /sunshine/],
		},
		{
			id: 'tools',
			name: 'Tools & Overlays',
			patterns: [/goverlay/, /ludusavi/],
		},
		{
			id: 'ports',
			name: 'Source Ports',
			patterns: [/gzdoom/, /openmw/, /devilutionx/],
		},
	],
};

const PRESETS = [
	{
		id: 'magnum-opus',
		name: 'Magnum Opus',
		tagline: 'the full proven kit',
		patterns: [
			/^setup\/01-set-sudo-password/,
			/^setup\/02-install-dependencies/,
			/^gaming\/200-protonup-qt/,
			/^gaming\/240-heroic/,
			/^optimise\/01-install-cryoutilities/,
			/^optimise\/06-wifi-powersave/,
			/^optimise\/02-tablet-mode/,
			/^customise\/01-decky-install/,
			/^customise\/02-decky-cssloader/,
			/^customise\/03-decky-steamgriddb/,
			/^install\/40-flatseal/,
			/^install\/60-brave/,
		],
	},
	{
		id: 'retro-operator',
		name: 'Retro Operator',
		tagline: 'shaders & bezels',
		patterns: [
			/^setup\/01-set-sudo-password/,
			/^setup\/02-install-dependencies/,
			/^gaming\/100-retroarch/,
			/^gaming\/110-dolphin/,
			/^gaming\/120-duckstation/,
			/^gaming\/140-retrodeck/,
			/^gaming\/160-duimon-mega-bezel/,
			/^gaming\/200-protonup-qt/,
		],
	},
	{
		id: 'hush-mode',
		name: 'Hush Mode',
		tagline: 'a calm Deck',
		patterns: [
			/^setup\/01-set-sudo-password/,
			/^optimise\/06-wifi-powersave/,
			/^optimise\/02-tablet-mode/,
			/^install\/40-flatseal/,
			/^install\/60-brave/,
		],
	},
];

async function walk(dir) {
	const entries = await readdir(dir, { withFileTypes: true });
	const out = [];
	for (const entry of entries) {
		const p = join(dir, entry.name);
		if (entry.isDirectory()) out.push(...(await walk(p)));
		else if (entry.name.endsWith('.md')) out.push(p);
	}
	return out;
}

function groupFor(stageId, cmdSlug) {
	const groups = GROUPS[stageId];
	if (!groups) return null;
	for (const g of groups) {
		if (g.patterns.some((re) => re.test(cmdSlug))) return g.id;
	}
	return null;
}

const paths = await walk(SRC);
const byStage = new Map();

for (const path of paths) {
	const raw = await readFile(path, 'utf8');
	const { data } = matter(raw);
	const stageId = data.category;
	if (!STAGE_ORDER.includes(stageId)) continue;
	const slug = basename(path, '.md');
	const id = `${stageId}/${slug}`;
	const cmd = {
		id,
		slug,
		title: data.title,
		summary: data.summary ?? '',
		order: typeof data.order === 'number' ? data.order : 999,
		groupId: groupFor(stageId, slug),
		danger: data.danger ?? 'low',
		deckOnly: data.deck_only ?? false,
		run: (data.commands ?? []).map((c) => c.run).filter(Boolean),
	};
	if (!byStage.has(stageId)) byStage.set(stageId, []);
	byStage.get(stageId).push(cmd);
}

const stages = STAGE_ORDER.filter((id) => byStage.has(id)).map((id) => {
	const items = byStage.get(id).sort((a, b) => a.order - b.order);
	return {
		id,
		num: STAGE_NUM[id],
		short: STAGE_SHORT[id],
		tagline: STAGE_TAGLINE[id],
		sigil: STAGE_SIGIL[id],
		groups: GROUPS[id]?.map((g) => ({ id: g.id, name: g.name })) ?? [],
		items,
	};
});

const allCmdIds = stages.flatMap((s) => s.items.map((c) => c.id));
const presets = PRESETS.map((p) => {
	const matchIds = allCmdIds.filter((id) => p.patterns.some((re) => re.test(id)));
	return {
		id: p.id,
		name: p.name,
		tagline: p.tagline,
		commandIds: matchIds,
	};
});

const out = { stages, presets };
await writeFile(OUT, JSON.stringify(out, null, 2) + '\n');
console.log(
	`wrote ${OUT} · ${stages.length} stages · ${allCmdIds.length} commands · ${presets.length} presets`,
);
