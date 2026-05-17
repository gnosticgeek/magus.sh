---
title: DeSmuME
category: gaming
order: 116
group: retro
summary: Nintendo DS emulator with a long compatibility history.
commands:
  - run: flatpak install -y flathub org.desmume.DeSmuME
    description: Installs DeSmuME from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.desmume.DeSmuME
upstream:
  name: DeSmuME
  url: http://desmume.org
supported_devices: [any]
tags: [emulation, ds, retro]
---
