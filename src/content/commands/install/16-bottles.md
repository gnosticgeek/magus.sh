---
title: Bottles
category: install
order: 16
group: launchers
summary: Run Windows apps in pre-tuned wine bottles.
icon: lucide:flask-conical
commands:
  - run: flatpak install -y flathub com.usebottles.bottles
    description: Installs Bottles from Flathub. Great for the occasional Windows-only utility, modding tools, or installers that don't quite fit Lutris.
idempotent: true
reversible: true
undo: flatpak uninstall -y com.usebottles.bottles
upstream:
  name: Bottles
  url: https://usebottles.com
supported_devices: [any]
tags: [wine, windows, sandbox]
---
