---
title: DevilutionX
category: gaming
order: 520
group: ports
summary: Diablo 1 + Hellfire, rebuilt. Widescreen, controller support, bug fixes, no DRM scaffolding.
icon: lucide:skull
commands:
  - run: flatpak install -y flathub org.diasurgical.DevilutionX
    description: Installs DevilutionX from Flathub. Requires DIABDAT.MPQ (and optionally hellfire.mpq) from a legitimate Diablo install — drop them in ~/.var/app/org.diasurgical.DevilutionX/.local/share/diasurgical/devilution/.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.diasurgical.DevilutionX
upstream:
  name: DevilutionX
  url: https://github.com/diasurgical/devilutionX
supported_devices: [any]
tags: [source-port, arpg, blizzard]
---
