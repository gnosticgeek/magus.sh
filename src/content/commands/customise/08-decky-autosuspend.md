---
title: Decky AutoSuspend
category: customise
order: 8
group: decky
summary: Battery threshold alerts and automatic suspend for low power situations.
commands:
  - run: |
      mkdir -p ~/homebrew/plugins
      TMP=$(mktemp -d)
      curl -fL https://github.com/jurassicplayer/decky-autosuspend/releases/latest/download/decky-autosuspend.zip -o "$TMP/decky-autosuspend.zip"
      unzip -q "$TMP/decky-autosuspend.zip" -d "$TMP"
      sudo systemctl stop plugin_loader.service
      rm -rf ~/homebrew/plugins/decky-autosuspend
      mv "$TMP/decky-autosuspend" ~/homebrew/plugins/decky-autosuspend
      sudo chown -R deck:deck ~/homebrew/plugins/decky-autosuspend
      sudo systemctl start plugin_loader.service
      rm -rf "$TMP"
    description: Downloads the latest AutoSuspend plugin release and restarts Decky Loader.
idempotent: true
reversible: true
undo: rm -rf ~/homebrew/plugins/decky-autosuspend && sudo systemctl restart plugin_loader.service
upstream:
  name: jurassicplayer/decky-autosuspend
  url: https://github.com/jurassicplayer/decky-autosuspend
supported_devices: [deck, steam-machine]
deck_only: true
danger: low
tags: [decky, power, battery]
---
