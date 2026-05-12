---
title: DuckStation
category: retro
order: 30
summary: PS1 emulator. Per-game compatibility, sharp upscaling.
icon: lucide:joystick
commands:
  - run: flatpak install -y flathub org.duckstation.DuckStation
    description: Installs DuckStation from Flathub. The cleanest PS1 experience on Linux right now — fast cores, sane defaults, gorgeous output.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.duckstation.DuckStation
upstream:
  name: DuckStation
  url: https://www.duckstation.org
supported_devices: [any]
tags: [emulator, ps1]
---
