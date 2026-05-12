---
title: Spotify
category: install
order: 66
group: daily
summary: Music. Add as non-Steam game for Game Mode.
icon: simple-icons:spotify
commands:
  - run: flatpak install -y flathub com.spotify.Client
    description: Installs Spotify from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.spotify.Client
upstream:
  name: Spotify
  url: https://www.spotify.com
supported_devices: [any]
tags: [music]
---
