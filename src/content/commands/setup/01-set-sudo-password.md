---
title: Set a sudo password
category: setup
order: 10
summary: Required for `sudo`. SteamOS ships without one — set it before anything else.
icon: lucide:key-round
commands:
  - run: passwd
    description: Sets the password for the current user. Required for sudo to work.
idempotent: true
reversible: false
supported_devices: [deck, steam-machine, any]
tags: [setup, sudo]
---
