---
title: Steam ROM Manager
category: gaming
order: 106
group: retro
summary: Adds ROMs and emulator entries to Steam with artwork and collections.
commands:
  - run: flatpak install -y flathub com.steamgriddb.steam-rom-manager
    description: Installs Steam ROM Manager from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.steamgriddb.steam-rom-manager
upstream:
  name: Steam ROM Manager
  url: https://steamgriddb.github.io/steam-rom-manager
supported_devices: [any]
tags: [emulation, steam, artwork]
---
