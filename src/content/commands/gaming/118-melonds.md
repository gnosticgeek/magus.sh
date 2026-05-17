---
title: melonDS
category: gaming
order: 118
group: retro
summary: Nintendo DS emulator with local multiplayer and accuracy-focused emulation.
commands:
  - run: flatpak install -y flathub net.kuribo64.melonDS
    description: Installs melonDS from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y net.kuribo64.melonDS
upstream:
  name: melonDS
  url: https://melonds.kuribo64.net
supported_devices: [any]
tags: [emulation, ds, retro]
---
