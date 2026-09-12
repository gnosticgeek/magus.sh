# magus — Go binary

**Mac local alpha:** [Mac TUI guide](MAC.md). On macOS, no arguments and
`magus run` now open the same basket-based menu. `magus preview` inspects without
installing. The older SteamOS flow described below is retained.


One statically-linked binary with two faces:

- **No arguments** — the Bubble Tea TUI that walks the five magus.sh stages and
  writes `/tmp/magus.sh`. Replaces the @inquirer/prompts TUI proposed in
  [`TUI_SPECIFICATION.md`](../TUI_SPECIFICATION.md) — see the *Implementation
  Notes* section there for the visual + navigation contract.
- **With a verb** — the reconciler from
  [`STEAM_MACHINE_BRIEF.md`](../STEAM_MACHINE_BRIEF.md): reads a manifest and
  converges the machine to it.

## Install

```bash
curl -fsSL https://magus.sh/install | sh
```

Downloads the binary for this machine's architecture, verifies its SHA-256
against the release's `checksums.txt`, and puts it in `~/.local/bin`. `/run`
serves the same script — that's the URL the homepage advertises.

**It installs the tool and stops.** It changes no settings; you run `magus run`
yourself afterwards. Two reasons: a pipe-to-shell that also starts reconfiguring
your desktop asks for a lot of trust in one keystroke, and it couldn't work
anyway — the wizard needs a terminal, and a pipe isn't one.

The script is served as `text/plain`, so <https://magus.sh/install> is readable
in a browser before you run it. Source: [`scripts/install.sh`](scripts/install.sh).

Binaries come from GitHub Releases, published by
[`.github/workflows/release.yml`](../.github/workflows/release.yml) when a `v*`
tag is pushed. `MAGUS_VERSION` pins a tag; `MAGUS_BIN_DIR` changes the target.

## The reconciler

```bash
magus run --defaults     # write the opinionated manifest, then converge
magus reconcile          # converge to the existing manifest, no questions
magus doctor             # report drift; changes nothing; exits 1 if drift found
magus uninstall          # reverse what magus installed
magus <verb> --dry-run   # report what would change without changing it
```

### Machine-readable output

Every verb takes `--json`. The document goes to **stdout**; the human log stays
on **stderr** and is silenced, so the two can be captured separately.

```bash
magus doctor --json | jq '.steps[] | select(.needs_attention)'
```

This is the contract a GUI talks to, and it exists so a front end never composes
a shell string. EmuDeck's Electron wrapper pipes arbitrary bash from its renderer
straight to `exec`, which turns any front-end bug into arbitrary command
execution — the alternative, and the one omasteam reached independently (§8), is
that every action is a *named* command the backend dispatches.

Three versions travel in the document because they move for different reasons:
`schema` (this contract), `magus.version` (the binary), and
`magus.manifest_schema` (the manifest format). Each step carries its state
before and after, whether it changed, `why` it was skipped, and a single
`needs_attention` boolean for a front end to colour on.

**Failures appear in the document**, not only in the exit code — otherwise a
caller parsing stdout sees a valid-looking empty result and reads a failure as
"nothing to do". Exit codes are unchanged either way.

The manifest at `~/.config/magus/manifest.toml` is the only thing that connects
the two halves. The wizard's job is to produce it; everything downstream reads
only it, so `--defaults`, a hand-edited file and the eventual GUI all converge
through identical code.

**Re-running is the normal case.** Every step derives its state from the
filesystem rather than from a record of what it did last time, so a SteamOS
atomic update that removes an artifact shows up as drift on the next `doctor`
and is repaired by the next `reconcile`. There is no installed-version file to
fall out of sync with reality. Narrow ownership receipts exist only to make
uninstall safe: Magus records Flatpaks it installs and marks kitty and GE-Proton
directories it creates. Unmarked packages and directories are treated as
external and left in place. Read-only and dry-run commands do not create Magus
directories or sweep temporary files. Linux subprocesses inherit the active
run's cancellation context.

Reconciler files:

```
paths.go            # the ~/.local layout, atomic writes, temp sweeping
device.go           # Deck vs Machine detection (MAGUS_DEVICE overrides)
manifest.go         # schema, validation, migration, TOML I/O
reconcile.go        # the Step interface and the Reconcile/Doctor/Uninstall engines
steps.go            # manifest → ordered plan; the bundle catalogue
steps_flatpak.go    # Flathub remote + the generic flatpak step
steps_terminal.go   # kitty, keep-Konsole, and the not-yet placeholder
desktop.go          # .desktop entries (absolute Exec, always)
report.go           # the four output verbs: Section / OK / Warn / Die
```

### Testing it without Steam hardware

Unit tests run against a temp `$HOME`, so they cannot touch your dotfiles:

```bash
go test ./...
```

`scripts/sandbox.sh` exercises the whole CLI against a throwaway `$HOME`. The
default is a dry run — writes nothing, downloads nothing:

```bash
scripts/sandbox.sh
```

`--real` installs for real and runs the drift/repair loop: converge, delete an
artifact the way an atomic update would, prove `doctor` exits 1, prove one
`reconcile` repairs it, then uninstall. It is Linux-only, because kitty's
installer targets `/Applications` on macOS — `--stub-kitty` skips that download
and works anywhere.

```bash
scripts/sandbox.sh --real --keep
```

Bundles are excluded by default; `--bundles` adds them and pulls roughly a
gigabyte of Flatpak runtimes on first run. `MAGUS_DEVICE` (or `--device`) forces
the device branch — the reconciler never needs real Valve hardware to run.

### One tool, two devices

Deck and Steam Machine run the same wizard, write the same manifest and converge
through the same reconciler. Only the opinions differ:

| Optimisation | Deck | Steam Machine | Status |
|---|---|---|---|
| Disable Baloo (file indexer) | ✓ | ✓ | built |
| Double-click to open | ✓ | ✓ | built |
| Cursor size | 32px (thumbs) | 48px (across a room) | built |
| Proton-GE | ✓ | ✓ | built |
| HDMI colour range → full RGB | — | ✓ | recorded, not built |
| HDMI-CEC | — | ✓ | recorded, not built |
| Power profile | balanced | performance | recorded, not built |

The Deck gets no HDMI fixes because its display is its own, and no performance
profile because it runs off a battery — applying the console's set to a handheld
is the mistake §2 diagnoses, pointed the other way.

Two catalogue optimisations are deliberately excluded: the Wi-Fi power-save
tweak needs root, and the Btrfs `/home` conversion is irreversible.

### What is not built yet

Not yet: Ghostty and Alacritty (picking them yields a step that says so rather
than silently doing nothing), the three hardware-gated optimisations above, and
theming. The wizard still asks and the manifest still records, so an install done
today needs no surgery to gain them later — and `magus doctor` prints *why* each
is skipped rather than a bare `n/a`.

`magus run` needs a terminal to draw the wizard on. Piped, cron'd or otherwise
non-interactive, it exits 2 and points at `--defaults` rather than silently
taking choices the user never made.

## Quick start

```bash
# from the repo root
brew install go              # one-time, only needed to build
cd magus
go run .                     # iterate
go build -o magus .          # produce a single binary (size depends on build flags)
./magus
```

Or from the repo root: `npm run magus` (see `package.json`).

## Layout

```
magus/
├── go.mod          # module name: magus
├── commands.json   # embedded catalogue (regenerate via scripts/gen-commands.mjs)
├── scripts/
│   └── gen-commands.mjs  # walks src/content/commands → commands.json
├── data.go         # types + JSON load + catalogue lookups
├── styles.go       # lipgloss palette + status-bar key hierarchy
├── main.go         # Model, Update, View dispatcher
├── splash.go       # splash screen
├── menu.go         # pick menu (split pane)
├── stage.go        # pick stage (groups + commands + select-all)
├── search.go       # fuzzy search across catalogue
├── install.go      # per-stage install animation
├── review.go       # cross-stage review
├── write.go        # script-write animation
├── run.go          # final-run prompt + execution
└── (reconciler files — see "The reconciler" above)
```

## Refreshing the catalogue

After editing any `src/content/commands/**/*.md` frontmatter, regenerate:

```bash
node magus/scripts/gen-commands.mjs
```

The Go binary embeds `commands.json` at compile time via `//go:embed`, so
remember to rebuild after regenerating.

## Distribution

Target is `linux/amd64` (Steam Deck). To produce a release artefact:

```bash
GOOS=linux GOARCH=amd64 go build -o magus-linux-amd64 .
```

In practice nobody builds it by hand: the setup stage's "Install magus" command
and `magus.sh/install` both fetch the released binary. Single file, no runtime,
and it survives SteamOS's immutable updates because it sits in `~/.local/bin/`.

## Why Bubble Tea over @inquirer/prompts

*(Historical — the @inquirer prototype under `tui/` has since been deleted. The
reasoning is kept because it is why the Go binary looks the way it does.)*

- The spec describes a full-screen stateful app with split-pane previews,
  animated progress bars, a documented state machine, and key-hierarchy hints.
  Bubble Tea's Elm-style update/view/model maps onto `PickState` 1:1; Lipgloss
  handles the `t-bright`/`t-text`/`t-muted` palette as composable styles.
- `@inquirer/prompts` is built for one-question-at-a-time inquiry — full-screen
  TUIs are out of scope, you'd reimplement most of the rendering by hand.
- Static binary distribution beats the nvm/Node/tsx bootstrap on every axis:
  fewer steps, smaller surface, faster startup.

## Terminal interface

The terminal UI uses Bubble Tea 2, Bubbles 2 and Lip Gloss 2. Steam catalogue
search supports fuzzy matching, editable/pasted text, and paging through every
result. Enter toggles a result; spaces are part of the search query. Progress
uses the shared Bubbles component. Details use Glamour Markdown rendering.
Mac confirmations use layered dialogs when the terminal has room, with an
ordinary page on smaller screens.

### Mac implementation boundaries

The Mac interface keeps machine work outside rendering. `mac_tui.go` owns the
model and catalogue projection, `mac_update.go` owns asynchronous state
transitions, `mac_key.go` owns keyboard routing, and `mac_view.go` owns
rendering. Inventory inspection and Homebrew bootstrapping live in
`mac_inventory.go` and `mac_bootstrap.go`; screen and event types live in
`mac_state.go`.

Inventory, update, notification, and installation results carry a generation
from `mac_async.go`. The update loop accepts a result only while its generation
is current, preventing canceled or superseded work from changing a newer screen.
Installation identity is attached at the TUI channel boundary so the execution
worker remains independent of navigation.

Package inspection uses at most four concurrent filesystem/package probes after
one Homebrew metadata snapshot. A new inspection cancels the previous probe;
macOS preference reads remain sequential rather than competing for the same
system service.

The SteamOS `magus run` setup wizard uses Huh forms. Tab/Shift+Tab move through
fields, and the final review offers Apply, Edit choices, or Cancel. For a
screen-reader-friendly setup in an interactive terminal:

```sh
MAGUS_ACCESSIBLE=1 magus run
```

This mode uses plain numbered prompts, prints the review, and requires an
explicit yes before applying. Enter `0` to finish the bundle selection. Closing
input cancels setup. This option applies to the SteamOS manifest wizard; the
Mac catalogue remains a full-screen interface.
