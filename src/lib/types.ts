import type { Stage } from './stages';

export type StageId = Stage;

export type Danger = 'low' | 'medium' | 'high';

export interface StageMeta {
	id: StageId;
	num: string;
	label: string;
	short: string;
	title: string;
	tagline: string;
	stageLabel: string;
	rise: number;
	variant?: 'retro';
}

export interface PeekDetail {
	title: string;
	summary: string;
	cmd: string;
	undo?: string;
	danger?: Danger;
	deckOnly?: boolean;
}

export interface ScriptItem {
	title: string;
	cmd: string;
	category: string;
	danger?: string;
	deckOnly: boolean;
	requiresDecky: boolean;
}

export interface CardData {
	id: string;
	section: string;
	category: StageId;
	cmd: string;
	title: string;
	danger?: Danger;
	deckOnly: boolean;
	requiresDecky: boolean;
}
