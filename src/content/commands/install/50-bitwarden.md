---
title: Bitwarden
category: install
order: 64
group: daily
summary: Open-source password vault.
icon: simple-icons:bitwarden
commands:
  - run: flatpak install -y flathub com.bitwarden.desktop
    description: Installs the Bitwarden desktop client from Flathub.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.bitwarden.desktop
upstream:
  name: Bitwarden
  url: https://bitwarden.com
supported_devices: [any]
tags: [vault, security]
---
