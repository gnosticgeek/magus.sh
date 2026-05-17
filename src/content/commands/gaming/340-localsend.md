---
title: LocalSend
category: gaming
order: 340
group: streaming
summary: Share files with nearby devices over your local network, no cloud account needed.
commands:
  - run: flatpak install -y flathub org.localsend.localsend_app
    description: Installs LocalSend from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.localsend.localsend_app
upstream:
  name: LocalSend
  url: https://localsend.org
supported_devices: [any]
tags: [transfer, local-network, files]
---
