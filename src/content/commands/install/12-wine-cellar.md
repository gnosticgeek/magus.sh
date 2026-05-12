---
title: Wine Cellar
category: install
order: 12
group: launchers
summary: Manage Proton-GE and Wine-GE from Game Mode. Decky alternative to ProtonUp-Qt.
icon: lucide:wine
commands:
  - run: |
      sudo systemctl stop plugin_loader.service
      mkdir -p ~/homebrew/plugins
      curl -L https://github.com/FlashyReese/decky-wine-cellar/releases/latest/download/decky-wine-cellar.tar.gz -o /tmp/wine-cellar.tar.gz
      tar -xzf /tmp/wine-cellar.tar.gz -C ~/homebrew/plugins/
      sudo chown -R deck:deck ~/homebrew/plugins/
      sudo systemctl start plugin_loader.service
    description: Requires Decky Loader (Customise stage) first. Pulls the latest release tarball, drops it in ~/homebrew/plugins/, and bounces the Decky service. Browse and install Proton-GE / Wine-GE builds directly from the Quick Access menu.
idempotent: true
reversible: true
undo: rm -rf ~/homebrew/plugins/decky-wine-cellar && sudo systemctl restart plugin_loader.service
upstream:
  name: FlashyReese/decky-wine-cellar
  url: https://github.com/FlashyReese/decky-wine-cellar
supported_devices: [deck, steam-machine, any]
danger: low
tags: [decky-plugin, proton, wine, compatibility]
---
