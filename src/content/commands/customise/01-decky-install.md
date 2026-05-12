---
title: Install Decky Loader
category: customise
order: 1
group: decky
summary: The plugin runtime. Same script as Setup — safe to re-run, idempotent.
icon: lucide:puzzle
commands:
  - run: curl -L https://github.com/SteamDeckHomebrew/decky-loader/raw/main/dist/install_release.sh | sh
    description: Official Decky installer. Adds the desktop entry and integrates into the Steam UI on next reboot. Plugins below extract into ~/homebrew/plugins/.
idempotent: true
reversible: true
undo: ~/homebrew/services/uninstall.sh
upstream:
  name: SteamDeckHomebrew/decky-loader
  url: https://github.com/SteamDeckHomebrew/decky-loader
supported_devices: [deck, steam-machine, any]
danger: low
tags: [plugins, decky, prerequisite]
---
