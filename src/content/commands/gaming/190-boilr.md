---
title: BoilR
category: gaming
order: 190
group: tools
summary: Finds games from other launchers and adds them to Steam with SteamGridDB art.
commands:
  - run: flatpak install -y flathub io.github.philipk.boilr
    description: Installs BoilR from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y io.github.philipk.boilr
upstream:
  name: PhilipK/BoilR
  url: https://github.com/PhilipK/BoilR
supported_devices: [any]
tags: [steam, launcher, artwork]
---
