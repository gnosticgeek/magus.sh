import { checkbox, Separator } from '@inquirer/prompts';

import { STAGES } from '../src/lib/stages.ts';
import { loadCommands, type CommandDoc } from './load-commands.ts';

async function main(): Promise<void> {
	const docs = await loadCommands();

	const choices: Array<
		Separator | { name: string; value: string; description?: string }
	> = [];

	for (const stage of STAGES) {
		const inStage = docs
			.filter((d) => d.category === stage.id)
			.sort((a, b) => a.order - b.order);

		if (inStage.length === 0) continue;

		choices.push(new Separator(`── ${stage.num} · ${stage.short} ──`));
		for (const d of inStage) {
			choices.push({
				name: d.title,
				value: d.id,
				description: d.summary,
			});
		}
	}

	const selectedIds = await checkbox(
		{
			message: 'Select commands (space to toggle, enter to confirm)',
			choices,
			pageSize: 20,
			loop: false,
		},
		{ output: process.stderr, clearPromptOnDone: true },
	);

	const byId = new Map(docs.map((d) => [d.id, d]));
	const stageOrder = new Map(STAGES.map((s, i) => [s.id, i] as const));

	const selected = selectedIds
		.map((id) => byId.get(id))
		.filter((d): d is CommandDoc => Boolean(d))
		.sort(
			(a, b) =>
				(stageOrder.get(a.category) ?? 0) -
					(stageOrder.get(b.category) ?? 0) || a.order - b.order,
		);

	if (selected.length === 0) {
		process.stderr.write('Nothing selected.\n');
		return;
	}

	process.stdout.write('#!/usr/bin/env bash\n');
	process.stdout.write('# magus.sh — selected commands\n\n');
	for (const d of selected) {
		process.stdout.write(`# ${d.title}\n`);
		if (d.summary) process.stdout.write(`# ${d.summary}\n`);
		if (d.upstream) process.stdout.write(`# Source: ${d.upstream.name} — ${d.upstream.url}\n`);
		for (const cmd of d.commands) {
			if (cmd.description) process.stdout.write(`# ${cmd.description}\n`);
			process.stdout.write(`${cmd.run}\n`);
		}
		process.stdout.write('\n');
	}
}

main().catch((err: unknown) => {
	if (
		err &&
		typeof err === 'object' &&
		'name' in err &&
		err.name === 'ExitPromptError'
	) {
		process.exit(130);
	}
	console.error(err);
	process.exit(1);
});
