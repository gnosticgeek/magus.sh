---
title: shadPS4
category: gaming
order: 108
group: retro
summary: Early PlayStation 4 emulator for compatible titles and homebrew.
commands:
  - run: flatpak install -y flathub net.shadps4.shadPS4
    description: Installs the shadPS4 Flatpak.
idempotent: true
reversible: true
undo: flatpak uninstall -y net.shadps4.shadPS4
upstream:
  name: shadPS4
  url: https://shadps4.net
supported_devices: [any]
danger: medium
tags: [emulation, ps4, experimental]
---
