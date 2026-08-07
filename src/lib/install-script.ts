// The installer is imported from the Go module's scripts directory rather than
// copied into the site, so there is exactly one version of it: what the repo
// ships and what magus.sh serves cannot drift apart.
import script from '../../magus/scripts/install.sh?raw';

export const installScript = script;

/**
 * Serve the installer as plain text.
 *
 * text/plain matters: opening the URL in a browser should show the source
 * rather than downloading a file. Reading it before piping it to a shell is
 * the entire mitigation for `curl | sh`, so it has to be one click away.
 */
export const installScriptResponse = (): Response =>
	new Response(installScript, {
		headers: {
			'Content-Type': 'text/plain; charset=utf-8',
			'Cache-Control': 'public, max-age=300',
		},
	});
