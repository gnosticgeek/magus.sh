import { readdir, readFile } from 'node:fs/promises';
import path from 'node:path';
import matter from 'gray-matter';

import { STAGE_IDS, type Stage } from '../src/lib/stages.ts';

export type CommandEntry = {
	run: string;
	description?: string;
};

export type CommandDoc = {
	id: string;
	title: string;
	category: Stage;
	order: number;
	summary: string;
	commands: CommandEntry[];
	undo?: string;
	group?: string;
	upstream?: { name: string; url: string };
};

const COMMANDS_DIR = path.resolve(
	import.meta.dirname,
	'..',
	'src',
	'content',
	'commands',
);

export async function loadCommands(): Promise<CommandDoc[]> {
	const entries = await readdir(COMMANDS_DIR, {
		recursive: true,
		withFileTypes: true,
	});

	const docs: CommandDoc[] = [];
	for (const entry of entries) {
		if (!entry.isFile() || !/\.mdx?$/.test(entry.name)) continue;

		const filePath = path.join(entry.parentPath, entry.name);
		const raw = await readFile(filePath, 'utf8');
		const { data } = matter(raw);

		if (!isStageId(data.category)) {
			throw new Error(
				`Invalid or missing category "${data.category}" in ${filePath}`,
			);
		}

		docs.push({
			id: path.basename(entry.name, path.extname(entry.name)),
			title: String(data.title),
			category: data.category,
			order: typeof data.order === 'number' ? data.order : 100,
			summary: String(data.summary ?? ''),
			commands: Array.isArray(data.commands) ? data.commands : [],
			undo: data.undo,
			group: data.group,
			upstream: data.upstream,
		});
	}

	return docs;
}

function isStageId(v: unknown): v is Stage {
	return (
		typeof v === 'string' && (STAGE_IDS as readonly string[]).includes(v)
	);
}
