---
title: Cartridges
category: gaming
order: 230
group: launchers
summary: One library that reads from Steam, Heroic, Lutris, emulators.
icon: lucide:library
commands:
  - run: flatpak install -y flathub page.kramo.Cartridges
    description: Installs Cartridges from Flathub. Auto-imports games from every launcher you've installed — handy when your library lives in five places.
idempotent: true
reversible: true
undo: flatpak uninstall -y page.kramo.Cartridges
upstream:
  name: Cartridges
  url: https://github.com/kra-mo/cartridges
supported_devices: [any]
tags: [library, launcher]
---
