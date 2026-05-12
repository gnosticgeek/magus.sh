---
title: Ludusavi
category: install
order: 48
group: system
summary: Save game backup. Pair with cloud sync for resilience.
icon: lucide:save
commands:
  - run: flatpak install -y flathub com.github.mtkennerly.ludusavi
    description: Installs Ludusavi from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.github.mtkennerly.ludusavi
upstream:
  name: Ludusavi
  url: https://github.com/mtkennerly/ludusavi
supported_devices: [any]
tags: [saves, backup]
---
