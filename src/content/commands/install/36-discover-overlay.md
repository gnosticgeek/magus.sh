---
title: Discover Overlay
category: install
order: 36
group: streaming
summary: A Discord voice + chat overlay that actually works in Steam games.
icon: simple-icons:discord
commands:
  - run: flatpak install -y flathub io.github.trigg.Discover_overlay
    description: Installs Discover Overlay from Flathub. Pair with Vesktop or the regular Discord client — the overlay reads from either over IPC.
idempotent: true
reversible: true
undo: flatpak uninstall -y io.github.trigg.Discover_overlay
upstream:
  name: Discover Overlay
  url: https://github.com/trigg/Discover
supported_devices: [any]
tags: [discord, overlay, voice]
---
