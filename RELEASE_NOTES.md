# Magus v0.4.2

This release makes the Mac TUI easier to evolve safely and more resilient when
the system is slow. It does not broaden what Magus changes: the review-first
flow, preview mode, and user-file protections remain the same.

## Highlights

- **Clearer Mac TUI boundaries.** The model, rendering, keyboard routing,
  asynchronous updates, inventory checks, and Homebrew bootstrap lifecycle now
  have distinct, documented owners. This is an internal improvement with no new
  setup action or hidden behaviour.
- **Safer in-flight work.** Inventory, update, notification, and installation
  results carry an operation identity. If you leave a screen, refresh, or begin
  a newer operation, an old buffered result cannot overwrite the current view.
- **Bounded Mac inventory checks.** Magus takes one Homebrew metadata snapshot,
  then runs no more than four package/filesystem checks at once. Starting a new
  inspection cancels the old one; macOS preference checks remain sequential.
- **Stronger regression coverage.** The test suite now covers late messages,
  worker-pool bounds, typed screen capabilities, and all Mac screens at narrow
  and wide terminal sizes.

## Notes

Mac remains an alpha. Existing app files and Finder settings are still backed
up before Magus manages them, and restoration continues to leave later manual
edits alone. Steam Machine hardware detection and its HDMI/CEC and power
preferences remain alpha boundaries; this release does not change them.

## Upgrade

Run the usual installer again to fetch the latest binary:

```sh
curl -fsSL https://magus.sh/install | sh
```

Then open a new terminal and run `magus run`. Existing manifests remain on
schema `0.4.0`; v0.4.2 does not require a manifest migration.
