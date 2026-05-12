---
title: PlayTime
category: customise
order: 6
group: decky
summary: Per-session tracker that covers non-Steam games Steam ignores.
icon: lucide:hourglass
commands:
  - run: |
      mkdir -p ~/homebrew/plugins
      ASSET=$(curl -fsSL https://api.github.com/repos/FrogTheFrog/decky-playtime/releases/latest | grep -oE '"browser_download_url": *"[^"]+\.tar\.gz"' | head -1 | cut -d'"' -f4)
      curl -fL "$ASSET" -o /tmp/playtime.tar.gz
      sudo systemctl stop plugin_loader.service
      tar -xzf /tmp/playtime.tar.gz -C ~/homebrew/plugins/
      sudo chown -R deck:deck ~/homebrew/plugins/
      sudo systemctl start plugin_loader.service
    description: Logs sessions to a local SQLite db. Per-game totals, weekly charts, exports as CSV.
idempotent: true
reversible: true
undo: rm -rf ~/homebrew/plugins/Playtime && sudo systemctl restart plugin_loader.service
upstream:
  name: FrogTheFrog/decky-playtime
  url: https://github.com/FrogTheFrog/decky-playtime
supported_devices: [deck, steam-machine, any]
danger: low
tags: [decky-plugin, tracking]
---
