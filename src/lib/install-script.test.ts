import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { GET as install } from '../pages/install';
import { GET as run } from '../pages/run';

describe('installer route compatibility', () => {
	it('keeps both public URLs plain text and identical to the canonical installer', async () => {
		const source = readFileSync(new URL('../../magus/scripts/install.sh', import.meta.url), 'utf8');
		for (const handler of [install, run]) {
			const response = await handler({} as Parameters<typeof handler>[0]);
			expect(response.headers.get('Content-Type')).toBe('text/plain; charset=utf-8');
			expect(await response.text()).toBe(source);
		}
	});
});
