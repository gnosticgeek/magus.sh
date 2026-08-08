import { MockClient } from './adapters/mock';
import { TauriClient } from './adapters/tauri';
import type { MagusClient } from './types';

export * from './types';
export { MockClient } from './adapters/mock';
export { TauriClient } from './adapters/tauri';

/**
 * Pick a client for the environment.
 *
 * Inside Tauri, talk to the real binary. In a browser — `npm run dev`, a review
 * on someone else's laptop, a screenshot for a bug report — fall back to
 * fixtures so the whole UI is still exercisable.
 */
export async function createClient(): Promise<MagusClient> {
	const inTauri = typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window;
	if (!inTauri) return new MockClient();

	const tauri = new TauriClient();
	// Present but no magus binary on PATH is a real state, and one the UI has
	// to explain rather than hang on.
	return (await tauri.available()) ? tauri : new MockClient();
}
