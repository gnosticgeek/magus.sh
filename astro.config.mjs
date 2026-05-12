// @ts-check
import { defineConfig } from 'astro/config';
import tailwindcss from '@tailwindcss/vite';
import mdx from '@astrojs/mdx';
import icon from 'astro-icon';

import cloudflare from '@astrojs/cloudflare';

// https://astro.build/config
export default defineConfig({
  site: 'https://magus.sh',
  integrations: [mdx(), icon()],

  devToolbar: {
      enabled: false,
	},

  vite: {
      plugins: [tailwindcss()],
	},

  adapter: cloudflare(),
});