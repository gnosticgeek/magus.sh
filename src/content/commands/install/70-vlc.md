---
title: VLC
category: install
order: 70
group: daily
summary: Plays anything you throw at it. The forever-default.
icon: simple-icons:vlcmediaplayer
commands:
  - run: flatpak install -y flathub org.videolan.VLC
    description: Installs VLC from Flathub. For when you've downloaded a weird codec and can't be bothered to find a clean rip.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.videolan.VLC
upstream:
  name: VLC
  url: https://www.videolan.org/vlc/
supported_devices: [any]
tags: [media, video, audio]
---
