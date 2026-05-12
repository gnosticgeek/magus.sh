---
title: Firefox
category: install
order: 62
group: daily
summary: Mozilla browser.
icon: simple-icons:firefox
commands:
  - run: flatpak install -y flathub org.mozilla.firefox
    description: Installs Firefox from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.mozilla.firefox
upstream:
  name: Firefox
  url: https://www.mozilla.org/firefox/
supported_devices: [any]
tags: [browser]
---
