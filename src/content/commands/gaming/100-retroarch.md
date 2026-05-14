---
title: RetroArch
category: gaming
order: 100
group: retro
summary: The big-tent emulator frontend. Cores for nearly every retro console.
icon: simple-icons:retroarch
commands:
  - run: flatpak install -y flathub org.libretro.RetroArch
    description: Installs RetroArch from Flathub. EmuDeck users can skip this — but a vanilla RetroArch is lighter and easier to reason about.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.libretro.RetroArch
upstream:
  name: RetroArch
  url: https://www.retroarch.com
supported_devices: [any]
tags: [emulator, retro, libretro]
---
