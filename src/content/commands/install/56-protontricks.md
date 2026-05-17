---
title: Protontricks
category: install
order: 56
group: system
summary: Winetricks-style helper for changing individual Steam Proton prefixes.
commands:
  - run: flatpak install -y flathub com.github.Matoking.protontricks
    description: Installs the Protontricks Flatpak. Grant extra library access with Flatseal if your Steam games live outside the default path.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.github.Matoking.protontricks
upstream:
  name: Protontricks
  url: https://github.com/Matoking/protontricks
supported_devices: [any]
tags: [proton, wine, steam]
---
