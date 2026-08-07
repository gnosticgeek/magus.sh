import type { APIRoute } from 'astro';
import { installScriptResponse } from '../lib/install-script';

// Prerendered like every other page — a static file with a nicer URL.
export const prerender = true;

// The descriptive name. /run serves the same script, because that is the
// command the homepage has always advertised.
export const GET: APIRoute = () => installScriptResponse();
