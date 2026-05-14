---
title: OpenMW
category: gaming
order: 510
group: ports
summary: Open-source Morrowind engine. Modern resolution, longer view distance, better mod support than the original.
icon: lucide:book-open
commands:
  - run: flatpak install -y flathub org.openmw.OpenMW
    description: Installs OpenMW from Flathub. Requires a legitimate Morrowind installation — point OpenMW-CS at your Data Files folder on first run.
idempotent: true
reversible: true
undo: flatpak uninstall -y org.openmw.OpenMW
upstream:
  name: OpenMW
  url: https://openmw.org
supported_devices: [any]
tags: [source-port, rpg, bethesda]
---
