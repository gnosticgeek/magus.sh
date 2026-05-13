---
title: Moonlight
category: install
order: 30
group: streaming
summary: Stream games from a GPU-equipped PC to your Deck. Open-source NVIDIA GameStream client.
icon: lucide:moon
commands:
  - run: flatpak install -y flathub com.moonlight_stream.Moonlight
    description: Installs Moonlight from Flathub. Pair with Sunshine (next) on the host PC for low-latency 4K60+ streaming over LAN.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.moonlight_stream.Moonlight
upstream:
  name: Moonlight
  url: https://moonlight-stream.org
supported_devices: [any]
tags: [streaming, remote-play, gamestream]
---
