#!/usr/bin/env bash
#
# sandbox.sh — exercise the magus reconciler against a throwaway $HOME.
#
# Everything magus writes lives under $HOME, so overriding it is a complete
# sandbox: nothing here can touch your real dotfiles. Safe to run on a laptop.
#
# usage:
#   scripts/sandbox.sh                    dry run only — writes nothing, downloads nothing
#   scripts/sandbox.sh --real             actually install; run the full drift/repair loop
#   scripts/sandbox.sh --real --bundles   include the app bundles (pulls ~1GB of runtimes)
#   scripts/sandbox.sh --real --stub-kitty  skip the kitty download; use a stub binary
#   scripts/sandbox.sh --device steam-deck   exercise a different device branch
#   scripts/sandbox.sh --keep             leave the sandbox behind for poking at
#   scripts/sandbox.sh --clean            delete the sandbox and exit
#
# --real is Linux-only. kitty's installer targets /Applications on macOS, which
# is outside the sandbox — use --stub-kitty there, or just take the dry run.
#
# The interesting part is the drift/repair loop under --real: it installs, then
# deletes an artifact the way a SteamOS atomic update would, then proves that
# doctor sees the drift and one reconcile repairs it.

# Deliberately no -e: doctor exits 1 when it finds drift, and that is a pass
# here, not a failure. Exit codes are checked explicitly instead.
set -uo pipefail

SANDBOX="${MAGUS_SANDBOX:-${TMPDIR:-/tmp}/magus-sandbox}"
SANDBOX="${SANDBOX%/}"
DEVICE="steam-machine"
REAL=0
BUNDLES=0
KEEP=0
STUB_KITTY=0

# Four log verbs and nothing else, matching the installer convention in the brief.
if [ -t 1 ]; then
  C_OK=$'\033[32m'; C_WARN=$'\033[33m'; C_HEAD=$'\033[1m'; C_DIM=$'\033[2m'; C_OFF=$'\033[0m'
else
  C_OK=""; C_WARN=""; C_HEAD=""; C_DIM=""; C_OFF=""
fi
log()  { printf '\n%s══ %s%s\n' "$C_HEAD" "$*" "$C_OFF"; }
ok()   { printf '%s  ✓%s %s\n' "$C_OK" "$C_OFF" "$*"; }
warn() { printf '%s  !%s %s\n' "$C_WARN" "$C_OFF" "$*"; }
die()  { printf '%s  ✗%s %s\n' "$C_WARN" "$C_OFF" "$*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

while [ $# -gt 0 ]; do
  case "$1" in
    --real)        REAL=1 ;;
    --bundles)     BUNDLES=1 ;;
    --keep)        KEEP=1 ;;
    --stub-kitty)  STUB_KITTY=1 ;;
    --device)  shift; DEVICE="${1:-}"; [ -n "$DEVICE" ] || die "--device needs a value" ;;
    --clean)   rm -rf "$SANDBOX"; ok "removed $SANDBOX"; exit 0 ;;
    -h|--help)
      # Generated from the header comment so it cannot drift from the real flags.
      awk 'NR>=3{ if($0 ~ /^#/){sub(/^# ?/,"");print} else exit }' "${BASH_SOURCE[0]}"; exit 0 ;;
    *) die "unknown flag: $1" ;;
  esac
  shift
done

# Preflight dies early, before anything is built or downloaded.
# kitty's installer targets /Applications on macOS — outside the sandbox, and
# not something a test script gets to decide for you.
if [ "$REAL" = 1 ] && [ "$STUB_KITTY" = 0 ] && [ "$(uname -s)" != "Linux" ]; then
  die "--real is Linux-only (kitty's installer would write to /Applications here). Use --stub-kitty."
fi

cd "$(dirname "${BASH_SOURCE[0]}")/.." || die "cannot find the magus directory"

have go || die "go is required to build the binary"

log "building"
BIN="$SANDBOX/magus-bin"
rm -rf "$SANDBOX"
mkdir -p "$SANDBOX" || die "cannot create $SANDBOX"
go build -o "$BIN" . || die "build failed"
ok "built $BIN"

HOMEDIR="$SANDBOX/home"
mkdir -p "$HOMEDIR"

# magus() runs the binary against the sandbox home. XDG_* are cleared so the
# layout derives from the sandbox home alone — a developer with XDG_CONFIG_HOME
# set would otherwise have magus write outside the sandbox.
magus() {
  env -u XDG_CONFIG_HOME -u XDG_DATA_HOME -u XDG_STATE_HOME \
    HOME="$HOMEDIR" MAGUS_DEVICE="$DEVICE" "$BIN" "$@"
}

MANIFEST="$HOMEDIR/.config/magus/manifest.toml"

log "environment"
ok "sandbox home  $HOMEDIR"
ok "device        $DEVICE (forced via MAGUS_DEVICE)"
if have flatpak; then
  ok "flatpak       present — the flatpak steps will really run"
else
  warn "flatpak       missing — every flatpak step will report n/a"
fi

log "1. dry run — writes nothing, downloads nothing"
magus run --defaults --dry-run
[ $? -eq 0 ] || die "dry run failed"

if [ "$REAL" = 0 ]; then
  log "done"
  ok "dry run only. Re-run with --real to install and exercise the drift/repair loop."
  [ "$KEEP" = 1 ] && ok "sandbox kept at $SANDBOX" || rm -rf "$SANDBOX"
  exit 0
fi

if [ "$STUB_KITTY" = 1 ]; then
  # Stand in for the upstream installer so the step skips the download and only
  # does its own work: symlinks and the launcher entry.
  mkdir -p "$HOMEDIR/.local/kitty.app/bin"
  printf '#!/bin/sh\necho stub kitty\n' > "$HOMEDIR/.local/kitty.app/bin/kitty"
  printf '#!/bin/sh\necho stub kitten\n' > "$HOMEDIR/.local/kitty.app/bin/kitten"
  chmod +x "$HOMEDIR/.local/kitty.app/bin/kitty" "$HOMEDIR/.local/kitty.app/bin/kitten"
  ok "seeded a stub kitty — the installer download will be skipped"
fi

log "2. real run"
if [ "$BUNDLES" = 0 ]; then
  # Write the manifest first, strip the bundles, then converge — so a laptop
  # test does not pull a gigabyte of KDE runtime unless it was asked to.
  magus run --defaults --dry-run >/dev/null 2>&1
  mkdir -p "$(dirname "$MANIFEST")"
  cat > "$MANIFEST" <<EOF
[magus]
version = "0.2.0"
device = "$DEVICE"

[choices]
terminal = "kitty"
browser = "firefox"
bundles = []
theme = "tokyo-night"

[optimisations]
hdmi_full_range = true
cec = true
power_profile = "performance"
EOF
  ok "wrote a no-bundles manifest (use --bundles for the full set)"
  magus reconcile
else
  magus run --defaults
fi
[ $? -eq 0 ] || warn "the run reported failures — read the output above"

log "3. re-run — the normal case, should change nothing"
magus reconcile 2>&1 | tail -4

log "4. simulating an atomic update eating an artifact"
rm -f "$HOMEDIR/.local/bin/kitten" "$HOMEDIR/.local/share/applications/kitty.desktop"
ok "deleted ~/.local/bin/kitten and the kitty launcher entry"

log "5. doctor should now report drift and exit 1"
magus doctor
code=$?
if [ "$code" -eq 1 ]; then
  ok "doctor exited 1 as expected"
else
  warn "doctor exited $code — expected 1 once an artifact is missing"
fi

log "6. one reconcile repairs it"
magus reconcile 2>&1 | tail -4

log "7. doctor should be clean and exit 0"
magus doctor >/dev/null 2>&1
code=$?
if [ "$code" -eq 0 ]; then
  ok "doctor exited 0 — the machine converged back to the manifest"
else
  warn "doctor exited $code — the repair did not fully converge"
fi

log "8. uninstall"
magus uninstall 2>&1 | tail -6

log "done"
if [ "$KEEP" = 1 ]; then
  ok "sandbox kept at $SANDBOX"
  ok "poke at it:  env HOME=$HOMEDIR MAGUS_DEVICE=$DEVICE $BIN doctor"
else
  rm -rf "$SANDBOX"
  ok "sandbox removed (use --keep to inspect it)"
fi
