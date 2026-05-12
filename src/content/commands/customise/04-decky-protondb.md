---
title: ProtonDB Badges
category: customise
order: 4
group: decky
summary: Shows the ProtonDB rating next to every game in the library.
icon: lucide:badge-check
commands:
  - run: |
      mkdir -p ~/homebrew/plugins
      ASSET=$(curl -fsSL https://api.github.com/repos/sersorrel/decky-protondb-badges/releases/latest | grep -oE '"browser_download_url": *"[^"]+\.tar\.gz"' | head -1 | cut -d'"' -f4)
      curl -fL "$ASSET" -o /tmp/protondb.tar.gz
      sudo systemctl stop plugin_loader.service
      tar -xzf /tmp/protondb.tar.gz -C ~/homebrew/plugins/
      sudo chown -R deck:deck ~/homebrew/plugins/
      sudo systemctl start plugin_loader.service
    description: Adds a Platinum/Gold/Silver/Bronze chip to each store page. Verdict pulled from ProtonDB's API at first boot.
idempotent: true
reversible: true
undo: rm -rf ~/homebrew/plugins/protondb-badges && sudo systemctl restart plugin_loader.service
upstream:
  name: sersorrel/decky-protondb-badges
  url: https://github.com/sersorrel/decky-protondb-badges
supported_devices: [deck, steam-machine, any]
danger: low
tags: [decky-plugin, proton, compatibility]
---
