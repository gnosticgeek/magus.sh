---
title: Brave
category: install
order: 60
group: daily
summary: Chromium browser with ad/tracker blocking.
icon: simple-icons:brave
commands:
  - run: flatpak install -y flathub com.brave.Browser
    description: Installs Brave from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.brave.Browser
upstream:
  name: Brave
  url: https://brave.com
supported_devices: [any]
tags: [browser]
---
