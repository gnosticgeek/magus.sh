---
title: ProtonUp-Qt
category: install
order: 10
group: launchers
summary: Proton-GE, Wine-GE, Luxtorpeda installer.
icon: lucide:wine
commands:
  - run: flatpak install -y flathub net.davidotek.pupgui2
    description: Installs ProtonUp-Qt from Flathub. Launch from desktop mode after install to fetch the latest Proton-GE.
idempotent: true
reversible: true
undo: flatpak uninstall -y net.davidotek.pupgui2
upstream:
  name: ProtonUp-Qt
  url: https://github.com/DavidoTek/ProtonUp-Qt
supported_devices: [any]
tags: [proton, wine, compatibility]
---
