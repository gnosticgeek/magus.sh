---
title: NonSteamLaunchers
category: gaming
order: 270
group: launchers
summary: Installs popular Windows launchers and adds them to Steam automatically.
commands:
  - run: /bin/bash -c 'curl -Ls https://raw.githubusercontent.com/moraroy/NonSteamLaunchers-On-Steam-Deck/main/NonSteamLaunchers.sh | nohup /bin/bash -s --'
    description: Opens the official NonSteamLaunchers installer flow so you can choose launchers.
idempotent: true
reversible: false
upstream:
  name: moraroy/NonSteamLaunchers-On-Steam-Deck
  url: https://github.com/moraroy/NonSteamLaunchers-On-Steam-Deck
supported_devices: [deck, steam-machine, any]
danger: medium
tags: [launcher, epic, gog, ea, battle-net]
---
