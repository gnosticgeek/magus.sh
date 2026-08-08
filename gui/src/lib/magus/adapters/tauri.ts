import type { CommandName, CommandOptions, MagusClient, MagusResult } from '../types';
import { SCHEMA, SchemaMismatchError, isMagusResult } from '../types';

/**
 * The real client: hands a *named* command to the Rust side.
 *
 * Note what is not here — there is no way to pass a command line. The name is
 * a member of a closed union, and Rust maps it to a fixed argv. That is the
 * whole point of the boundary: a bug in a view cannot become a shell.
 */
export class TauriClient implements MagusClient {
	readonly kind = 'tauri' as const;

	private async invoke<T>(cmd: string, args?: Record<string, unknown>): Promise<T> {
		// Imported lazily so a plain browser can load this module without the
		// Tauri API present — the app falls back to the mock in that case.
		const { invoke } = await import('@tauri-apps/api/core');
		return invoke<T>(cmd, args);
	}

	async available(): Promise<boolean> {
		try {
			return await this.invoke<boolean>('magus_available');
		} catch {
			return false;
		}
	}

	async run(command: CommandName, options: CommandOptions = {}): Promise<MagusResult> {
		const raw = await this.invoke<unknown>('magus_run', {
			command,
			dryRun: options.dryRun ?? false,
		});

		if (!isMagusResult(raw)) {
			throw new Error('magus returned something that is not a result document');
		}
		// Fail loudly on a contract mismatch rather than rendering undefined.
		if (raw.schema !== SCHEMA) throw new SchemaMismatchError(raw.schema);
		return raw;
	}
}
