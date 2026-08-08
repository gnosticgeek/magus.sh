<script lang="ts">
	import { onMount } from 'svelte';
	import StepRow from '$lib/components/StepRow.svelte';
	import { createClient, type CommandName, type MagusClient, type MagusResult } from '$lib/magus';
	import { THEMES, applyTheme, loadTheme, type ThemeId } from '$lib/theme';

	let client = $state<MagusClient | null>(null);
	let result = $state<MagusResult | null>(null);
	let busy = $state(false);
	let failure = $state<string | null>(null);
	let selected = $state(0);
	let query = $state('');
	let theme = $state<ThemeId>('tokyo-night');

	onMount(async () => {
		theme = loadTheme();
		client = await createClient();
		await run('doctor');
	});

	/**
	 * Every action goes through here, and `command` is a member of a closed
	 * union — there is no path from this view to a shell string.
	 */
	async function run(command: CommandName) {
		if (!client || busy) return;
		busy = true;
		failure = null;
		try {
			result = await client.run(command);
		} catch (e) {
			// A thrown error is the contract failing, not a step failing — the
			// two deserve different words.
			failure = e instanceof Error ? e.message : String(e);
		} finally {
			busy = false;
		}
	}

	function setTheme(id: ThemeId) {
		theme = id;
		applyTheme(id);
	}

	const steps = $derived(result?.steps ?? []);
	const visible = $derived(
		query.trim()
			? steps.filter((s) =>
					`${s.id} ${s.describe} ${s.why ?? ''}`.toLowerCase().includes(query.trim().toLowerCase()),
				)
			: steps,
	);
	const attention = $derived(result?.summary.needs_attention ?? 0);

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') {
			selected = Math.min(visible.length - 1, selected + 1);
			e.preventDefault();
		} else if (e.key === 'ArrowUp') {
			selected = Math.max(0, selected - 1);
			e.preventDefault();
		}
	}
</script>

<svelte:window onkeydown={onKeydown} />

<main>
	<div class="panel">
		<header>
			<div class="title">
				<h1>magus</h1>
				{#if result}
					<span class="device">
						{result.device.kind}
						{#if !result.device.confident}
							<!-- An unverified guess must not be presented as a fact (§10). -->
							<span class="hedge" title="Detection is a heuristic on this hardware">guess</span>
						{/if}
					</span>
				{/if}
			</div>

			<div class="themes" role="group" aria-label="Theme">
				{#each THEMES as t (t.id)}
					<button
						class="swatch"
						class:active={theme === t.id}
						title={t.name}
						aria-label={t.name}
						aria-pressed={theme === t.id}
						onclick={() => setTheme(t.id)}
					>
						{#each t.chips as c (c)}<i style="background:{c}"></i>{/each}
					</button>
				{/each}
			</div>
		</header>

		<input
			class="search"
			bind:value={query}
			placeholder="Search your machine…"
			spellcheck="false"
			autocomplete="off"
		/>

		<div class="list" role="listbox" aria-label="Steps" tabindex="-1">
			{#if busy && !result}
				<p class="empty">checking…</p>
			{:else if failure}
				<p class="empty bad selectable">{failure}</p>
			{:else if result?.error}
				<p class="empty selectable">{result.error}</p>
			{:else if visible.length === 0}
				<p class="empty">nothing matches “{query}”</p>
			{:else}
				{#each visible as step, i (step.id)}
					<StepRow {step} selected={i === selected} onselect={() => (selected = i)} />
				{/each}
			{/if}
		</div>

		<footer>
			<span class="counts">
				{#if result}
					{#if attention > 0}
						<strong class="warn">{attention}</strong> need attention
					{:else}
						<strong class="ok">all correct</strong>
					{/if}
					<span class="dim">· {result.summary.total} steps</span>
				{/if}
			</span>

			<span class="actions">
				<button onclick={() => run('doctor')} disabled={busy}>Doctor</button>
				<button class="primary" onclick={() => run('reconcile')} disabled={busy || attention === 0}>
					{busy ? 'Working…' : 'Repair'}
				</button>
			</span>
		</footer>

		{#if client?.kind === 'mock'}
			<!-- Never let a demo be mistaken for a real machine. -->
			<p class="banner">fixtures — no magus binary attached</p>
		{/if}
	</div>
</main>

<style>
	main {
		height: 100vh;
		display: grid;
		place-items: center;
		padding: 24px;
	}

	.panel {
		width: 100%;
		max-width: 34rem;
		display: flex;
		flex-direction: column;
		background: var(--panel);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: 20px;
		box-shadow: 0 19px 38px rgba(0, 0, 0, 0.3), 0 15px 12px rgba(0, 0, 0, 0.22);
		max-height: calc(100vh - 48px);
	}

	header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 14px;
	}
	.title { display: flex; align-items: baseline; gap: 10px; min-width: 0; }
	h1 { margin: 0; font-size: 1.05rem; font-weight: 600; letter-spacing: -0.01em; }
	.device { font-family: var(--mono); font-size: 0.72rem; color: var(--fg-dim); }
	.hedge {
		font-size: 0.62rem;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		padding: 0.1rem 0.3rem;
		border-radius: 4px;
		color: var(--warn);
		background: color-mix(in srgb, var(--warn) 16%, transparent);
	}

	.themes { display: flex; gap: 4px; }
	.swatch {
		display: flex;
		gap: 1px;
		padding: 3px;
		border: 1px solid transparent;
		border-radius: 6px;
		background: none;
		cursor: pointer;
	}
	.swatch i { width: 6px; height: 14px; border-radius: 1px; display: block; }
	.swatch:hover { border-color: var(--border); }
	.swatch.active { border-color: var(--accent); }

	.search {
		width: 100%;
		background: var(--panel-lift);
		border: 0;
		border-radius: var(--radius);
		padding: 10px 12px;
		color: var(--fg);
		font: inherit;
		caret-color: var(--accent);
		outline: none;
		margin-bottom: 6px;
		user-select: text;
	}
	.search::placeholder { color: var(--fg-dim); opacity: 0.5; }

	.list {
		flex: 1;
		min-height: 8rem;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 2px;
		scrollbar-width: none;
	}
	.list::-webkit-scrollbar { width: 0; }

	.empty { padding: 24px 10px; text-align: center; opacity: 0.5; font-size: 0.9rem; }
	.empty.bad { color: var(--bad); opacity: 0.9; white-space: pre-wrap; }

	footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		padding-top: 12px;
		margin-top: 10px;
		border-top: 1px solid color-mix(in srgb, var(--fg) 9%, transparent);
	}
	.counts { font-size: 0.82rem; color: var(--fg-dim); }
	.counts .warn { color: var(--warn); }
	.counts .ok { color: var(--ok); }
	.counts .dim { opacity: 0.6; }

	.actions { display: flex; gap: 8px; }
	.actions button {
		padding: 6px 14px;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: transparent;
		cursor: pointer;
		font-size: 0.85rem;
		transition: border-color 0.12s, background 0.12s;
	}
	.actions button:hover:not(:disabled) { border-color: var(--accent); }
	.actions button:disabled { opacity: 0.4; cursor: default; }
	.actions .primary {
		border-color: var(--accent);
		background: color-mix(in srgb, var(--accent) 22%, transparent);
	}

	.banner {
		margin: 10px 0 0;
		text-align: center;
		font-family: var(--mono);
		font-size: 0.68rem;
		color: var(--warn);
		opacity: 0.8;
	}
</style>
