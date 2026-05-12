---
title: Flatseal
category: install
order: 40
group: system
summary: Flatpak permissions GUI.
icon: lucide:shield-check
commands:
  - run: flatpak install -y flathub com.github.tchx84.Flatseal
    description: Installs Flatseal from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.github.tchx84.Flatseal
upstream:
  name: Flatseal
  url: https://github.com/tchx84/Flatseal
supported_devices: [any]
tags: [flatpak, permissions, hygiene]
---
