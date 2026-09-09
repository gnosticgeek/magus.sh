/** Report success only after the browser confirms the clipboard write. */
export const copyWithFeedback = async (
	text: string,
	button: HTMLElement,
	label: HTMLElement,
	vibratePattern: number | number[] = 10,
	feedbackMs = 1600,
): Promise<void> => {
	if (!text || button.dataset.copyPending === 'true') return;
	button.dataset.copyPending = 'true';
	const original = label.textContent ?? '';
	label.setAttribute('aria-live', 'polite');
	let copied = false;
	try {
		await navigator.clipboard.writeText(text);
		copied = true;
	} catch {
		// Keep the command visible for manual copying when permission is denied.
	}
	label.textContent = copied ? 'Copied' : 'Copy failed — select text';
	button.classList.toggle('copied', copied);
	if (copied) {
		try { navigator.vibrate?.(vibratePattern); } catch { /* Optional feedback. */ }
	}
	setTimeout(() => {
		label.textContent = original;
		button.classList.remove('copied');
		delete button.dataset.copyPending;
	}, copied ? feedbackMs : 5000);
};
