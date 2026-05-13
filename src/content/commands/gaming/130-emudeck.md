---
title: EmuDeck
category: retro
order: 40
summary: One-shot installer that sets up a curated emulator stack with sane defaults.
icon: lucide:wand-2
commands:
  - run: curl -L https://www.emudeck.com/EmuDeck.sh -o /tmp/EmuDeck.sh && chmod +x /tmp/EmuDeck.sh && /tmp/EmuDeck.sh
    description: Downloads and launches the EmuDeck installer. Walks you through ROM paths, themes, and per-system tweaks — fastest route to a complete retro setup if you'd rather not pick emulators à la carte.
idempotent: true
reversible: false
upstream:
  name: EmuDeck
  url: https://www.emudeck.com
supported_devices: [any]
tags: [emulator, retro, installer, all-in-one]
---
