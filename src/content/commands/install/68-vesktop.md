---
title: Vesktop
category: install
order: 68
group: daily
summary: Lightweight Discord with Vencord baked in. Better screen-share on Linux.
icon: simple-icons:discord
commands:
  - run: flatpak install -y flathub dev.vencord.Vesktop
    description: Installs Vesktop from Flathub. Replaces the official Discord client with one that screen-shares with audio on Wayland and respects your battery.
idempotent: true
reversible: true
undo: flatpak uninstall -y dev.vencord.Vesktop
upstream:
  name: Vesktop
  url: https://github.com/Vencord/Vesktop
supported_devices: [any]
tags: [discord, voice, chat]
---
