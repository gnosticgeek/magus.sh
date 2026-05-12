---
title: Double-click to open
category: optimise
order: 40
summary: Switches off KDE's single-click activation. The behaviour everyone else uses.
icon: lucide:mouse-pointer-click
commands:
  - run: |
      kwriteconfig5 --file kdeglobals --group KDE --key SingleClick false
      qdbus org.kde.KWin /KWin reconfigure
    description: Existing Dolphin windows may need a restart to pick up the new behaviour.
idempotent: true
reversible: true
undo: |
  kwriteconfig5 --file kdeglobals --group KDE --key SingleClick true
  qdbus org.kde.KWin /KWin reconfigure
supported_devices: [any]
tags: [kde, click, dolphin]
---
