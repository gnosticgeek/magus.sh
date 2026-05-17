---
title: Cemu
category: gaming
order: 122
group: retro
summary: Wii U emulator with broad compatibility and graphics pack support.
commands:
  - run: flatpak install -y flathub info.cemu.Cemu
    description: Installs Cemu from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y info.cemu.Cemu
upstream:
  name: Cemu
  url: https://cemu.info
supported_devices: [any]
tags: [emulation, wii-u, retro]
---
