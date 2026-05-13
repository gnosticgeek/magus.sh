---
title: GOverlay
category: install
order: 46
group: system
summary: GUI for MangoHud — FPS, frametime, sensors. Skip the config file edit.
icon: lucide:gauge
commands:
  - run: flatpak install -y flathub io.github.benjamimgois.goverlay
    description: Installs GOverlay from Flathub. The Deck ships MangoHud anyway — this just gives it a sane UI instead of MANGOHUD_CONFIG.
idempotent: true
reversible: true
undo: flatpak uninstall -y io.github.benjamimgois.goverlay
upstream:
  name: GOverlay
  url: https://github.com/benjamimgois/goverlay
supported_devices: [any]
tags: [mangohud, overlay, performance]
---
