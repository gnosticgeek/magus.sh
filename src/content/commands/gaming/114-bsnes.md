---
title: bsnes
category: gaming
order: 114
group: retro
summary: Accurate SNES emulator with HD Mode 7 and strong shader support.
commands:
  - run: flatpak install -y flathub dev.bsnes.bsnes
    description: Installs bsnes from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y dev.bsnes.bsnes
upstream:
  name: bsnes
  url: https://github.com/bsnes-emu/bsnes
supported_devices: [any]
tags: [emulation, snes, retro]
---
