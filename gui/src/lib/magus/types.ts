/**
 * The magus JSON contract, mirrored in TypeScript.
 *
 * This is a hand-written mirror of `magus/json.go`. It is deliberately the only
 * place in the GUI that knows the wire format — every view consumes these types
 * rather than raw JSON, so a contract change breaks the build here instead of
 * silently rendering `undefined` somewhere in the UI.
 *
 * Keep `SCHEMA` in step with `jsonSchema` in json.go.
 */

/** The contract version this front end was written against. */
export const SCHEMA = 1;

/** What a step's Check found. Mirrors State.String() in reconcile.go. */
export type StepState = 'ok' | 'missing' | 'drifted' | 'n/a' | 'unknown';

/** Which machine magus thinks it is on. Mirrors DeviceKind. */
export type DeviceKind = 'steam-deck' | 'steam-machine' | 'steamos' | 'other';

export interface MagusInfo {
	/** The binary's release tag, or "dev" for a build from source. */
	version: string;
	/** The manifest format it reads and writes — not the same as `version`. */
	manifest_schema: string;
	dry_run: boolean;
}

export interface DeviceInfo {
	kind: DeviceKind;
	/**
	 * False when detection fell back to a heuristic — currently any Valve board
	 * that is not a known Deck. Present it as a guess, not a fact.
	 */
	confident: boolean;
	vendor: string;
	product: string;
	os_id: string;
	steamos: boolean;
}

export interface ToolInfo {
	name: string;
	present: boolean;
}

export interface ManifestInfo {
	path: string;
	present: boolean;
	valid: boolean;
	/** Validation problems. User-fixable, unlike a fatal `error`. */
	invalid?: string;
	schema?: string;
	device?: string;
	/** The manifest was written for a different machine than this one. */
	device_mismatch?: boolean;
}

export interface StepResult {
	id: string;
	describe: string;
	state: StepState;
	/** Absent when nothing was applied. */
	state_after?: StepState;
	/** Why a step was skipped, e.g. "not built yet: …". */
	why?: string;
	changed: boolean;
	error?: string;
	/** The one field to colour on: still wrong, or errored. */
	needs_attention: boolean;
}

export interface Summary {
	total: number;
	changed: number;
	ok: number;
	not_applicable: number;
	failed: number;
	needs_attention: number;
}

/** One whole `magus <verb> --json` document. */
export interface MagusResult {
	schema: number;
	command: string;
	magus: MagusInfo;
	device: DeviceInfo;
	tooling?: ToolInfo[];
	manifest: ManifestInfo;
	steps: StepResult[];
	summary: Summary;
	/** Set when the command could not complete at all. */
	error?: string;
	exit_code: number;
}

/**
 * The commands a front end may ask for.
 *
 * A closed set on purpose. EmuDeck's Electron wrapper hands arbitrary bash from
 * its renderer to `exec`, which makes any front-end bug arbitrary command
 * execution. Here the UI picks a name from this union and the Rust side maps it
 * to a fixed argv — there is no path by which a string from the view layer
 * reaches a shell.
 */
export type CommandName = 'doctor' | 'reconcile' | 'run-defaults' | 'uninstall' | 'version';

/** Options a command may carry. Still not a string that reaches a shell. */
export interface CommandOptions {
	/** Report what would change without changing it. */
	dryRun?: boolean;
}

/**
 * What every adapter implements. Views depend on this, never on Tauri, so the
 * whole UI runs in a plain browser against fixtures.
 */
export interface MagusClient {
	/** Human-readable name of the backing implementation, for the status bar. */
	readonly kind: 'tauri' | 'mock';
	/** True when a real magus binary is reachable. */
	available(): Promise<boolean>;
	run(command: CommandName, options?: CommandOptions): Promise<MagusResult>;
}

/**
 * Raised when the binary is present but speaks a contract this build does not
 * understand. Better a clear message than fields quietly reading undefined.
 */
export class SchemaMismatchError extends Error {
	constructor(
		readonly got: number,
		readonly want: number = SCHEMA,
	) {
		super(
			`magus speaks JSON schema ${got}, this app understands ${want}. ` +
				`Update ${got > want ? 'the app' : 'magus'}.`,
		);
		this.name = 'SchemaMismatchError';
	}
}

/** Narrowing guard — a malformed document should fail here, not in a view. */
export function isMagusResult(value: unknown): value is MagusResult {
	if (typeof value !== 'object' || value === null) return false;
	const v = value as Record<string, unknown>;
	return (
		typeof v.schema === 'number' &&
		typeof v.command === 'string' &&
		typeof v.exit_code === 'number' &&
		Array.isArray(v.steps) &&
		typeof v.summary === 'object' &&
		v.summary !== null
	);
}
