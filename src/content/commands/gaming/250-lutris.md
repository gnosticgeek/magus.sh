---
title: Lutris
category: gaming
order: 250
group: launchers
summary: Wine, emulators, Battle.net, Origin, Ubisoft Connect.
icon: simple-icons:lutris
commands:
  - run: flatpak install -y flathub net.lutris.Lutris
    description: Installs Lutris from Flathub. Pair with ProtonUp-Qt's Wine-GE builds for best compatibility.
idempotent: true
reversible: true
undo: flatpak uninstall -y net.lutris.Lutris
upstream:
  name: Lutris
  url: https://lutris.net
supported_devices: [any]
tags: [launcher, wine, battlenet]
---
