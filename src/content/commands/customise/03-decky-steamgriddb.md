---
title: SteamGridDB
category: customise
order: 3
group: decky
summary: Replace ugly default artwork on non-Steam shortcuts and EmuDeck entries.
icon: lucide:image
commands:
  - run: |
      sudo systemctl stop plugin_loader.service
      mkdir -p ~/homebrew/plugins
      curl -L https://github.com/SteamGridDB/decky-steamgriddb/releases/latest/download/SteamGridDB.tar.gz -o /tmp/sgdb.tar.gz
      tar -xzf /tmp/sgdb.tar.gz -C ~/homebrew/plugins/
      sudo chown -R deck:deck ~/homebrew/plugins/
      sudo systemctl start plugin_loader.service
    description: Pulls the latest release tarball, drops it in the plugins directory, and bounces the Decky service. A SteamGridDB API key is requested in-plugin (free).
idempotent: true
reversible: true
undo: rm -rf ~/homebrew/plugins/SteamGridDB && sudo systemctl restart plugin_loader.service
upstream:
  name: SteamGridDB/decky-steamgriddb
  url: https://github.com/SteamGridDB/decky-steamgriddb
supported_devices: [deck, steam-machine, any]
danger: low
tags: [decky-plugin, artwork, library]
---
