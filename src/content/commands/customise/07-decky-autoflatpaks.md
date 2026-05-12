---
title: AutoFlatpaks
category: customise
order: 7
group: decky
summary: Background Flatpak updates without dropping to desktop mode.
icon: lucide:refresh-cw
commands:
  - run: |
      mkdir -p ~/homebrew/plugins
      ASSET=$(curl -fsSL https://api.github.com/repos/jurassicplayer/decky-autoflatpaks/releases/latest | grep -oE '"browser_download_url": *"[^"]+\.tar\.gz"' | head -1 | cut -d'"' -f4)
      curl -fL "$ASSET" -o /tmp/autoflatpaks.tar.gz
      sudo systemctl stop plugin_loader.service
      tar -xzf /tmp/autoflatpaks.tar.gz -C ~/homebrew/plugins/
      sudo chown -R deck:deck ~/homebrew/plugins/
      sudo systemctl start plugin_loader.service
    description: Polls Flathub on a schedule, shows a notification badge, applies updates in the background. Pairs well with everything in /02 install.
idempotent: true
reversible: true
undo: rm -rf ~/homebrew/plugins/AutoFlatpaks && sudo systemctl restart plugin_loader.service
upstream:
  name: jurassicplayer/decky-autoflatpaks
  url: https://github.com/jurassicplayer/decky-autoflatpaks
supported_devices: [deck, steam-machine, any]
danger: low
tags: [decky-plugin, flatpak, updates]
---
