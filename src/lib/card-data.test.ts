import { describe, it, expect } from 'vitest';
import { serializeCardData, parseCardData } from './card-data';
import type { CardData } from './types';

const mockEl = (dataset: Record<string, string | undefined>): HTMLElement =>
	({ dataset } as unknown as HTMLElement);

const sample: CardData = {
	id: 'setup/01-foo',
	section: 'setup',
	category: 'setup',
	cmd: 'echo hi',
	title: 'Foo',
	danger: 'high',
	deckOnly: true,
	requiresDecky: false,
};

describe('serializeCardData', () => {
	it('emits every required attribute', () => {
		expect(serializeCardData(sample)).toEqual({
			'data-id': 'setup/01-foo',
			'data-section': 'setup',
			'data-category': 'setup',
			'data-cmd': 'echo hi',
			'data-title': 'Foo',
			'data-deck-only': 'true',
			'data-requires-decky': 'false',
			'data-danger': 'high',
		});
	});

	it('omits data-danger when absent', () => {
		const { danger: _ignored, ...rest } = sample;
		expect(serializeCardData(rest)).not.toHaveProperty('data-danger');
	});

	it('coerces booleans to literal strings', () => {
		const out = serializeCardData({ ...sample, deckOnly: false, requiresDecky: true });
		expect(out['data-deck-only']).toBe('false');
		expect(out['data-requires-decky']).toBe('true');
	});
});

describe('parseCardData', () => {
	it('round-trips through serialize → parse', () => {
		const attrs = serializeCardData(sample);
		const dataset = Object.fromEntries(
			Object.entries(attrs).map(([k, v]) => [
				k.replace(/^data-/, '').replace(/-([a-z])/g, (_, c) => c.toUpperCase()),
				v,
			]),
		);
		expect(parseCardData(mockEl(dataset))).toEqual(sample);
	});

	it('returns null when id is missing', () => {
		expect(parseCardData(mockEl({ category: 'setup' }))).toBeNull();
	});

	it('returns null for unknown category', () => {
		expect(parseCardData(mockEl({ id: 'x', category: 'bogus' }))).toBeNull();
	});

	it('treats string "false" and missing values as false for booleans', () => {
		const result = parseCardData(
			mockEl({ id: 'x', category: 'setup', deckOnly: 'false' }),
		);
		expect(result?.deckOnly).toBe(false);
		expect(result?.requiresDecky).toBe(false);
	});

	it('drops invalid danger values rather than passing them through', () => {
		const result = parseCardData(
			mockEl({ id: 'x', category: 'setup', danger: 'extreme' }),
		);
		expect(result?.danger).toBeUndefined();
	});

	it('falls back title to id when title attribute missing', () => {
		const result = parseCardData(mockEl({ id: 'x', category: 'setup' }));
		expect(result?.title).toBe('x');
	});
});
