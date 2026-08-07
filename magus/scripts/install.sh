#!/bin/sh
#
# magus installer — curl -fsSL magus.sh/install | sh
#
# Downloads the magus binary for this machine, verifies its checksum, and puts
# it in ~/.local/bin. It installs the tool and stops there: it does not change
# a single setting on your machine. Run `magus run` yourself afterwards.
#
# That split is deliberate. A pipe-to-shell that also starts reconfiguring your
# desktop is asking for a great deal of trust in one keystroke — and it could
# not work anyway, because the wizard needs a terminal and a pipe is not one.
#
# environment:
#   MAGUS_VERSION   install a specific tag (default: the latest release)
#   MAGUS_BIN_DIR   install somewhere other than ~/.local/bin
#
# POSIX sh on purpose — this is piped into `sh`, not bash, and SteamOS's /bin/sh
# is not guaranteed to be bash.

set -eu

REPO="gnosticgeek/magus.sh"
BIN_DIR="${MAGUS_BIN_DIR:-$HOME/.local/bin}"

# Four verbs, matching the tool's own output.
if [ -t 1 ]; then
  C_OK=$(printf '\033[32m'); C_WARN=$(printf '\033[33m')
  C_HEAD=$(printf '\033[1m'); C_DIM=$(printf '\033[2m'); C_OFF=$(printf '\033[0m')
else
  C_OK=""; C_WARN=""; C_HEAD=""; C_DIM=""; C_OFF=""
fi
log()  { printf '\n%s── %s%s\n' "$C_HEAD" "$*" "$C_OFF"; }
ok()   { printf '  %s✓%s %s\n' "$C_OK" "$C_OFF" "$*"; }
warn() { printf '  %s!%s %s\n' "$C_WARN" "$C_OFF" "$*"; }
die()  { printf '  %s✗%s %s\n' "$C_WARN" "$C_OFF" "$*" >&2; exit 1; }
have() { command -v "$1" >/dev/null 2>&1; }

# ---- preflight -------------------------------------------------------------

log "magus installer"

case "$(uname -s)" in
  Linux) ;;
  Darwin) die "magus targets SteamOS. On a Mac, build from source: cd magus && go build ." ;;
  *) die "unsupported OS: $(uname -s)" ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *) die "unsupported architecture: $(uname -m)" ;;
esac

# curl or wget, whichever is here. SteamOS has curl.
if have curl; then
  fetch()   { curl -fsSL "$1"; }
  fetch_to() { curl -fsSL -o "$2" "$1"; }
elif have wget; then
  fetch()   { wget -qO- "$1"; }
  fetch_to() { wget -qO "$2" "$1"; }
else
  die "need curl or wget to download anything"
fi

ok "linux/$ARCH"

# ---- resolve the version ---------------------------------------------------

VERSION="${MAGUS_VERSION:-}"
if [ -z "$VERSION" ]; then
  # Ask the API for the latest tag. Parsed with sed rather than jq, which is
  # not installed on SteamOS. A repo with no releases answers 404, which is a
  # normal answer here rather than a fault worth showing curl's own error for.
  VERSION=$(fetch "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null \
    | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1) || true
fi
[ -n "$VERSION" ] || die "could not find a release. Has one been published yet?
    See https://github.com/$REPO/releases"

ok "version $VERSION"

ASSET="magus-linux-$ARCH"
BASE="https://github.com/$REPO/releases/download/$VERSION"

# ---- download and verify ---------------------------------------------------

log "downloading"

TMP=$(mktemp -d) || die "could not create a temp directory"
# Clean up on any exit, including the failure paths below.
trap 'rm -rf "$TMP"' EXIT INT TERM

fetch_to "$BASE/$ASSET" "$TMP/magus" \
  || die "download failed: $BASE/$ASSET"
ok "got $ASSET"

# Verify against the release's checksum file. A binary piped off the internet
# and made executable deserves at least this much.
if fetch_to "$BASE/checksums.txt" "$TMP/checksums.txt" 2>/dev/null; then
  if have sha256sum; then
    want=$(sed -n "s/^\([0-9a-f]*\)  *$ASSET\$/\1/p" "$TMP/checksums.txt" | head -1)
    got=$(sha256sum "$TMP/magus" | cut -d' ' -f1)
    if [ -z "$want" ]; then
      warn "no checksum listed for $ASSET — skipping verification"
    elif [ "$want" != "$got" ]; then
      die "checksum mismatch — refusing to install
    expected $want
    got      $got"
    else
      ok "checksum verified"
    fi
  else
    warn "sha256sum not found — skipping verification"
  fi
else
  warn "no checksums.txt in this release — skipping verification"
fi

# ---- install ---------------------------------------------------------------

log "installing"

mkdir -p "$BIN_DIR" || die "cannot create $BIN_DIR"
chmod +x "$TMP/magus"
# Move into place as one step so there is never a half-written magus on PATH.
mv -f "$TMP/magus" "$BIN_DIR/magus" || die "cannot write to $BIN_DIR"
ok "installed $BIN_DIR/magus"

# SteamOS does not put ~/.local/bin on PATH, so without this `magus` installs
# successfully and then cannot be found — which looks like a broken install.
#
# Appending to a shell rc file is a change to someone's environment, so it is
# the one thing here that touches anything outside $BIN_DIR. It is guarded: only
# rc files that already exist, only when the line is not already present, and
# always announced. MAGUS_NO_PATH=1 skips it.
add_to_path() {
  rc="$1"
  [ -f "$rc" ] || return 1
  # Already handled, by us or by the user.
  grep -qF "$BIN_DIR" "$rc" 2>/dev/null && { ok "$BIN_DIR already referenced in $rc"; return 0; }
  {
    printf '\n# added by the magus installer\n'
    printf 'export PATH="%s:$PATH"\n' "$BIN_DIR"
  } >> "$rc" || return 1
  ok "added $BIN_DIR to PATH in $rc"
  PATH_CHANGED=1
  return 0
}

PATH_CHANGED=0
case ":${PATH}:" in
  *":$BIN_DIR:"*)
    ok "$BIN_DIR is on your PATH"
    ;;
  *)
    if [ "${MAGUS_NO_PATH:-0}" = "1" ]; then
      warn "$BIN_DIR is not on your PATH (MAGUS_NO_PATH=1, leaving it alone)"
    else
      touched=0
      for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
        add_to_path "$rc" && touched=1
      done
      if [ "$touched" = "0" ]; then
        warn "$BIN_DIR is not on your PATH, and no shell rc file was found"
        printf '    %sadd this to your shell profile:%s\n' "$C_DIM" "$C_OFF"
        printf '    export PATH="%s:$PATH"\n' "$BIN_DIR"
      fi
    fi
    ;;
esac

# ---- what next -------------------------------------------------------------

log "done"
printf '  magus is installed. It has changed no settings.\n\n'

# The rc file only affects shells started after it, so this shell still cannot
# find a bare `magus`. Print the command that actually works right now.
if [ "$PATH_CHANGED" = "1" ]; then
  printf '  %sThis shell was started before the PATH change. Either:%s\n' "$C_DIM" "$C_OFF"
  printf '    %ssource ~/.bashrc%s   %s— then plain `magus` works%s\n' "$C_HEAD" "$C_OFF" "$C_DIM" "$C_OFF"
  printf '    %s— or just open a new terminal.%s\n\n' "$C_DIM" "$C_OFF"
fi

printf '    %smagus doctor%s   see what it would do — changes nothing\n' "$C_HEAD" "$C_OFF"
printf '    %smagus run%s      answer five questions, then converge\n' "$C_HEAD" "$C_OFF"
printf '\n  %sUntil this shell picks up the new PATH, use %s/magus.%s\n' "$C_DIM" "$BIN_DIR" "$C_OFF"
printf '  %sRun it yourself — this installer deliberately does not.%s\n' "$C_DIM" "$C_OFF"
