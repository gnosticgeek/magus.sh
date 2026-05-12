---
title: Decky Cloud Save
category: customise
order: 8
group: decky
summary: rclone-backed sync for emulator and non-Steam saves. Drive, OneDrive, S3.
icon: lucide:cloud
commands:
  - run: |
      sudo systemctl stop plugin_loader.service
      mkdir -p ~/homebrew/plugins
      curl -L https://github.com/GedasFX/decky-cloud-save/releases/latest/download/decky-cloud-save.tar.gz -o /tmp/dcs.tar.gz
      tar -xzf /tmp/dcs.tar.gz -C ~/homebrew/plugins/
      sudo chown -R deck:deck ~/homebrew/plugins/
      sudo systemctl start plugin_loader.service
    description: Wraps rclone with a touch-friendly UI. Configure the remote once in desktop mode, then back up on every game close.
idempotent: true
reversible: true
undo: rm -rf ~/homebrew/plugins/decky-cloud-save && sudo systemctl restart plugin_loader.service
upstream:
  name: GedasFX/decky-cloud-save
  url: https://github.com/GedasFX/decky-cloud-save
supported_devices: [deck, steam-machine, any]
danger: low
tags: [decky-plugin, saves, backup]
---
