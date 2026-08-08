import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
export default {
	preprocess: vitePreprocess(),
	kit: {
		// Tauri serves a static bundle from disk — there is no Node server in
		// the app, so everything must prerender.
		adapter: adapter({ fallback: 'index.html' }),
	},
};
