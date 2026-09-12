# Magus v0.4.3

This point release streamlines the Mac TUI and adds a safe, built-in path for
keeping Magus itself current. It does not broaden what Magus changes: the
review-first flow, preview mode, and user-file protections remain the same.

## Highlights

- **Built-in Magus update check.** The menu checks GitHub Releases in the
  background. When a newer release exists, **Update Magus** is marked in green
  and shows the available version.
- **Verified self-update.** After confirmation, **Update Magus** downloads the
  release through the project installer, verifies its published SHA-256 checksum,
  and atomically replaces the current executable. Restart Magus to use it.
- **Simpler menu navigation.** The unused package-preset menu has been removed,
  and Backspace now works as a back key everywhere except the search box, where
  it continues to delete text.

## Notes

Mac remains an alpha. Existing app files and Finder settings are still backed
up before Magus manages them, and restoration continues to leave later manual
edits alone. The update check is read-only; a failed or offline check stays
quiet. Steam Machine hardware detection and its HDMI/CEC and power preferences
remain alpha boundaries; this release does not change them.

## Upgrade

Choose **Update Magus** from the Mac menu, or run the usual installer:

```sh
curl -fsSL https://magus.sh/install | sh
```

Then open a new terminal and run `magus run`. Existing manifests remain on
schema `0.4.0`; v0.4.3 does not require a manifest migration.
