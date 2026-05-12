import { lockBodyScroll } from '../body-scroll-lock';

const CLOSE_ANIM_FALLBACK_MS = 320;

export interface SheetDialog {
	/** Show the dialog and lock body scroll. Pre-open class management is the caller's job. */
	open(): void;
	/** Run the close animation, close the dialog, and unlock body scroll. */
	close(): void;
}

export interface SheetDialogOptions {
	/** Selector inside the dialog for buttons that trigger close. */
	closeSelector?: string;
	/** Called synchronously when close starts, before the animation. */
	beforeClose?: () => void;
	/** Called after the dialog has fully closed and scroll is unlocked. */
	onClosed?: () => void;
}

export function createSheetDialog(
	dialog: HTMLDialogElement,
	{ closeSelector = '[data-action="close"]', beforeClose, onClosed }: SheetDialogOptions = {},
): SheetDialog {
	let unlockScroll: (() => void) | undefined;

	const open = () => {
		unlockScroll = lockBodyScroll();
	};

	const close = () => {
		beforeClose?.();
		dialog.classList.add('closing');
		const onEnd = () => {
			dialog.classList.remove('closing');
			dialog.close();
			dialog.removeEventListener('animationend', onEnd);
			unlockScroll?.();
			unlockScroll = undefined;
			onClosed?.();
		};
		dialog.addEventListener('animationend', onEnd);
		setTimeout(() => {
			if (dialog.classList.contains('closing')) onEnd();
		}, CLOSE_ANIM_FALLBACK_MS);
	};

	dialog.querySelectorAll<HTMLElement>(closeSelector).forEach((btn) => {
		btn.addEventListener('click', close);
	});

	dialog.addEventListener('click', (e) => {
		if (e.target === dialog) close();
	});

	dialog.addEventListener('cancel', (e) => {
		e.preventDefault();
		close();
	});

	return { open, close };
}
