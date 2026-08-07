import type { APIRoute } from 'astro';
import { installScriptResponse } from '../lib/install-script';

export const prerender = true;

// The homepage has advertised `curl -fsSL https://magus.sh/run | bash` since
// before there was anything to download, so this route has to exist and has to
// serve the installer. /install is the same script under a clearer name.
export const GET: APIRoute = () => installScriptResponse();
