---
title: Shader Cache Killer
category: gaming
order: 440
group: tools
summary: Steam Deck script for clearing or moving shader cache and compatdata.
commands:
  - run: curl -sSL https://raw.githubusercontent.com/scawp/Steam-Deck.Shader-Cache-Killer/main/curl_install.sh | bash
    description: Runs the upstream installer, which can add cache cleanup tools to Steam.
idempotent: true
reversible: true
undo: rm -rf /home/deck/.local/share/scawp/SDSCK
upstream:
  name: scawp/Steam-Deck.Shader-Cache-Killer
  url: https://github.com/scawp/Steam-Deck.Shader-Cache-Killer
supported_devices: [deck, steam-machine]
deck_only: true
danger: high
tags: [storage, shader-cache, compatdata]
---
