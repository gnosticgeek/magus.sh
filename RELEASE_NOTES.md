# Magus v0.5.0

This release gives the Mac TUI a clearer home for AI, agent skills, and developer
languages, restores the full Magus identity, and rounds out the app catalogue.
The manifest schema remains `0.4.0`.

## Highlights

- **AI & agents hub.** AI apps and shared Codex and Claude skills now have a
  dedicated top-level menu, keeping agent workflows out of general app setup.
- **Languages in one place.** Node.js, Python, uv, Go, and Rust now live in a
  focused Languages & runtimes menu under Developer & Terminal.
- **Ghostty selection fix.** A selected Ghostty theme can now be toggled off
  instead of becoming stuck in the review basket.
- **The big wordmark returns.** The full block-letter Magus logo is back on the
  Mac home screen with an animated rainbow gradient and a plain-terminal
  fallback.
- **Stronger app categories.** Sparse categories now contain at least five
  choices, favouring open-source additions including LibreWolf, Element,
  Mattermost, Thunderbird, Strawberry, Mixxx, LMMS, Nextcloud, Syncthing, and
  Cryptomator.

## Upgrade

Choose **Update Magus** from the Mac menu, or run:

```sh
curl -fsSL https://magus.sh/install | sh
```

Existing manifests and profiles continue to work without a schema change.
