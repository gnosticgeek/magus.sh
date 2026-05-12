export const copyWithFeedback = async (
	text: string,
	button: HTMLElement,
	label: HTMLElement,
	vibratePattern: number | number[] = 10,
	feedbackMs = 1600,
): Promise<void> => {
	if (!text) return;
	try {
		await navigator.clipboard.writeText(text);
	} catch {
		/* clipboard unavailable */
	}
	if ('vibrate' in navigator) navigator.vibrate?.(vibratePattern);
	const original = label.textContent ?? '';
	label.textContent = 'Copied';
	button.classList.add('copied');
	setTimeout(() => {
		label.textContent = original;
		button.classList.remove('copied');
	}, feedbackMs);
};
