// @ts-check
import { defineConfig } from 'astro/config';
import tailwindcss from '@tailwindcss/vite';
import mdx from '@astrojs/mdx';
import icon from 'astro-icon';

import cloudflare from '@astrojs/cloudflare';

/* `astro dev` under the Cloudflare adapter's vite-plugin runner currently throws
   "module is not defined" because Miniflare can't load Astro's dynamic page
   imports as ESM. The site is fully static (all pages prerender), so the
   adapter is only needed for `build` / `preview` / `deploy`. Skip it for `dev`. */
const isDev = process.env.npm_lifecycle_event === 'dev';

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

  adapter: isDev ? undefined : cloudflare(),
});