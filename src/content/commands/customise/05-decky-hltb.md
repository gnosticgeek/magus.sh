---
title: HLTB for Deck
category: customise
order: 5
group: decky
summary: How Long to Beat estimates on every game page. Useful for handheld sessions.
icon: lucide:clock
commands:
  - run: |
      sudo systemctl stop plugin_loader.service
      mkdir -p ~/homebrew/plugins
      curl -L https://github.com/OMGDuke/HLTB-for-Deck/releases/latest/download/HLTB-for-Deck.tar.gz -o /tmp/hltb.tar.gz
      tar -xzf /tmp/hltb.tar.gz -C ~/homebrew/plugins/
      sudo chown -R deck:deck ~/homebrew/plugins/
      sudo systemctl start plugin_loader.service
    description: Scrapes HowLongToBeat for the active title and surfaces a Main / Main+Extras / Completionist row in the store page.
idempotent: true
reversible: true
undo: rm -rf ~/homebrew/plugins/HLTB-for-Deck && sudo systemctl restart plugin_loader.service
upstream:
  name: OMGDuke/HLTB-for-Deck
  url: https://github.com/OMGDuke/HLTB-for-Deck
supported_devices: [deck, steam-machine, any]
danger: low
tags: [decky-plugin, time, library]
---
