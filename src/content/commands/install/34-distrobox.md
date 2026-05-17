---
title: Distrobox
category: install
order: 34
group: system
summary: User-space Linux containers for development tools without unlocking SteamOS.
commands:
  - run: curl -s https://raw.githubusercontent.com/89luca89/distrobox/main/install | sh -s -- --prefix "$HOME/.local"
    description: Installs Distrobox into ~/.local using the official upstream installer. You still need a container runtime such as Podman before creating containers.
idempotent: true
reversible: true
undo: curl -s https://raw.githubusercontent.com/89luca89/distrobox/main/uninstall | sh -s -- --prefix "$HOME/.local"
upstream:
  name: 89luca89/distrobox
  url: https://github.com/89luca89/distrobox
supported_devices: [any]
danger: medium
tags: [containers, development, terminal]
---
