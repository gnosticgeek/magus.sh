# magus-gui

The desktop app. SvelteKit + TypeScript in a Tauri shell, talking to the `magus`
binary over its `--json` contract.

## The one rule

**The front end names a command; it never supplies one.**

```
  Svelte view                Rust                     magus
  ───────────                ────                     ─────
  client.run('doctor')  →  Command::Doctor  →  magus doctor --json
                           (fixed argv)             (stdout)
```

`CommandName` is a closed TypeScript union; `Command` is a Rust enum. Serde
rejects anything that is not a known variant before it reaches a function body,
and `std::process::Command` executes the binary directly with an argument
vector — no `sh -c`, so quoting and word-splitting never arise.

This is deliberate. EmuDeck's Electron wrapper exposes
`ipcMain.on('bash', (_, cmd) => exec(cmd))`: the renderer hands over a shell
string and the main process runs it. For a local app the blast radius is
bounded, but it makes any front-end bug arbitrary command execution rather than
a rendering glitch. There is no equivalent path here.

## Running it without Rust

The UI is fully exercisable in a plain browser. `createClient()` detects Tauri;
outside it, or when no `magus` binary is reachable, it falls back to fixtures
and the window says so.

```bash
npm install
npm run dev          # http://localhost:5173, fixture-backed
```

That is the same property the Go side has — every action expressible without
the wizard (§9) — applied to the front end. It means the UI can be built,
reviewed and screenshotted with no Rust toolchain, no Tauri window and no Steam
hardware.

The fixtures are deliberately awkward rather than flattering: a drifted step, a
step that fails with a real flatpak error, a skipped step with a reason, and an
unconfident device. Repairing clears what it can and **leaves the failure**, so
the partial-success case is the default thing you see.

## Running the desktop shell

Needs Rust and the Tauri CLI:

```bash
cargo install tauri-cli --version '^2'
npm run tauri dev
```

## Theming

One `data-theme` attribute on `<html>` re-colours everything; no component
contains a hex value. The four palettes are the same ones the wizard offers and
that the Plasma colour scheme and kitty config are generated from — one choice
driving every surface is the §4-step-5 claim, and here it is literal.

## Layout

```
src/lib/magus/types.ts        the JSON contract, mirrored in TypeScript
src/lib/magus/index.ts        createClient() — picks an adapter
src/lib/magus/adapters/
  tauri.ts                    real: invoke() into Rust
  mock.ts                     fixtures, for browser development
src/lib/theme.ts              palettes + persistence
src/lib/components/           presentational, no knowledge of transport
src-tauri/src/magus.rs        the command enum and the argv mapping
```

Views depend on `MagusClient`, never on Tauri. Swapping transport is one file.

## Shipping

AppImage — `"targets": ["appimage"]`. Two things to remember:

- **Build in an Ubuntu 22.04 / Debian 12 container.** Tauri v2 needs
  `libwebkit2gtk-4.1-dev`, and the glibc floor is set by the build machine.
- **Do not run it as an AppImage on SteamOS.** SteamOS ships fuse3; AppImages
  `dlopen` `libfuse.so.2`, and installing fuse2 does not survive an atomic
  update. Extract once at install time into `~/.local/magus-gui.app/` and
  symlink `AppRun` into `~/.local/bin` — the same shape `kittyStep` already
  uses, so it becomes an ordinary reconciler step with no FUSE involved.

## Not built yet

One view: status. The wizard, the converge progress view and uninstall are next.
Progress needs `magus reconcile` to stream events — the current contract is a
single document emitted at the end, so a long run would sit silent.
