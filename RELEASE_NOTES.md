# Magus v0.4.7

This Mac alpha release makes a fresh-Mac setup faster to browse, easier to
navigate, and reusable across machines. The manifest schema remains `0.4.0`.

## Highlights

- **Broader catalogue.** Games and Security & Privacy have stronger, more
  useful app choices, and every selectable app and command-line tool now has a
  concise description in its detail view.
- **Developer & Terminal hub.** Terminal tools, fonts, terminal setup and app
  setup now share a compact menu. It also includes Developer environments for
  Apple Container, Node.js and npm, Python, uv, Go, Rust, Docker CLI and Colima.
- **Clearer review flow.** Review & install is part of the selection journey,
  with consistent labels and colours for installed, external, unavailable and
  failed states across the app.
- **Reliable navigation.** Breadcrumbs show the active path, and Escape or
  Backspace follows the screen history rather than guessing a parent screen.
- **Reusable profiles.** Save a named, human-readable TOML selection with
  `magus profile save NAME`; list, inspect and load it later. Loading always
  opens Review & install and never overwrites an existing profile.

## Upgrade

Choose **Update Magus** from the Mac menu, or run:

```sh
curl -fsSL https://magus.sh/install | sh
```

Existing manifests continue to work unchanged; no migration is required.
