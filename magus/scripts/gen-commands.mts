#!/usr/bin/env tsx
// Walks src/content/commands/**/*.md and emits magus/commands.json.
// Source of truth for the Go TUI's command catalogue.
// Stage metadata, groups, and presets all come from src/lib/stages.ts.

import { readdir, readFile, writeFile } from 'node:fs/promises';
import { join, dirname, basename } from 'node:path';
import { fileURLToPath } from 'node:url';
import matter from 'gray-matter';
import {
	STAGES,
	STAGE_SIGILS,
	STAGE_GROUPS,
	PRESETS,
	matchPattern,
	type Stage,
} from '../../src/lib/stages.ts';

const HERE = dirname(fileURLToPath(import.meta.url));
const ROOT = join(HERE, '..', '..');
const SRC = join(ROOT, 'src', 'content', 'commands');
const OUT = join(HERE, '..', 'commands.json');

const STAGE_ORDER = STAGES.map((s) => s.id);

async function walk(dir: string): Promise<string[]> {
	const entries = await readdir(dir, { withFileTypes: true });
	const out: string[] = [];
	for (const entry of entries) {
		const p = join(dir, entry.name);
		if (entry.isDirectory()) out.push(...(await walk(p)));
		else if (entry.name.endsWith('.md')) out.push(p);
	}
	return out;
}

function groupFor(stageId: Stage, cmdId: string): string | null {
	const groups = STAGE_GROUPS[stageId];
	if (!groups) return null;
	for (const g of groups) {
		if (g.patterns.some((p) => matchPattern(cmdId, p))) return g.id;
	}
	return null;
}

const paths = await walk(SRC);
const byStage = new Map<Stage, any[]>();

for (const path of paths) {
	const raw = await readFile(path, 'utf8');
	const { data } = matter(raw);
	const stageId = data.category as Stage;
	if (!STAGE_ORDER.includes(stageId)) continue;
	const slug = basename(path, '.md');
	const id = `${stageId}/${slug}`;
	const cmd = {
		id,
		slug,
		title: data.title,
		summary: data.summary ?? '',
		order: typeof data.order === 'number' ? data.order : 999,
		groupId: groupFor(stageId, id),
		danger: data.danger ?? 'low',
		deckOnly: data.deck_only ?? false,
		run: (data.commands ?? [])
			.map((c: { run?: string }) => c.run)
			.filter(Boolean),
	};
	if (!byStage.has(stageId)) byStage.set(stageId, []);
	byStage.get(stageId)!.push(cmd);
}

const stages = STAGE_ORDER.filter((id) => byStage.has(id)).map((id) => {
	const meta = STAGES.find((s) => s.id === id)!;
	const items = byStage.get(id)!.sort((a, b) => a.order - b.order);
	return {
		id,
		num: meta.num,
		short: meta.short,
		tagline: meta.tagline,
		sigil: STAGE_SIGILS[id].char,
		groups: STAGE_GROUPS[id]?.map((g) => ({ id: g.id, name: g.name })) ?? [],
		items,
	};
});

const allCmdIds = stages.flatMap((s) => s.items.map((c) => c.id));
const presets = PRESETS.map((p) => {
	const matchIds = allCmdIds.filter((id) =>
		p.patterns.some((pat) => matchPattern(id, pat)),
	);
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
