---
title: PPSSPP
category: gaming
order: 112
group: retro
summary: PSP emulator with high resolution rendering and save-state support.
commands:
  - run: flatpak install -y flathub org.ppsspp.PPSSPP
    description: Installs PPSSPP from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.ppsspp.PPSSPP
upstream:
  name: PPSSPP
  url: https://www.ppsspp.org
supported_devices: [any]
tags: [emulation, psp, retro]
---
