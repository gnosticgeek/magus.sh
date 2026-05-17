---
title: Syncthing GTK
category: gaming
order: 420
group: tools
summary: Desktop GUI for Syncthing continuous file sync between devices.
commands:
  - run: flatpak install -y flathub me.kozec.syncthingtk
    description: Installs Syncthing GTK from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y me.kozec.syncthingtk
upstream:
  name: Syncthing GTK
  url: https://github.com/syncthing/syncthing-gtk
supported_devices: [any]
tags: [sync, saves, files]
---
