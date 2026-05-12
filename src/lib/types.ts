export interface PeekDetail {
	title: string;
	summary: string;
	cmd: string;
	undo?: string;
	danger?: 'low' | 'medium' | 'high';
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
