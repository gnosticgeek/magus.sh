# Magus v0.4.5

This patch release makes Magus's built-in Mac updater finish more cleanly while
preserving its existing verification and failure-safety guarantees.

## Highlights

- **Automatic handoff after updating.** Once the latest release has downloaded,
  passed checksum verification, and been installed, Magus now quits so the next
  launch immediately uses the new binary.
- **Safe failures remain visible.** If the update fails, Magus stays open and
  confirms that the existing executable was not changed.
- **Shorter confirmation.** The update prompt now states the outcome directly,
  with Enter clearly confirming the update.

## Notes

Mac remains an alpha. The updater still downloads from GitHub, verifies the
published SHA-256 checksum, and replaces the executable atomically. Preview mode
does not download or write anything.

## Upgrade

Choose **Update Magus** from the Mac menu, or run the usual installer:

```sh
curl -fsSL https://magus.sh/install | sh
```

After a built-in update, launch Magus again to use the new binary. Existing
manifests remain on schema `0.4.0`; v0.4.5 does not require a manifest migration.
