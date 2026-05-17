---
title: itch
category: gaming
order: 295
group: launchers
summary: Client for itch.io indie games, jams, tools, and prototypes.
commands:
  - run: flatpak install -y flathub io.itch.itch
    description: Installs the itch app from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y io.itch.itch
upstream:
  name: itch
  url: https://itch.io/app
supported_devices: [any]
tags: [launcher, itch, indie]
---
