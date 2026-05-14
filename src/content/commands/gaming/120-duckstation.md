---
title: DuckStation
category: gaming
order: 120
group: retro
summary: PS1 emulator. Per-game compatibility, sharp upscaling. Installed as an AppImage — no longer on Flathub.
icon: lucide:joystick
commands:
  - run: mkdir -p ~/Applications
    description: Creates ~/Applications if it doesn't exist — a safe home for AppImages on SteamOS.
  - run: |
      curl -L "https://github.com/stenzek/duckstation/releases/latest/download/duckstation-qt-x64-appimage-build.AppImage" \
        -o ~/Applications/DuckStation.AppImage
      chmod +x ~/Applications/DuckStation.AppImage
    description: Downloads the latest DuckStation AppImage from GitHub and makes it executable. Run ~/Applications/DuckStation.AppImage to launch.
idempotent: true
reversible: true
undo: rm -f ~/Applications/DuckStation.AppImage
upstream:
  name: DuckStation
  url: https://github.com/stenzek/duckstation
supported_devices: [any]
tags: [emulator, ps1]
---
