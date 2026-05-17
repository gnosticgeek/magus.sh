---
title: Parsec
category: gaming
order: 370
group: streaming
summary: Low-latency remote desktop and co-op streaming client.
commands:
  - run: flatpak install -y flathub com.parsecgaming.parsec
    description: Installs Parsec from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.parsecgaming.parsec
upstream:
  name: Parsec
  url: https://parsec.app
supported_devices: [any]
tags: [streaming, remote-play, co-op]
---
