---
title: Install Decky Loader
category: setup
order: 20
summary: Plugin platform. Powers PowerTools, CSS Loader, SteamGridDB, and dozens more.
icon: lucide:puzzle
commands:
  - run: curl -L https://github.com/SteamDeckHomebrew/decky-loader/raw/main/dist/install_release.sh | sh
    description: Official Decky Loader install script. Adds a desktop entry and integrates into the Steam UI on next reboot.
idempotent: true
reversible: true
undo: ~/homebrew/services/uninstall.sh
upstream:
  name: SteamDeckHomebrew/decky-loader
  url: https://github.com/SteamDeckHomebrew/decky-loader
supported_devices: [deck, steam-machine, any]
danger: low
tags: [plugins, decky]
---
