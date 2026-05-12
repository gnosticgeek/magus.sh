import type { PeekDetail } from '../types';

const HINT_KEY = 'magus:card-hint-dismissed';
const PEEK_EVENT = 'command-peek:open';
const LONG_PRESS_MS = 500;
const LONG_PRESS_MOVE_THRESHOLD_PX = 8;
const VIBRATE_TAP_MS = 8;
const VIBRATE_LONG_PRESS_MS = 16;

const readPayload = (card: Element): PeekDetail | null => {
	const node = card.querySelector<HTMLScriptElement>('script[data-payload="peek"]');
	if (!node?.textContent) return null;
	try {
		return JSON.parse(node.textContent) as PeekDetail;
	} catch {
		return null;
	}
};

const dispatchPeek = (card: Element) => {
	const detail = readPayload(card);
	if (!detail) return;
	document.dispatchEvent(new CustomEvent(PEEK_EVENT, { detail }));
};

const dismissHint = () => {
	document.documentElement.classList.remove('show-card-hint');
	try {
		localStorage.setItem(HINT_KEY, '1');
	} catch {
		/* private mode or storage disabled — non-fatal */
	}
};

const wireDelegatedClicks = (root: ParentNode) => {
	root.addEventListener('click', (e) => {
		const target = e.target as HTMLElement;

		const undo = target.closest<HTMLButtonElement>('.undo-toggle');
		if (undo) {
			e.stopPropagation();
			const article = undo.closest('article');
			const content = article?.querySelector<HTMLElement>('.undo-content');
			if (!content) return;
			const open = !content.classList.contains('hidden');
			content.classList.toggle('hidden', open);
			undo.setAttribute('aria-expanded', String(!open));
			return;
		}

		const viewBtn = target.closest<HTMLElement>('.view-cmd-glyph, .view-cmd-preview');
		if (viewBtn) {
			e.stopPropagation();
			const card = viewBtn.closest('article');
			if (card) dispatchPeek(card);
			return;
		}

		const tapTarget = target.closest<HTMLElement>('.card .tap-target');
		if (tapTarget) {
			if (target.closest('input, a, button')) return;
			const checkbox = tapTarget.querySelector<HTMLInputElement>('input.card-checkbox');
			if (!checkbox) return;
			checkbox.checked = !checkbox.checked;
			checkbox.dispatchEvent(new Event('change', { bubbles: true }));
			if ('vibrate' in navigator) navigator.vibrate?.(VIBRATE_TAP_MS);
			dismissHint();
		}
	});
};

/* Long-press a card to peek without ticking the checkbox. Per-card state, so
   we attach listeners per element rather than delegating. */
const wireLongPress = (card: HTMLElement) => {
	let timer: number | undefined;
	let startX = 0;
	let startY = 0;
	let armed = false;
	let fired = false;

	const cancel = () => {
		if (timer) window.clearTimeout(timer);
		timer = undefined;
		armed = false;
	};

	card.addEventListener('pointerdown', (e) => {
		if (e.pointerType !== 'touch') return;
		const target = e.target as HTMLElement;
		if (target.closest('input, a, button, .view-cmd-preview, .view-cmd-glyph')) return;
		armed = true;
		fired = false;
		startX = e.clientX;
		startY = e.clientY;
		timer = window.setTimeout(() => {
			if (!armed) return;
			dispatchPeek(card);
			if ('vibrate' in navigator) navigator.vibrate?.(VIBRATE_LONG_PRESS_MS);
			fired = true;
		}, LONG_PRESS_MS);
	});

	card.addEventListener('pointermove', (e) => {
		if (!armed) return;
		if (Math.hypot(e.clientX - startX, e.clientY - startY) > LONG_PRESS_MOVE_THRESHOLD_PX) {
			cancel();
		}
	});

	const swallow = (e: Event) => {
		if (fired) {
			e.preventDefault();
			e.stopPropagation();
			fired = false;
		}
		cancel();
	};

	card.addEventListener('pointerup', swallow);
	card.addEventListener('pointercancel', cancel);
	card.addEventListener(
		'click',
		(e) => {
			if (fired) {
				e.preventDefault();
				e.stopPropagation();
				fired = false;
			}
		},
		true,
	);
};

const restoreHintVisibility = () => {
	try {
		if (!localStorage.getItem(HINT_KEY)) {
			document.documentElement.classList.add('show-card-hint');
		}
	} catch {
		/* private mode or storage disabled — non-fatal */
	}
};

export const initCommandCards = (root: ParentNode = document) => {
	wireDelegatedClicks(root);
	root.querySelectorAll<HTMLElement>('.card').forEach(wireLongPress);
	restoreHintVisibility();
};
