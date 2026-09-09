import { afterEach, describe, expect, it, vi } from 'vitest';
import { copyWithFeedback } from './clipboard';

afterEach(() => { vi.useRealTimers(); vi.unstubAllGlobals(); });

function elements() {
	const classes = new Set<string>();
	const button = { dataset: {} as Record<string, string>, classList: {
		toggle: (name: string, on: boolean) => on ? classes.add(name) : classes.delete(name),
		remove: (name: string) => classes.delete(name),
	} } as unknown as HTMLElement;
	const label = { textContent: 'Copy', setAttribute: vi.fn() } as unknown as HTMLElement;
	return { button, label, classes };
}

describe('clipboard feedback', () => {
	it('only reports success after the write resolves, then restores the label', async () => {
		vi.useFakeTimers();
		let complete!: () => void;
		const writeText = vi.fn(() => new Promise<void>(resolve => { complete = resolve; }));
		vi.stubGlobal('navigator', { clipboard: { writeText } });
		const { button, label, classes } = elements();
		const pending = copyWithFeedback('magus run', button, label);
		expect(label.textContent).toBe('Copy');
		await copyWithFeedback('magus run', button, label);
		expect(writeText).toHaveBeenCalledTimes(1);
		complete(); await pending;
		expect(label.textContent).toBe('Copied');
		expect(classes.has('copied')).toBe(true);
		vi.runAllTimers();
		expect(label.textContent).toBe('Copy');
		expect(button.dataset.copyPending).toBeUndefined();
	});

	it.each([undefined, { writeText: vi.fn().mockRejectedValue(new Error('Denied')) }])('provides manual-copy feedback if clipboard access fails', async clipboard => {
		vi.useFakeTimers();
		vi.stubGlobal('navigator', { clipboard });
		const { button, label, classes } = elements();
		await copyWithFeedback('magus run', button, label);
		expect(label.textContent).toBe('Copy failed — select text');
		expect(classes.has('copied')).toBe(false);
		vi.runAllTimers();
		expect(label.textContent).toBe('Copy');
	});
});
