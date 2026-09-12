import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';
import { RELEASE_VERSION } from './release';

describe('release version', () => {
	it('matches the package version advertised by the site', () => {
		const pkg = JSON.parse(readFileSync(resolve(process.cwd(), 'package.json'), 'utf8')) as { version: string };
		expect(RELEASE_VERSION).toBe(`v${pkg.version}`);
	});
});
