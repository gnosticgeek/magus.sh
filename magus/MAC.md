# Magus for Mac — local alpha

A single basket for Mac apps, terminal tools, and six reversible Finder preferences.
The Mac alpha is available from the verified installer for Apple silicon and Intel
machines; local builds remain useful for development.

## Try it

Install a published build:

```sh
curl -fsSL https://magus.sh/install | sh
magus run
```

Or build locally with Go 1.26.3 or later, from the repository's `magus` directory:

```sh
go build -o dist/magus .
./dist/magus preview   # real catalogue and inspection; no installation or writes
./dist/magus           # browse, review, then explicitly install your selection
```

`magus run` opens the same menu, including after a previous setup. Picks are
restored from the last confirmed manifest, then installed/already-applied items
are removed from the basket after inspection. A new machine starts with no picks.
The home screen uses the website's lavender, pink and ice-blue gradient across
an ASCII seal and wordmark, with a compact mark in shorter terminals. Colours
adapt to light backgrounds and plain terminals.

Apps opens thirteen categories, including AI & LLMs, Audio & Music, Design &
Graphics, Cloud & Storage, Security & Privacy, Games, Menu Bar and Utilities.
Each shows its selection count and app
preview. Escape returns to the category list and keeps your focus and picks.
Fuzzy search spans all categories and also matches their names. Bubbles List
provides paging, result counts and custom rows with a separate basket. The AI category adds
Ollama and LM Studio alongside desktop assistants; model downloads remain separate.

The catalogue includes 53 apps, six fonts and 16 command-line tools, including Mole.
**Fonts** opens a handpicked selection: Inter, Source Serif 4, Newsreader, Fraunces,
Space Grotesk and JetBrains Mono. Select fonts individually or use Ctrl+S for all
six, then review and install through the shared basket. Choose the installed font
in your app; some apps need reopening.

The main menu calls this section **Command-line tools**. App and tool browsing
uses columns at 80 columns wide, with a side preview from 120 columns and a
single list below 80. Tab opens details at any size. Arrow keys move between
cells; Page Up/Down change pages. During search, Left/Right edit the query.
**Ctrl+S** selects all results in the current category or search, across pages;
press it again to deselect those results. Installed items are excluded from
individual selection, presets, and select-all. Other selections stay untouched.
Enter on an app or tool opens details; only Space toggles an individual selection.
Mole is installed only; its cleanup and removal actions must be started separately.
Installed entries show a
green tick (a `+` in plain mode), independently of the basket checkbox. External
app bundles are labelled `external` and are left alone. The local reference
inventory is in [INSTALLED_APPS.md](INSTALLED_APPS.md).

**Update all** first shows a scrollable Bubbles Table of installed and available
versions from local Homebrew metadata, including packages outside the catalogue.
Press `r` to refresh metadata explicitly (disabled in preview). Enter opens a
separate confirmation screen. Before upgrading, Magus rechecks the version list;
if it changed, review again. Only the reviewed targets are requested, with
Homebrew resolving their dependencies. Its live output and prompts take over the terminal;
Ctrl+C stops the run. Preview never executes updates. The installation lock also
covers updates, and inventory is refreshed afterwards. No automatic cleanup is
requested. Homebrew's default exclusions for pinned formulae and self-updating
or unversioned casks apply. This does not update App Store or unmanaged apps.
Real upgrades have not been run as part of development validation.

- Arrows move, Page Up/Down change pages, Home/End jump to first/last.
  Space selects, Enter opens, Escape goes back. Selection notices clear after
  three seconds without clearing newer error messages.
- `/` searches across all categories. Letters—including `r`, `q`, and `?`—are
  text while searching. Space toggles the focused result.
- Tab focuses the scrollable details pane (full width on a narrow terminal).
  Arrows and Page Up/Down scroll it; Tab or Escape returns to the list.
  `?` explains the controls for the current screen, including during installation.
- Presets add selections and never remove existing picks.
- Review shows the entire basket before installation. Missing Homebrew can be
  installed by pressing `b` in Review; Magus hands the terminal to the official
  Homebrew installer, then inspects prerequisites again.
- During installation, a spinner, smoothly animated item-completion bar, current item, elapsed time, and
  latest Homebrew message remain visible. `l` shows full logs; a failed item
  offers `r` retry, `s` skip, or `q` stop. Completion is item-based because
  Homebrew does not expose reliable byte-level progress. Completed packages
  remain installed. Reopen the menu to resume. Scrolling logs backwards pauses
  auto-follow; End returns to live output.
- The summary can be scrolled with arrows. `f` refreshes Finder explicitly.

Homebrew installs one item at a time. Already-installed formulae/casks are skipped
without upgrades. Apps found in `/Applications` or `~/Applications` outside
Homebrew are left alone and reported as such. Packages are never automatically
adopted, replaced, or removed. Homebrew retains responsibility for its own
platform requirements and dependency resolution. Shell changes are optional: Terminal setup → Modern commands lets you enable individual
Zsh integrations, review them, and undo them later. direnv remains manually configured.

## Headless commands and files

```sh
magus reconcile --dry-run --json
magus reconcile
magus doctor --json
magus restore --dry-run
magus restore
```

`run --defaults` on a fresh Mac writes an **empty** manifest, with no implicit
package or setting choices. Use `reconcile` for unattended execution of a saved
manifest. `--manifest` chooses a different manifest; platform mismatches and
unknown IDs are rejected before execution. Mac uninstall is deliberately not
implemented; `restore` handles recorded preferences only.

- `~/.config/magus/manifest.toml`: confirmed intent, schema 0.4.0, platform `darwin`.
- `~/.local/state/magus/preferences.json`: original typed values, including absence.
- `~/.local/state/magus/last-run.json`: outcome of the last confirmed operation.
- `~/.local/state/magus/install.lock`: exclusive operation lock, automatically
  released by the OS on exit. Do not delete this file to unlock a running process.

The existing XDG overrides still apply. Shared settings history belongs to the
machine, even when using different manifest files. A restored key is removed
from the current manifest so `reconcile` does not immediately apply it again.
Other saved manifests remain independent.

Preference snapshots are saved before writes and preserved across retries.
Restoration verifies the original value and leaves externally modified settings
alone. Existing values with unsupported types fail inspection rather than being
coerced. Refreshing Finder is optional and is never part of an automated run.

## Distribution

The installer supports `magus-darwin-arm64`, `magus-darwin-amd64`, and the existing
Linux assets. A valid SHA-256 entry in `checksums.txt` is mandatory. Mac uses
`shasum` when `sha256sum` is absent. On Mac, an interactive download opens the
menu using `/dev/tty`; a non-interactive download prints the launch command.
`MAGUS_NO_LAUNCH=1` suppresses launch. Mac installation does not modify shell
profiles: use the printed absolute launch path if `~/.local/bin` is not on PATH.
Linux retains its separate launch and existing PATH behavior.

Release publishing waits for Linux and macOS checks. Local builds and publishing
are separate. No accounts or telemetry are added by Magus.

## Validation and remaining hardware checks

```sh
go test ./...
go test -race ./...
go vet ./...
```

The default tests use fake executables and preference data, not the user's Homebrew or
Finder settings. `MAGUS_TEST_MAC_DEFAULTS=1 go test -run TestMacRealPreferenceFile`
also exercises the real macOS defaults command against an isolated temporary plist
(outside restricted sandboxes). That apply/restore integration passed on macOS 27.0. They cover formula/cask installation and reruns, inspection
failure, unmanaged applications, dry runs, exact restoration, external edits,
retry/stop, operation locks, child-process cancellation, manifest isolation,
selection/search behavior, terminal sizes, and verified installer downloads.

### TUI architecture and asynchronous state

The Mac Bubble Tea implementation is divided by ownership. `mac_tui.go` contains
the model and catalogue projection, `mac_update.go` contains message-driven
asynchronous transitions, `mac_key.go` contains keyboard routing, and
`mac_view.go` renders without starting I/O. Read-only inventory collection lives
in `mac_inventory.go`, the Homebrew bootstrap lifecycle lives in
`mac_bootstrap.go`, and installation workers remain in `mac_execution.go`.

Screen destinations and installation event kinds are typed in `mac_state.go`.
Every asynchronous UI subsystem uses the generation token in `mac_async.go`.
Results from a canceled inventory probe, update check, notification timer, or
previous installation session are ignored once a newer generation exists. Tests
pin current-result acceptance, stale-result rejection, and terminal-size limits.

Package inspection takes one Homebrew list snapshot, then uses no more than four
concurrent package/filesystem probes. Starting a new inspection cancels the
older one; preference inspection stays sequential because it uses macOS system
preferences rather than the package-probe resource pool.

`mac-catalogue-audit.json` records the official Homebrew metadata check for the
original 32 package identifiers plus seven additional Menu Bar apps. Another 29 casks use locally installed
Homebrew metadata, recorded in `mac-installed-catalogue-audit.json`.
The catalogue is embedded in Go and browses offline.
Descriptions, presets, and preference definitions are adapted from the user's
Machinist/AthanorKit source; no Swift runtime or source checkout is needed.

The local development host is Apple Silicon on macOS 27.0. Do not describe
preference UI behavior, fresh Homebrew installation, or Intel hardware as
verified solely from these tests. Before public release, manually exercise:

1. One real formula and cask on a disposable Mac; rerun and confirm both skip.
2. Apply/read back each Finder setting, visually confirm it, then restore it.
3. Missing Homebrew and command-line tools in a disposable Mac environment.
4. The menu in macOS Terminal at 80×24 and larger, plus physical Intel testing.

Linux catalogue expansion, shell themes, Brewfiles, App Store apps, and updates
remain outside this Mac alpha.

## Ghostty terminal setup

Menu item **8. Terminal setup** offers Catppuccin, TokyoNight and Rose Pine.
Enter adds Ghostty, JetBrains Mono and a configuration step to the basket. Escape
returns to the menu; **Review & install** applies the selection. The profile writes
`~/.config/ghostty/config.ghostty` (respecting XDG_CONFIG_HOME) and backs up the
original file and permissions in Magus state. Reload Ghostty with Cmd+Shift+,.

Existing legacy/macOS-specific settings that could override the profile, symlinks,
and manually edited Magus configs are left unchanged. **Export selected Ghostty
dotfile** saves a portable copy under the Magus config directory at
`exports/ghostty/config.ghostty`; an existing export is never overwritten.
**Restore previous Ghostty settings** restores the backup after confirmation,
provided the generated file is unchanged. Preview mode performs no writes.
This configures Ghostty's appearance, not the shell prompt or iTerm2.

## Modern terminal commands

The Developer preset includes eza, zoxide, btop, dust, duf, Atuin and tlrc alongside
ripgrep, fd, fzf, bat and delta. `tlrc` is the maintained Homebrew client providing
`tldr`; the old `tldr` formula is disabled.

Under **Terminal setup → Modern commands**, turn on **Use modern commands in my
terminal**, choose individual integrations, then **Review & apply**. This is off
by default; installing a preset alone does not enable shell changes. The initial
switch choices are ls/eza, tree/eza, top/btop, cd/zoxide, fzf and delta. Atuin is
separately selectable and takes Ctrl+R after fzf when both are enabled. It records
commands locally; Magus does not create an account or configure history sync.

Required packages appear in the basket and are installed before configuration.
Magus appends a marked, interactive-only block to `.zshrc` in inherited absolute
`ZDOTDIR`, or the home directory when unset. Set/export ZDOTDIR before launching
Magus if your shell uses a different configuration directory. Existing content
and permissions are retained; the original file is backed up in Magus state.
Symlinked configurations and edited/unrecognised Magus blocks are left unchanged.
Edits outside the block survive updates and undo. Open a **new terminal** after
applying, changing switches, or undoing; existing sessions retain their definitions.

The original tools remain available (for example `/bin/ls` and `builtin cd`).
Some old flags differ. grep, find, cat, sed, curl, du, df and man keep their names
and behaviour. Git uses `GIT_PAGER=delta` in interactive terminals only when no
GIT_PAGER is already set; global Git configuration is not rewritten.

**Undo modern terminal commands** queues removal for review. `magus restore`
also removes the recorded block, preserving installed tools and unrelated edits.
Dry runs and preview do not change shell files or backups.

Headless manifests can use:

```toml
[mac.modern_shell]
enabled = true
commands = ["ls", "tree", "top", "cd", "fzf", "delta"]
```

Set `enabled = false` and reconcile to undo. Omit the section to leave shell
configuration unmanaged. Accepted command IDs also include `atuin`.

## App setups

The Mac menu includes **App setups** for Ghostty, Zed, Firefox, Modern CLI and
Raycast. See [setup contents, exports and restore limitations](APP_SETUPS.md).
