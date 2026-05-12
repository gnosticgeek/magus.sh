---
title: Install CryoUtilities
category: optimise
order: 10
summary: Swap, swappiness, huge pages — one tool. Click "Recommended" after install.
icon: lucide:gauge
commands:
  - run: wget -O - https://raw.githubusercontent.com/CryoByte33/steam-deck-utilities/main/install-cryo-utilities.sh | bash
    description: Downloads and runs the official installer. After install, launch CryoUtilities from desktop mode and apply Recommended Settings.
idempotent: true
reversible: true
undo: rm -rf ~/.local/share/cryo_utilities ~/Desktop/CryoUtilities.desktop
upstream:
  name: CryoByte33/steam-deck-utilities
  url: https://github.com/CryoByte33/steam-deck-utilities
supported_devices: [deck]
deck_only: true
danger: low
tags: [perf, swap, swappiness, huge-pages]
---
