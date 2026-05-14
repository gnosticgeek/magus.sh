---
title: Dolphin Emulator
category: gaming
order: 110
group: retro
summary: GameCube and Wii. Buttery-smooth on the Deck.
icon: lucide:gamepad-2
commands:
  - run: flatpak install -y flathub org.DolphinEmu.dolphin-emu
    description: Installs Dolphin from Flathub. Pair with a controller profile from the community for Wiimote ergonomics on the touchpads.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.DolphinEmu.dolphin-emu
upstream:
  name: Dolphin Emulator
  url: https://dolphin-emu.org
supported_devices: [any]
tags: [emulator, gamecube, wii]
---
