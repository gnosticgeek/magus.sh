---
title: Minigalaxy
category: gaming
order: 280
group: launchers
summary: Lightweight client for downloading and playing GOG Linux games.
commands:
  - run: flatpak install -y flathub io.github.sharkwouter.Minigalaxy
    description: Installs Minigalaxy from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y io.github.sharkwouter.Minigalaxy
upstream:
  name: Minigalaxy
  url: https://github.com/sharkwouter/minigalaxy
supported_devices: [any]
tags: [launcher, gog, linux]
---
