# Magus v0.4.1

This point release makes the Mac alpha easier to install, browse, and recover
from when something needs attention. It keeps the same careful review-first
approach: Magus still changes nothing until you confirm your selection.

## Highlights

- **Mac installs are ready for release.** The verified installer now supports
  both Apple silicon and Intel Mac binaries. It downloads the matching build,
  checks it against the release checksum, and installs Magus without applying
  any setup choices.
- **Find the right apps faster.** In the Mac catalogue, press `f` to cycle
  between all items, items not installed, and the current selection. Search,
  categories, and your review basket keep working as before.
- **Clearer help when an install fails.** The installation view names the item
  that failed, shows the useful error detail, and includes the Magus version,
  Mac version, and item ID needed for a useful bug report. Retry with `r`, skip
  with `s`, or open the full log with `l`.
- **A more useful finish line.** Completion now separates installed, already
  present, skipped, and failed items, then points to the appropriate next step:
  inspect logs, refresh Finder, return to the menu, or run `magus doctor`.
- **One release number across the product.** The website, installer card,
  package metadata, TUI prototype, and desktop-app fixture now identify this
  release consistently. A test prevents the website and package versions from
  drifting apart.

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
schema `0.4.0`; v0.4.1 does not require a manifest migration.
