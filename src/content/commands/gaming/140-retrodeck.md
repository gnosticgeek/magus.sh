---
title: RetroDeck
category: gaming
order: 140
group: retro
summary: Every major emulator in a single flatpak. A drop-in EmuDeck alternative.
icon: lucide:package
commands:
  - run: flatpak install -y flathub net.retrodeck.retrodeck
    description: Installs RetroDeck from Flathub. Everything sandboxed in one app — ROMs live in a single folder, updates ship as one package, no host-side mess.
idempotent: true
reversible: true
undo: flatpak uninstall -y net.retrodeck.retrodeck
upstream:
  name: RetroDeck
  url: https://retrodeck.net
supported_devices: [any]
tags: [emulator, retro, all-in-one, flatpak]
---
