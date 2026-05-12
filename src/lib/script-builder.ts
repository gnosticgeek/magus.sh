import { STAGE_LABELS } from './stages';
import type { ScriptItem } from './types';

const stageLabel = (item: ScriptItem): string =>
	STAGE_LABELS[item.category as keyof typeof STAGE_LABELS] ?? item.category;

export const buildHeader = (items: ScriptItem[]): string => {
	const selectedStages = Array.from(new Set(items.map(stageLabel))).join(', ');
	return [
		'#!/usr/bin/env bash',
		'# magus.sh review script',
		`# ${items.length} ${items.length === 1 ? 'command' : 'commands'} selected`,
		`# Stages: ${selectedStages}`,
		'set -e',
	].join('\n');
};

export const itemBlock = (item: ScriptItem): string => {
	const head = `# --- ${item.title.trim()}${item.danger ? ` (${item.danger} risk)` : ''} ---`;
	return `${head}\n${item.cmd.trim()}`;
};

export const buildFullScript = (items: ScriptItem[]): string => {
	if (items.length === 0) return '';
	const header = buildHeader(items);
	const sections: string[] = [];
	let activeStage = '';
	for (const item of items) {
		const stage = stageLabel(item);
		if (stage !== activeStage) {
			activeStage = stage;
			sections.push(`# === ${stage} ===`);
		}
		sections.push(itemBlock(item));
	}
	return `${header}\n\n${sections.join('\n\n')}\n`;
};

export const buildStageScript = (stage: string, items: ScriptItem[]): string => {
	const lines = [
		'#!/usr/bin/env bash',
		`# magus.sh — ${stage}`,
		'set -e',
		'',
	];
	for (const item of items) {
		lines.push(itemBlock(item));
		lines.push('');
	}
	return lines.join('\n');
};

export interface StageGroup {
	stage: string;
	items: ScriptItem[];
}

export const groupByStage = (items: ScriptItem[]): StageGroup[] => {
	const order: string[] = [];
	const map = new Map<string, ScriptItem[]>();
	for (const item of items) {
		const stage = stageLabel(item);
		if (!map.has(stage)) {
			map.set(stage, []);
			order.push(stage);
		}
		map.get(stage)!.push(item);
	}
	return order.map((stage) => ({ stage, items: map.get(stage)! }));
};
