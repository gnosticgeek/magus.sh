---
title: GeForce Now
category: install
order: 28
group: streaming
summary: Cloud-stream your Steam, Epic, and Ubisoft library. RTX 4080 tier on a handheld.
icon: lucide:cloud
commands:
  - run: flatpak install -y flathub io.github.hmlendea.geforcenow-electron
    description: Installs the community Electron wrapper from Flathub. Sign in with your GeForce Now account and add it as a non-Steam game for Game Mode launch. Free tier works without a subscription.
idempotent: true
reversible: true
undo: flatpak uninstall -y io.github.hmlendea.geforcenow-electron
upstream:
  name: hmlendea/geforcenow-electron
  url: https://github.com/hmlendea/geforcenow-electron
supported_devices: [any]
tags: [streaming, cloud-gaming, nvidia]
---
