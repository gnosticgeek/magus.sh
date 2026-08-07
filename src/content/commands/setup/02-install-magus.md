---
title: Install magus
category: setup
order: 20
summary: One 4 MB binary into ~/.local/bin. Nothing to install alongside it, and it survives SteamOS updates because it lives entirely in $HOME.
icon: lucide:package
commands:
  - run: curl -fsSL https://magus.sh/install | sh
    description: Downloads the binary for this machine's architecture, verifies its SHA-256 against the release checksums, and installs it to ~/.local/bin. It changes nothing else — run `magus doctor` (which is read-only) or `magus run` yourself afterwards.
idempotent: true
reversible: true
undo: rm -f ~/.local/bin/magus
upstream:
  name: gnosticgeek/magus.sh releases
  url: https://github.com/gnosticgeek/magus.sh/releases
supported_devices: [deck, steam-machine, any]
danger: low
tags: [setup, magus, install]
---
