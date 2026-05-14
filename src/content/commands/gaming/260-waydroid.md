---
title: Waydroid (Android container)
category: gaming
order: 260
group: launchers
summary: Run Android apps and games inside a Wayland container. Useful for the handful of titles that never came to PC — gacha launchers, mobile-only ports, that one casual game your friends still play.
icon: lucide:smartphone
commands:
  - run: |
      if [ -d ~/SteamOS-Waydroid-Installer ]; then
        git -C ~/SteamOS-Waydroid-Installer pull --ff-only
      else
        git clone --depth 1 https://github.com/ryanrudolfoba/SteamOS-Waydroid-Installer ~/SteamOS-Waydroid-Installer
      fi
    description: Clones (or fast-forwards) ryanrudolfoba's installer to `~/SteamOS-Waydroid-Installer`. magus.sh stops here on purpose — the installer itself is interactive (it asks which Android 13 build and whether to add Google Apps, and needs the sudo password from Setup). Finish in a terminal — `cd ~/SteamOS-Waydroid-Installer && ./steamos-waydroid-installer.sh` — or fire it from the desktop entry the script creates. Re-running this command just keeps the installer up to date. Requires ≥5 GB free in `~`, ≥100 MB in `/var`, and SteamOS 3.7+.
idempotent: true
reversible: true
undo: |
  cd ~/SteamOS-Waydroid-Installer && ./uninstall-waydroid.sh
  rm -rf ~/SteamOS-Waydroid-Installer
upstream:
  name: ryanrudolfoba/SteamOS-Waydroid-Installer
  url: https://github.com/ryanrudolfoba/SteamOS-Waydroid-Installer
danger: medium
deck_only: true
supported_devices: [deck]
tags: [android, waydroid, container, mobile]
---
