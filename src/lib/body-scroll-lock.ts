export const lockBodyScroll = (): (() => void) => {
	const savedScrollY = window.scrollY;
	const { style } = document.body;
	style.overflow = 'hidden';
	style.position = 'fixed';
	style.width = '100%';
	style.top = `-${savedScrollY}px`;
	return () => {
		style.overflow = '';
		style.position = '';
		style.width = '';
		style.top = '';
		window.scrollTo(0, savedScrollY);
	};
};
