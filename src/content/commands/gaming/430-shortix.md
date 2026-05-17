---
title: Shortix
category: gaming
order: 430
group: tools
summary: Human-readable symlinks for Proton prefixes and shader cache folders.
commands:
  - run: |
      TMP=$(mktemp -d)
      git clone --depth 1 https://github.com/Jannomag/shortix "$TMP/shortix"
      mkdir -p /home/deck/Shortix /home/deck/.config/systemd/user
      cp "$TMP/shortix/shortix.sh" /home/deck/Shortix/
      cp "$TMP/shortix/remove_prefix.sh" /home/deck/Shortix/
      cp "$TMP/shortix/shortix.service" /home/deck/.config/systemd/user/
      systemctl --user daemon-reload
      systemctl --user enable --now shortix.service
      rm -rf "$TMP"
    description: Installs Shortix manually as a user service. Protontricks should be installed first.
idempotent: true
reversible: true
undo: systemctl --user disable --now shortix.service; rm -rf /home/deck/Shortix /home/deck/.config/systemd/user/shortix.service; systemctl --user daemon-reload
upstream:
  name: Jannomag/shortix
  url: https://github.com/Jannomag/shortix
supported_devices: [deck, steam-machine]
deck_only: true
danger: medium
tags: [proton, prefixes, shader-cache]
---
