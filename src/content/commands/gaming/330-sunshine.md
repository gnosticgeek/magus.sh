---
title: Sunshine
category: gaming
order: 330
group: streaming
summary: Self-hosted GameStream server. The Deck-side counterpart to Moonlight.
icon: lucide:sun
commands:
  - run: flatpak install -y flathub dev.lizardbyte.app.Sunshine
    description: Installs Sunshine from Flathub. Run on the host you want to stream FROM — turn your Deck (or laptop, or desktop) into a remote target.
idempotent: true
reversible: true
undo: flatpak uninstall -y dev.lizardbyte.app.Sunshine
upstream:
  name: Sunshine
  url: https://app.lizardbyte.dev/Sunshine
supported_devices: [steam-machine, any]
tags: [streaming, host, gamestream]
---
