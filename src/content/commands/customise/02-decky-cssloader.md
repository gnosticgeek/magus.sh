---
title: CSS Loader
category: customise
order: 2
group: decky
summary: Custom themes for SteamOS — fonts, colours, layouts. Has its own installer.
icon: lucide:paintbrush
commands:
  - run: curl -L https://github.com/suchmememanyskill/SDH-CssLoader/raw/main/install_release.sh | sh
    description: CSS Loader ships a maintained installer that handles the Decky plugins directory and service restart for you.
idempotent: true
reversible: true
undo: rm -rf ~/homebrew/plugins/SDH-CssLoader && sudo systemctl restart plugin_loader.service
upstream:
  name: suchmememanyskill/SDH-CssLoader
  url: https://github.com/suchmememanyskill/SDH-CssLoader
supported_devices: [deck, steam-machine, any]
danger: low
tags: [decky-plugin, themes, css]
---
