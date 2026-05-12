import { isStage } from './stages';
import type { CardData, Danger } from './types';

const DANGERS: readonly Danger[] = ['low', 'medium', 'high'];

const isDanger = (s: string | undefined): s is Danger =>
	s !== undefined && (DANGERS as readonly string[]).includes(s);

export type SerializedCardData = Record<string, string>;

export const serializeCardData = (data: CardData): SerializedCardData => {
	const attrs: SerializedCardData = {
		'data-id': data.id,
		'data-section': data.section,
		'data-category': data.category,
		'data-cmd': data.cmd,
		'data-title': data.title,
		'data-deck-only': data.deckOnly ? 'true' : 'false',
		'data-requires-decky': data.requiresDecky ? 'true' : 'false',
	};
	if (data.danger) attrs['data-danger'] = data.danger;
	return attrs;
};

export const parseCardData = (el: HTMLElement): CardData | null => {
	const id = el.dataset.id;
	if (!id) return null;
	const category = el.dataset.category;
	if (!category || !isStage(category)) return null;
	const dangerRaw = el.dataset.danger;
	return {
		id,
		section: el.dataset.section ?? '',
		category,
		cmd: el.dataset.cmd ?? '',
		title: el.dataset.title ?? id,
		danger: isDanger(dangerRaw) ? dangerRaw : undefined,
		deckOnly: el.dataset.deckOnly === 'true',
		requiresDecky: el.dataset.requiresDecky === 'true',
	};
};
