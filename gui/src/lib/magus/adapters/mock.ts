import type {
	CommandName,
	CommandOptions,
	MagusClient,
	MagusResult,
	StepResult,
} from '../types';
import { SCHEMA } from '../types';

/**
 * A fixture-backed client.
 *
 * This is not only a test double — it is what makes the UI developable and
 * reviewable in a plain browser, with no Rust toolchain, no Tauri window and no
 * Steam hardware. That is the same property the Go side has (§9: every action
 * expressible without the wizard), applied to the front end.
 *
 * The fixtures are shaped to be *awkward* rather than flattering: a drifted
 * step, an errored step, a skipped step with a reason, and an unconfident
 * device. A UI that looks right against these looks right against a real
 * machine having a bad day.
 */

const step = (s: Partial<StepResult> & Pick<StepResult, 'id'>): StepResult => ({
	describe: s.describe ?? s.id,
	state: 'ok',
	changed: false,
	needs_attention: false,
	...s,
});

/** A Deck mid-life: mostly converged, one thing broken by an OS update. */
const DECK_STEPS: StepResult[] = [
	step({ id: 'flathub', describe: 'add the Flathub remote for this user' }),
	step({
		id: 'terminal:kitty',
		describe: 'install kitty into ~/.local/kitty.app',
		state: 'drifted',
		needs_attention: true,
	}),
	step({ id: 'browser:firefox', describe: 'install Firefox (org.mozilla.firefox) as a user flatpak' }),
	step({ id: 'bundle:essentials/org.kde.ark', describe: 'install Ark (archives)' }),
	step({ id: 'bundle:essentials/org.kde.kate', describe: 'install Kate (text editor)' }),
	step({ id: 'bundle:essentials/io.mpv.Mpv', describe: 'install mpv (media player)' }),
	step({ id: 'bundle:gaming/net.davidotek.pupgui2', describe: 'install ProtonUp-Qt' }),
	step({
		id: 'bundle:gaming/com.heroicgameslauncher.hgl',
		describe: 'install Heroic Games Launcher',
		state: 'missing',
		needs_attention: true,
		error: 'flatpak: exit status 1\nerror: Unable to load summary from remote flathub',
	}),
	step({ id: 'optimise:double-click', describe: 'double-click to open, not single' }),
	step({ id: 'optimise:baloo-off', describe: "disable Baloo, KDE's file indexer" }),
	step({ id: 'optimise:cursor-size', describe: 'cursor size 32px' }),
	step({
		id: 'optimise:proton-ge',
		describe: "install the latest GE-Proton into Steam's compatibilitytools.d",
	}),
];

/** A Machine additionally carries the steps that are recorded but unbuilt. */
const MACHINE_ONLY: StepResult[] = [
	step({
		id: 'optimise:hdmi-full-range',
		describe: 'HDMI colour range → full RGB — not built yet',
		state: 'n/a',
		why: 'not built yet: needs a verified DRM property on real hardware — §10',
	}),
	step({
		id: 'optimise:cec',
		describe: 'HDMI-CEC — not built yet',
		state: 'n/a',
		why: 'not built yet: SteamOS added CEC controls; the userland hook is unverified — §10',
	}),
];

function summarise(steps: StepResult[]) {
	return {
		total: steps.length,
		changed: steps.filter((s) => s.changed).length,
		ok: steps.filter((s) => s.state === 'ok' && !s.changed).length,
		not_applicable: steps.filter((s) => s.state === 'n/a').length,
		failed: steps.filter((s) => s.error).length,
		needs_attention: steps.filter((s) => s.needs_attention).length,
	};
}

export interface MockOptions {
	/** Which machine to pretend to be. */
	device?: 'steam-deck' | 'steam-machine';
	/** Simulated latency, so loading states are visible during development. */
	delayMs?: number;
	/** Pretend no manifest has been written yet. */
	unconfigured?: boolean;
}

export class MockClient implements MagusClient {
	readonly kind = 'mock' as const;

	constructor(private readonly opts: MockOptions = {}) {}

	async available(): Promise<boolean> {
		return true;
	}

	async run(command: CommandName, options: CommandOptions = {}): Promise<MagusResult> {
		await new Promise((r) => setTimeout(r, this.opts.delayMs ?? 250));

		const isMachine = this.opts.device === 'steam-machine';
		const steps =
			command === 'version' ? [] : [...DECK_STEPS, ...(isMachine ? MACHINE_ONLY : [])];

		// Reconciling repairs what it can; the flatpak failure survives, because
		// a UI that never sees a partial success will not handle one.
		const applied =
			command === 'reconcile' || command === 'run-defaults'
				? steps.map((s) =>
						s.needs_attention && !s.error
							? { ...s, state_after: 'ok' as const, changed: true, needs_attention: false }
							: s,
					)
				: steps;

		const base: MagusResult = {
			schema: SCHEMA,
			command,
			magus: {
				version: 'v0.3.0',
				manifest_schema: '0.3.0',
				dry_run: options.dryRun ?? command === 'doctor',
			},
			device: {
				kind: isMachine ? 'steam-machine' : 'steam-deck',
				// A Steam Machine is only ever a heuristic match today (§10).
				confident: !isMachine,
				vendor: 'Valve',
				product: isMachine ? '' : 'Jupiter',
				os_id: 'steamos',
				steamos: true,
			},
			tooling: [
				{ name: 'flatpak', present: true },
				{ name: 'curl', present: true },
			],
			manifest: {
				path: '/home/deck/.config/magus/manifest.toml',
				present: !this.opts.unconfigured,
				valid: !this.opts.unconfigured,
				schema: '0.3.0',
				device: isMachine ? 'steam-machine' : 'steam-deck',
			},
			steps: applied,
			summary: summarise(applied),
			exit_code: 0,
		};

		if (this.opts.unconfigured) {
			return {
				...base,
				steps: [],
				summary: summarise([]),
				error: 'no manifest — this machine has not been set up',
				exit_code: 1,
			};
		}
		base.exit_code = base.summary.needs_attention > 0 ? 1 : 0;
		return base;
	}
}
