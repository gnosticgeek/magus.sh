---
title: Force touch mode
category: optimise
order: 20
summary: Switches Plasma to its touch-friendly layout. Use `auto` to let it decide.
icon: lucide:tablet
commands:
  - run: |
      kwriteconfig5 --file kwinrc --group Input --key TabletMode on
      qdbus org.kde.KWin /KWin reconfigure
    description: Writes to `~/.config/kwinrc` and reloads KWin live. Plasma 5 only — Plasma 6 uses `kwriteconfig6` / `qdbus6`.
idempotent: true
reversible: true
undo: |
  kwriteconfig5 --file kwinrc --group Input --key TabletMode auto
  qdbus org.kde.KWin /KWin reconfigure
supported_devices: [deck, steam-machine, any]
tags: [kde, touch, kwin]
---
