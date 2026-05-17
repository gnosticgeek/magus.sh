---
title: Greenlight
category: gaming
order: 360
group: streaming
summary: Open-source Xbox Cloud and Xbox home streaming client.
commands:
  - run: flatpak install -y flathub io.github.unknownskl.greenlight
    description: Installs Greenlight from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y io.github.unknownskl.greenlight
upstream:
  name: unknownskl/greenlight
  url: https://github.com/unknownskl/greenlight
supported_devices: [any]
tags: [streaming, xbox, cloud]
---
