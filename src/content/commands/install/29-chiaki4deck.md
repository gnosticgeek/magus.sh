---
title: PS Remote Play
category: install
order: 29
group: streaming
summary: Stream PS4 / PS5 over LAN or internet. Chiaki4deck — controller-mapped for Steam Deck.
icon: lucide:cast
commands:
  - run: flatpak install -y flathub io.github.streetpea.Chiaki4deck
    description: Installs Chiaki4deck from Flathub — a fork tuned for Deck controls and Game Mode. Register your console once (PSN account + registration PIN), then stream from anywhere.
idempotent: true
reversible: true
undo: flatpak uninstall -y io.github.streetpea.Chiaki4deck
upstream:
  name: streetpea/chiaki4deck
  url: https://github.com/streetpea/chiaki4deck
supported_devices: [any]
tags: [streaming, remote-play, playstation]
---
