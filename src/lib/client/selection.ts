import type { ScriptItem } from '../types';
import { parseCardData } from '../card-data';

const STORAGE_KEY = 'magus:selected-commands';
const VIBRATE_OPEN_SHEET_MS = 12;

type Card = HTMLInputElement;

const readPersisted = (): Set<string> => {
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (!raw) return new Set();
		const parsed = JSON.parse(raw);
		if (!Array.isArray(parsed)) return new Set();
		return new Set(parsed.filter((id): id is string => typeof id === 'string'));
	} catch {
		return new Set();
	}
};

const persist = (ids: Iterable<string>) => {
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(Array.from(ids)));
	} catch {
		/* private mode or storage disabled — non-fatal */
	}
};

const bucketBy = <K>(items: Card[], key: (c: Card) => K): Map<K, Card[]> => {
	const map = new Map<K, Card[]>();
	for (const c of items) {
		const k = key(c);
		let bucket = map.get(k);
		if (!bucket) {
			bucket = [];
			map.set(k, bucket);
		}
		bucket.push(c);
	}
	return map;
};

export const initSelection = () => {
	const bar = document.getElementById('selection-bar');
	if (!bar) return;
	const countEl = bar.querySelector<HTMLElement>('.count');
	const copyBtn = bar.querySelector<HTMLButtonElement>('[data-action="copy"]');
	const copyLabel = copyBtn?.querySelector<HTMLElement>('.copy-action-label');
	const clearBtn = bar.querySelector<HTMLButtonElement>('[data-action="clear"]');
	const stageCountEls = new Map(
		Array.from(document.querySelectorAll<HTMLElement>('[data-stage-count]')).map(
			(el) => [el.dataset.stageCount ?? '', el],
		),
	);
	const stageTabs = Array.from(
		document.querySelectorAll<HTMLElement>('a[role="tab"][data-stage-tab-id]'),
	);
	if (!countEl || !copyBtn || !copyLabel || !clearBtn) return;

	const cards = Array.from(document.querySelectorAll<Card>('input.card-checkbox'));
	const sectionToggles = Array.from(
		document.querySelectorAll<HTMLButtonElement>('.section-select'),
	);

	const cardsBySection = bucketBy(cards, (c) => c.dataset.section ?? '');
	const cardsByStage = bucketBy(cards, (c) => c.dataset.category ?? '');

	const selected = new Map<string, string>();
	const selectedMeta = new Map<string, ScriptItem>();

	const render = () => {
		const count = selected.size;
		countEl.textContent = String(count);
		bar.classList.toggle('visible', count > 0);
		copyBtn.disabled = count === 0;
		copyLabel.textContent =
			count === 0
				? 'Script'
				: count === 1
					? 'Script · 1'
					: `Script · ${count}`;

		stageTabs.forEach((tab) => {
			const stage = tab.dataset.stageTabId ?? '';
			const bucket = cardsByStage.get(stage) ?? [];
			const total = bucket.length;
			const checked = bucket.filter((c) => c.checked).length;
			const ratio = total === 0 ? 0 : checked / total;
			tab.style.setProperty('--stage-progress', String(ratio));
			const state = checked === 0 ? 'empty' : checked === total ? 'full' : 'some';
			tab.dataset.progressState = state;
			const stageCountEl = stageCountEls.get(stage);
			if (stageCountEl) {
				stageCountEl.textContent = total === 0 ? '' : `${checked}/${total}`;
			}
		});

		sectionToggles.forEach((btn) => {
			const sectionId = btn.dataset.sectionId;
			if (!sectionId) return;
			const inSection = cardsBySection.get(sectionId) ?? [];
			const checked = inSection.filter((c) => c.checked).length;
			const total = inSection.length;
			const state = checked === 0 ? 'none' : checked === total ? 'all' : 'some';
			btn.dataset.state = state;
			const label = btn.querySelector<HTMLElement>('.label');
			if (label) {
				label.textContent = state === 'all' ? 'clear all' : 'select all';
			}
		});
	};

	const setCard = (cb: Card, on: boolean) => {
		cb.checked = on;
		const article = cb.closest('article');
		article?.classList.toggle('selected', on);
		const card = parseCardData(cb);
		if (!card) return;
		if (on) {
			selected.set(card.id, card.cmd);
			selectedMeta.set(card.id, {
				title: card.title,
				cmd: card.cmd,
				category: card.category,
				danger: card.danger,
				deckOnly: card.deckOnly,
				requiresDecky: card.requiresDecky,
			});
		} else {
			selected.delete(card.id);
			selectedMeta.delete(card.id);
		}
	};

	cards.forEach((cb) => {
		cb.addEventListener('change', () => {
			setCard(cb, cb.checked);
			render();
			persist(selected.keys());
		});
	});

	sectionToggles.forEach((btn) => {
		btn.addEventListener('click', () => {
			const sectionId = btn.dataset.sectionId;
			if (!sectionId) return;
			const inSection = cardsBySection.get(sectionId) ?? [];
			const allChecked = inSection.every((c) => c.checked);
			inSection.forEach((c) => setCard(c, !allChecked));
			render();
			persist(selected.keys());
		});
	});

	copyBtn.addEventListener('click', () => {
		if (selected.size === 0) return;
		const items = cards
			.filter((c) => c.checked)
			.map((c) => selectedMeta.get(c.dataset.id ?? ''))
			.filter((item): item is ScriptItem => Boolean(item));
		if ('vibrate' in navigator) navigator.vibrate?.(VIBRATE_OPEN_SHEET_MS);
		document.dispatchEvent(
			new CustomEvent('script-sheet:open', { detail: { items } }),
		);
	});

	clearBtn.addEventListener('click', () => {
		cards.forEach((cb) => setCard(cb, false));
		render();
		persist(selected.keys());
	});

	/* restore prior session selection before first render */
	const persisted = readPersisted();
	if (persisted.size > 0) {
		for (const cb of cards) {
			if (cb.dataset.id && persisted.has(cb.dataset.id)) {
				setCard(cb, true);
			}
		}
	}

	render();
};
