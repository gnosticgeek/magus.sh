---
title: Prism Launcher
category: gaming
order: 290
group: launchers
summary: Minecraft launcher with modpack and instance management.
commands:
  - run: flatpak install -y flathub org.prismlauncher.PrismLauncher
    description: Installs Prism Launcher from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.prismlauncher.PrismLauncher
upstream:
  name: Prism Launcher
  url: https://prismlauncher.org
supported_devices: [any]
tags: [launcher, minecraft, mods]
---
