---
title: Flycast
category: gaming
order: 128
group: retro
summary: Dreamcast, Naomi, and Atomiswave emulator.
commands:
  - run: flatpak install -y flathub org.flycast.Flycast
    description: Installs Flycast from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.flycast.Flycast
upstream:
  name: Flycast
  url: https://flyinghead.github.io/flycast-builds/
supported_devices: [any]
tags: [emulation, dreamcast, arcade]
---
