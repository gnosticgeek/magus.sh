---
title: xemu
category: gaming
order: 124
group: retro
summary: Original Xbox emulator with local controller and resolution scaling support.
commands:
  - run: flatpak install -y flathub app.xemu.xemu
    description: Installs xemu from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y app.xemu.xemu
upstream:
  name: xemu
  url: https://xemu.app
supported_devices: [any]
tags: [emulation, xbox, retro]
---
