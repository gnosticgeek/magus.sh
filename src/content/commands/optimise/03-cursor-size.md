---
title: Larger cursor for touch
category: optimise
order: 30
summary: Bumps cursor size to 32 px. Easier to track with thumbs in desktop mode.
icon: lucide:mouse-pointer-2
commands:
  - run: |
      kwriteconfig5 --file kcminputrc --group Mouse --key cursorSize 32
      qdbus org.kde.KWin /KWin reconfigure
    description: Default is 24 px. Logout/login may be needed for full effect.
idempotent: true
reversible: true
undo: |
  kwriteconfig5 --file kcminputrc --group Mouse --key cursorSize 24
  qdbus org.kde.KWin /KWin reconfigure
supported_devices: [deck, steam-machine, any]
tags: [kde, cursor, touch]
---
