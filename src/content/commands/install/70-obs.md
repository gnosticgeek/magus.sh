---
title: OBS Studio
category: install
order: 34
group: streaming
summary: Streaming and recording.
icon: simple-icons:obsstudio
commands:
  - run: flatpak install -y flathub com.obsproject.Studio
    description: Installs OBS Studio from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.obsproject.Studio
upstream:
  name: OBS Studio
  url: https://obsproject.com
supported_devices: [any]
tags: [capture, streaming]
---
