---
title: Btrfs /home conversion
category: optimise
order: 70
summary: Convert /home to Btrfs for transparent compression. Big storage wins on 64 GB Decks. Survives system updates.
icon: lucide:hard-drive
commands:
  - run: |
      t="$(mktemp -d)" && curl -sSL https://gitlab.com/popsulfr/steamos-btrfs/-/archive/main/steamos-btrfs-main.tar.gz | tar -xzf - -C "$t" --strip-components=1 && "$t/install.sh" && rm -rf "$t"
    description: Downloads the installer payload, runs it, then cleans up. The payload persists across SteamOS updates. WARNING — once /home is converted to Btrfs you cannot return to ext4 without reformatting. Make sure at least 10–20 GiB is free before starting (df -h /home).
idempotent: false
reversible: false
upstream:
  name: popsUlfr/steamos-btrfs
  url: https://github.com/popsUlfr/steamos-btrfs
supported_devices: [deck]
deck_only: true
danger: high
tags: [storage, btrfs, compression, filesystem]
---
