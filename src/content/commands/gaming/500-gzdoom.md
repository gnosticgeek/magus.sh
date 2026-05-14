---
title: GZDoom
category: gaming
order: 500
group: ports
summary: Modern source port of Doom, Doom II, Heretic, Hexen, Strife. Point it at your WADs and play.
icon: lucide:flame
commands:
  - run: flatpak install -y flathub org.zdoom.GZDoom
    description: Installs GZDoom from Flathub. Drop your WAD/PK3 files in ~/.var/app/org.zdoom.GZDoom/.config/gzdoom/ (or load them via the menu). The shareware WADs are free; bring your own commercial WADs.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.zdoom.GZDoom
upstream:
  name: GZDoom
  url: https://zdoom.org
supported_devices: [any]
tags: [source-port, doom, fps]
---
