# Mac app setups

Open **App setups** in the Mac menu. This ports the Ghostty, Zed, Firefox,
Modern CLI and Raycast setup work from Machinist / Athanor into Magus.
Selecting a preset adds it to **Review & apply**; nothing is applied just by
opening a setup. Installing an app alone does not select its preset.

## Available setups

- **Ghostty:** retains Magus's three light/dark themes and font installation,
  with Athanor's window restoration, Display P3 rendering, text thickening,
  clipboard trimming, copy-on-select, Option-as-Alt and shell integration.
  Its installed validator checks the generated file before replacement.
- **Zed:** 14pt text, system theme, formatting and whitespace cleanup on save,
  telemetry off, inline blame, steady cursor and build-folder exclusions.
  Close Zed before applying. This replaces the whole settings file; export
  instead if you want to merge selected values into an existing setup.
- **Firefox:** reduced telemetry and promotional clutter, tracking protection,
  Global Privacy Control and session restoration via `user.js`. Launch Firefox
  once and quit before applying. Magus resolves a single registered profile;
  for multiple profiles, export and use `about:profiles` to choose the destination.
  The five add-on links from Athanor are available for manual installation.
- **Modern CLI:** adds zoxide's `z`, fd/bat-powered fzf previews, coloured man
  pages, thefuck and a bat alias to Magus's existing configurable shell setup.
  Retains Magus's btop replacement for top. Required tools are included in the
  review plan. Each integration is guarded for missing tools and only runs in
  interactive Zsh. Enabling both zoxide options provides `z` and `cd` integration.
- **Raycast:** six settings tips, six extension pages, an installation basket
  action, import/export documentation and an exportable Markdown guide.
  Settings and extensions are configured inside Raycast. The guide is not a
  generated `.rayconfig` bundle and contains no account data.

## Backups and restoration

Zed and Firefox preserve the original file and permissions before writing.
Snapshots live in the Magus state directory, with retry support for interrupted
writes. Magus refuses symlinked configuration paths and refuses to replace or
restore files edited after it applied a preset.

**Restore Magus settings** restores recorded Zed/Firefox files along with the
existing Mac preference and shell restoration. Ghostty retains its separate
restore action under Terminal setup. Installed apps remain installed.

**Firefox limitation:** restoring or deleting `user.js` does not reset values
Firefox has already copied into `prefs.js`. Reset affected preferences through
`about:config` when undoing the preset; the exported file lists those values.
The port keeps the about:config warning enabled and omits the retired Pocket
switch from the original preset. No browsing data is modified by Magus.

## Portable exports and manifests

Exports are saved under the Magus config directory:

- `exports/zed/settings.json`
- `exports/firefox/user.js`
- `exports/raycast/README.md`

An existing export is never overwritten. Ghostty's export remains in Terminal
setup. Preview mode makes no exports or configuration changes and opens no links.

Headless Mac manifests can include:

```toml
[mac]
app_configs = ["zed", "firefox"]
packages = []
settings = []
```

The reconciler includes the required apps before their configuration steps.
Use the existing `reconcile --dry-run` or `doctor` commands for inspection.
Unknown or duplicate configuration IDs are rejected. Raycast's guided setup
is intentionally not an automatic manifest configuration step.

## References

- [Ghostty configuration](https://ghostty.org/docs/config/reference)
- [Zed settings](https://zed.dev/docs/reference/all-settings)
- [Firefox profiles](https://support.mozilla.org/en-US/kb/profiles-where-firefox-stores-user-data)
- [Raycast import/export](https://manual.raycast.com/import-export)
- [fzf](https://github.com/junegunn/fzf) and [bat](https://github.com/sharkdp/bat)

Validation covers temporary-home apply/restore, dry runs, preservation of edits
and file modes, pending-write recovery, symlinks, Firefox ambiguity and invalid
paths, manifest round trips, ordered app dependencies, export protection and
terminal screen bounds. Live Zed/Firefox behaviour still needs app-level testing.
