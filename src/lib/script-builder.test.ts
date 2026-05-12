import { describe, it, expect } from 'vitest';
import {
	buildHeader,
	itemBlock,
	buildFullScript,
	buildStageScript,
	groupByStage,
} from './script-builder';
import type { ScriptItem } from './types';

const item = (
	overrides: Partial<ScriptItem> & { title: string; cmd: string; category: string },
): ScriptItem => ({
	deckOnly: false,
	requiresDecky: false,
	...overrides,
});

describe('buildHeader', () => {
	it('singular grammar for one item', () => {
		const header = buildHeader([item({ title: 'a', cmd: 'echo a', category: 'setup' })]);
		expect(header).toContain('# 1 command selected');
	});

	it('plural grammar for many items', () => {
		const header = buildHeader([
			item({ title: 'a', cmd: 'echo a', category: 'setup' }),
			item({ title: 'b', cmd: 'echo b', category: 'install' }),
		]);
		expect(header).toContain('# 2 commands selected');
	});

	it('lists stage labels deduped in first-appearance order', () => {
		const header = buildHeader([
			item({ title: 'a', cmd: 'a', category: 'install' }),
			item({ title: 'b', cmd: 'b', category: 'setup' }),
			item({ title: 'c', cmd: 'c', category: 'install' }),
		]);
		expect(header).toContain('# Stages: Install, Setup');
	});

	it('falls back to raw category for unknown stage', () => {
		const header = buildHeader([item({ title: 'a', cmd: 'a', category: 'bogus' })]);
		expect(header).toContain('# Stages: bogus');
	});
});

describe('itemBlock', () => {
	it('omits risk marker when danger is absent', () => {
		expect(itemBlock(item({ title: 'Tidy', cmd: 'rm -rf /tmp/x', category: 'setup' })))
			.toBe('# --- Tidy ---\nrm -rf /tmp/x');
	});

	it('includes risk marker when danger present', () => {
		expect(itemBlock(item({ title: 'Boom', cmd: 'wipe', category: 'setup', danger: 'high' })))
			.toBe('# --- Boom (high risk) ---\nwipe');
	});

	it('trims surrounding whitespace from title and cmd', () => {
		expect(itemBlock(item({ title: '  Pad  ', cmd: '\n echo hi \n', category: 'setup' })))
			.toBe('# --- Pad ---\necho hi');
	});
});

describe('buildFullScript', () => {
	it('returns empty string for no items', () => {
		expect(buildFullScript([])).toBe('');
	});

	it('inserts stage markers and preserves order', () => {
		const out = buildFullScript([
			item({ title: 'a', cmd: 'echo a', category: 'setup' }),
			item({ title: 'b', cmd: 'echo b', category: 'install' }),
			item({ title: 'c', cmd: 'echo c', category: 'install' }),
		]);
		expect(out).toMatchInlineSnapshot(`
			"#!/usr/bin/env bash
			# magus.sh review script
			# 3 commands selected
			# Stages: Setup, Install
			set -e

			# === Setup ===

			# --- a ---
			echo a

			# === Install ===

			# --- b ---
			echo b

			# --- c ---
			echo c
			"
		`);
	});
});

describe('buildStageScript', () => {
	it('emits per-stage header and blank-separated items', () => {
		const out = buildStageScript('Setup', [
			item({ title: 'a', cmd: 'echo a', category: 'setup' }),
			item({ title: 'b', cmd: 'echo b', category: 'setup' }),
		]);
		expect(out).toMatchInlineSnapshot(`
			"#!/usr/bin/env bash
			# magus.sh — Setup
			set -e

			# --- a ---
			echo a

			# --- b ---
			echo b
			"
		`);
	});
});

describe('groupByStage', () => {
	it('groups items and preserves first-appearance order', () => {
		const a = item({ title: 'a', cmd: 'a', category: 'install' });
		const b = item({ title: 'b', cmd: 'b', category: 'setup' });
		const c = item({ title: 'c', cmd: 'c', category: 'install' });
		expect(groupByStage([a, b, c])).toEqual([
			{ stage: 'Install', items: [a, c] },
			{ stage: 'Setup', items: [b] },
		]);
	});
});
