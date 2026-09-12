# Magus v0.4.6

This bug-fix release tightens read-only behavior, uninstall ownership, and Mac
update validation. The catalogue and manifest format are unchanged from v0.4.5.

## Highlights

- **Strict release versions.** The updater now accepts only complete
  `vMAJOR.MINOR.PATCH` release tags. A malformed or unexpected latest-release
  response is ignored instead of being presented as an update.
- **Truly read-only checks.** Linux help, version, doctor, unknown commands, and
  every dry run no longer create Magus directories or sweep temporary files.
- **Ownership-safe uninstall.** Magus now records the Flatpaks it installs and
  marks kitty and GE-Proton directories it creates. Uninstall leaves unmarked
  external installations and launcher files untouched.
- **Reliable cancellation.** Linux probes and mutations now inherit the active
  run's cancellation context, and timeouts stop the complete subprocess group,
  so interrupted installers cannot continue changing the machine in background.
- **Shell-safe GE-Proton downloads.** Release metadata is parsed as JSON, assets
  must belong to the expected upstream release path, and download/extraction use
  direct process arguments instead of interpolating a URL into a shell command.

## Notes

Mac remains an alpha. The built-in updater downloads from GitHub, verifies the
published SHA-256 checksum, and replaces the executable atomically. Preview mode
does not download or write anything.

## Upgrade

Choose **Update Magus** from the Mac menu, or run the usual installer:

```sh
curl -fsSL https://magus.sh/install | sh
```

After a built-in update, launch Magus again to use the new binary. Existing
manifests remain on schema `0.4.0`; v0.4.6 does not require a manifest migration.
