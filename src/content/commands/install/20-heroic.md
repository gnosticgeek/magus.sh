---
title: Heroic Games Launcher
category: install
order: 12
group: launchers
summary: Epic, GOG, Amazon Prime in one launcher.
icon: simple-icons:heroicgameslauncher
commands:
  - run: flatpak install -y flathub com.heroicgameslauncher.hgl
    description: Installs Heroic from Flathub. After install, add it as a non-Steam game so it shows in Game Mode.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.heroicgameslauncher.hgl
upstream:
  name: Heroic Games Launcher
  url: https://heroicgameslauncher.com
supported_devices: [any]
tags: [launcher, epic, gog]
---
