#!/bin/sh
#
# magus installer — curl -fsSL magus.sh/install | sh
#
# Downloads and verifies Magus. On a Mac with an interactive terminal, opens
# the menu; applications/settings change only after review inside the TUI.
# Linux retains its separate `magus run` launch.
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
  Linux) OS="linux" ;;
  Darwin) OS="darwin" ;;
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

ok "$OS/$ARCH"

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

ASSET="magus-$OS-$ARCH"
BASE="https://github.com/$REPO/releases/download/$VERSION"

# ---- download and verify ---------------------------------------------------

log "downloading"

TMP=$(mktemp -d) || die "could not create a temp directory"
# Clean up on any exit, including the failure paths below.
STAGED=""
trap 'rm -rf "$TMP"; [ -z "$STAGED" ] || rm -f "$STAGED"' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

fetch_to "$BASE/$ASSET" "$TMP/magus" \
  || die "download failed: $BASE/$ASSET"
ok "got $ASSET"

# Verify against the release's checksum file. A binary piped off the internet
# and made executable deserves at least this much.
fetch_to "$BASE/checksums.txt" "$TMP/checksums.txt" \
  || die "cannot download checksums — refusing an unverified binary"
want=$(awk -v asset="$ASSET" '$2 == asset { print $1 }' "$TMP/checksums.txt")
[ "${#want}" = 64 ] || die "missing or invalid checksum for $ASSET"
case "$want" in *[!0-9a-f]*) die "invalid checksum for $ASSET" ;; esac
if have sha256sum; then
  got=$(sha256sum "$TMP/magus" | cut -d' ' -f1)
elif have shasum; then
  got=$(shasum -a 256 "$TMP/magus" | cut -d' ' -f1)
else
  die "need sha256sum or shasum to verify this download"
fi
[ "$want" = "$got" ] || die "checksum mismatch — refusing to install"
ok "checksum verified"

# ---- install ---------------------------------------------------------------

log "installing"

mkdir -p "$BIN_DIR" || die "cannot create $BIN_DIR"
STAGED=$(mktemp "$BIN_DIR/.magus.XXXXXX") || die "cannot stage binary"
cp "$TMP/magus" "$STAGED" || die "cannot copy binary"
chmod 755 "$STAGED"
# Stage on the destination filesystem so the final rename is atomic.
mv -f "$STAGED" "$BIN_DIR/magus" || die "cannot write to $BIN_DIR"
STAGED=""
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
    if [ "$OS" = "darwin" ] || [ "${MAGUS_NO_PATH:-0}" = "1" ]; then
      warn "$BIN_DIR is not on your PATH; launch with $BIN_DIR/magus"
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

# ---- launch ---------------------------------------------------------------

log "done"
printf '  Open the setup menu: %s/magus run\n' "$BIN_DIR"
if [ "$OS" = "darwin" ] && [ "${MAGUS_NO_LAUNCH:-0}" != "1" ] && [ -t 1 ]; then
  # Reattach stdin: the installer itself arrived over a pipe.
  if ( : < /dev/tty ) 2>/dev/null; then
    "$BIN_DIR/magus" run < /dev/tty
  fi
fi
