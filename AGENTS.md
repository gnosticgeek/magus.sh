# Magus contributor guide

This is the shared repository contract for coding agents and contributors.
`CLAUDE.md` points here so tool-specific instructions do not drift.

## Repository map

- `magus/` contains the Go CLI, terminal interfaces, and reconciler.
- `src/` contains the Astro website and content catalogue.
- `gui/` contains the desktop GUI prototype.
- `TUI_SPECIFICATION.md` is the historical and visual TUI contract.
- `magus/MAC.md` documents Mac behavior, safety boundaries, and validation.

## Mac TUI architecture

The Mac TUI follows Bubble Tea's model/update/view boundary:

- `mac_tui.go` owns the model, catalogue projection, selections, and lifecycle.
- `mac_state.go` owns typed screen and session-event state.
- `mac_update.go` is the only place asynchronous messages publish model state.
- `mac_key.go` owns keyboard routing and user-intent transitions.
- `mac_view.go` renders state and must not start I/O or mutations.
- `mac_inventory.go` performs read-only package and preference inspection through
  a cancelable, bounded package-probe pool; preference reads stay sequential.
- `mac_bootstrap.go` owns the Homebrew temporary-file and lock lifecycle.
- `mac_async.go` supplies generation tokens that reject superseded results.
- `mac_execution.go` owns installation sessions and machine mutations.

Do not replace `macScreen`, `macEventKind`, or generation tokens with free-form
strings. New asynchronous subsystems must identify their result and reject it
after cancellation, replacement, or navigation out of scope.

## Safety and behavior

- Preview and dry-run modes must not write, install, refresh metadata, or launch
  external setup pages.
- Keep review and confirmation separate from execution.
- Restore only content Magus can prove it created or backed up.
- External applications and packages remain outside Magus ownership.
- Bound subprocesses with contexts and operation-specific timeouts.
- A valid late result is still rejected unless it belongs to the current run.

## Verification

Run from `magus/` after Go changes:

```sh
gofmt -w *.go
go test ./...
go test -race ./...
go vet ./...
```

TUI coverage must include narrow and wide terminals, typed screen capabilities,
cancellation, late messages, filtering, selection identity, and preview safety.
Use temporary homes and fake executables, never the developer's Homebrew state
or preferences. Preserve unrelated changes. Update `magus/README.md`,
`magus/MAC.md`, and this contract when ownership or public behavior changes.
