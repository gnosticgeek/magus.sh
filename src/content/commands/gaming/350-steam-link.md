---
title: Steam Link
category: gaming
order: 350
group: streaming
summary: Stream games from another computer running Steam on your local network.
commands:
  - run: flatpak install -y flathub com.valvesoftware.SteamLink
    description: Installs Steam Link from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.valvesoftware.SteamLink
upstream:
  name: Steam Link
  url: https://store.steampowered.com/steamlink/about/
supported_devices: [any]
tags: [streaming, remote-play, steam]
---
