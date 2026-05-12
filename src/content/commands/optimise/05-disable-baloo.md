---
title: Disable Baloo (file indexer)
category: optimise
order: 50
summary: KDE's file indexer can pin CPU for hours after a fresh install. Disabling reclaims desktop-mode battery and quiets the fans.
icon: lucide:search-x
commands:
  - run: |
      balooctl disable
      kwriteconfig5 --file baloofilerc --group "Basic Settings" --key "Indexing-Enabled" false
    description: Plasma 5. Plasma 6 uses `balooctl6` / `kwriteconfig6`.
idempotent: true
reversible: true
undo: |
  balooctl enable
  kwriteconfig5 --file baloofilerc --group "Basic Settings" --key "Indexing-Enabled" true
upstream:
  name: Baloo (ArchWiki)
  url: https://wiki.archlinux.org/title/Baloo
supported_devices: [deck, steam-machine, any]
tags: [kde, battery, indexer]
---
