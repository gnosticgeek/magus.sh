# Magus v0.4.9

This release makes the Mac TUI easier to navigate and inspect while keeping its
review-before-change safety model. The manifest schema remains `0.4.0`.

## Highlights

- **Search with context.** Fuzzy search now identifies the originating area,
  reports matches and selected items, highlights matching characters, gives an
  actionable empty state, and restores the exact previous list position on
  Escape.
- **Always-visible basket.** Every screen shows a categorized basket summary.
  Press `v` to inspect and edit the basket, then continue separately to Review
  & install—there is no direct installation shortcut from the drawer.
- **Action palette.** `?` now opens a navigable menu of actions valid for the
  current context. Chosen actions follow the exact same guarded code paths as
  their keyboard shortcuts.
- **Reliable navigation and previews.** Typed history now carries filters and
  cursors, while preview results require both the current generation and focused
  target before they can render.
- **Clearer visual semantics.** Rainbow Magus wordmarks now appear across the
  interactive entry points. Success, caution, external, and failed states are
  styled consistently without sacrificing plain-terminal readability.

## Upgrade

Choose **Update Magus** from the Mac menu, or run:

```sh
curl -fsSL https://magus.sh/install | sh
```

Existing manifests and profiles continue to work without a schema change.
