---
title: SimpleDeckyTDP
category: customise
order: 9
group: decky
summary: Decky plugin for per-game TDP, power governor, CPU boost, and GPU controls.
commands:
  - run: curl -L https://github.com/aarron-lee/SimpleDeckyTDP/raw/main/install.sh | sh
    description: Runs the upstream SimpleDeckyTDP quick installer. Reboot after install for the plugin to settle.
idempotent: true
reversible: false
upstream:
  name: aarron-lee/SimpleDeckyTDP
  url: https://github.com/aarron-lee/SimpleDeckyTDP
supported_devices: [deck, steam-machine]
deck_only: true
danger: medium
tags: [decky, tdp, performance]
---
