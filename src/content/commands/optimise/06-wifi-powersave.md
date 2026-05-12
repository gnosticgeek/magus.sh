---
title: Persistent Wi-Fi power save
category: optimise
order: 60
summary: Lower background drain by letting the radio idle between frames. Sticks in desktop mode; Game Mode is known to re-enable management after sleep/wake.
icon: lucide:wifi
commands:
  - run: |
      sudo install -m 0644 /dev/stdin /etc/NetworkManager/conf.d/wifi-powersave.conf <<'EOF'
      [connection]
      wifi.powersave = 3
      EOF
      sudo systemctl restart NetworkManager
    description: Writes a NetworkManager drop-in. Re-running overwrites the same file with the same contents — safe and idempotent.
idempotent: true
reversible: true
undo: |
  sudo rm -f /etc/NetworkManager/conf.d/wifi-powersave.conf
  sudo systemctl restart NetworkManager
upstream:
  name: SteamOS Wi-Fi power management (issue tracker)
  url: https://github.com/ValveSoftware/SteamOS/issues/1696
danger: low
supported_devices: [deck, steam-machine, any]
tags: [wifi, battery, networkmanager]
---
