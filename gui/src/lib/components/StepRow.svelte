<script lang="ts">
	import type { StepResult } from '$lib/magus';

	interface Props {
		step: StepResult;
		selected?: boolean;
		onselect?: () => void;
	}

	let { step, selected = false, onselect }: Props = $props();

	/**
	 * One state word per row. `needs_attention` is what colours it — a step that
	 * was missing and has been repaired reports state_after: "ok" and must read
	 * as done, not as a problem.
	 */
	const label = $derived(step.error ? 'failed' : (step.state_after ?? step.state));
	const tone = $derived(
		step.error ? 'bad' : step.needs_attention ? 'warn' : step.changed ? 'accent' : step.state === 'n/a' ? 'dim' : 'ok',
	);
</script>

<div
	class="row"
	class:selected
	role="option"
	tabindex="0"
	aria-selected={selected}
	onclick={onselect}
	onkeydown={(e) => {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			onselect?.();
		}
	}}
>
	<span class="dot" data-tone={tone} aria-hidden="true"></span>

	<span class="text">
		<span class="id">{step.id}</span>
		<span class="sub">{step.why || step.describe}</span>
	</span>

	<span class="state" data-tone={tone}>{label}</span>
</div>

<style>
	/* Proportions follow walker's default theme: 10px radius, 10px padding, a
	   25%-alpha accent wash for selection rather than a solid fill, subtext at
	   12px/50%. It reads calm because the selection tints instead of inverting. */
	.row {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px;
		border-radius: var(--radius);
		cursor: pointer;
		transition: background 0.1s;
	}
	.row:hover { background: color-mix(in srgb, var(--fg) 5%, transparent); }
	.row.selected { background: color-mix(in srgb, var(--accent) 25%, transparent); }

	.dot {
		width: 8px;
		height: 8px;
		flex: none;
		border-radius: 50%;
		background: var(--fg-dim);
	}
	.dot[data-tone='ok'] { background: var(--ok); }
	.dot[data-tone='warn'] { background: var(--warn); }
	.dot[data-tone='bad'] { background: var(--bad); }
	.dot[data-tone='accent'] { background: var(--accent); }
	.dot[data-tone='dim'] { background: var(--fg-dim); opacity: 0.4; }

	.text { min-width: 0; flex: 1; display: flex; flex-direction: column; }

	.id {
		font-family: var(--mono);
		font-size: 0.85rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.sub {
		font-size: 12px;
		opacity: 0.5;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.state {
		flex: none;
		font-family: var(--mono);
		font-size: 0.68rem;
		letter-spacing: 0.06em;
		padding: 0.14rem 0.45rem;
		border-radius: 5px;
		color: var(--fg-dim);
		background: color-mix(in srgb, var(--fg) 8%, transparent);
	}
	.state[data-tone='ok'] { color: var(--ok); background: color-mix(in srgb, var(--ok) 14%, transparent); }
	.state[data-tone='warn'] { color: var(--warn); background: color-mix(in srgb, var(--warn) 16%, transparent); }
	.state[data-tone='bad'] { color: var(--bad); background: color-mix(in srgb, var(--bad) 16%, transparent); }
	.state[data-tone='accent'] { color: var(--accent); background: color-mix(in srgb, var(--accent) 16%, transparent); }
</style>
