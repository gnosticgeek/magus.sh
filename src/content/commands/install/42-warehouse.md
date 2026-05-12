---
title: Warehouse
category: install
order: 42
group: system
summary: Flatpak management GUI. Mass-uninstall, audit user data, snapshot installs.
icon: lucide:warehouse
commands:
  - run: flatpak install -y flathub io.github.flattool.Warehouse
    description: Installs Warehouse from Flathub. Pairs nicely with Flatseal — Flatseal does permissions, Warehouse does lifecycle.
idempotent: true
reversible: true
undo: flatpak uninstall -y io.github.flattool.Warehouse
upstream:
  name: Warehouse
  url: https://github.com/flattool/warehouse
supported_devices: [any]
tags: [flatpak, manager, hygiene]
---
