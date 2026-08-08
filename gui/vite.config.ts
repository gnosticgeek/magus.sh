import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	// Tauri drives the dev server; failing loudly beats silently picking a
	// different port that the desktop shell is not pointed at.
	server: { port: 5173, strictPort: true },
	clearScreen: false,
});
