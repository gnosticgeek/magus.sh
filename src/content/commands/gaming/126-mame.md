---
title: MAME
category: gaming
order: 126
group: retro
summary: Arcade and computer preservation emulator for MAME ROM sets.
commands:
  - run: flatpak install -y flathub org.mamedev.MAME
    description: Installs MAME from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.mamedev.MAME
upstream:
  name: MAME
  url: https://www.mamedev.org
supported_devices: [any]
tags: [emulation, arcade, retro]
---
