---
title: Duimon Mega Bezel Shaders
category: retro
order: 60
summary: CRT shader presets with reflections and bezels for RetroArch. Requires RetroArch and the Mega Bezel shader pack installed first.
icon: lucide:tv-2
commands:
  - run: |
      mkdir -p ~/.var/app/org.libretro.RetroArch/config/retroarch/shaders/Mega_Bezel_Packs
      cd ~/.var/app/org.libretro.RetroArch/config/retroarch/shaders/Mega_Bezel_Packs
      git clone https://github.com/Duimon/Duimon-Mega-Bezel
    description: Creates the required Mega_Bezel_Packs folder inside RetroArch's shader directory and clones the Duimon preset pack into it. To update later, run git pull inside the Duimon-Mega-Bezel folder.
idempotent: false
reversible: true
undo: rm -rf ~/.var/app/org.libretro.RetroArch/config/retroarch/shaders/Mega_Bezel_Packs/Duimon-Mega-Bezel
upstream:
  name: Duimon Mega Bezel
  url: https://github.com/Duimon/Duimon-Mega-Bezel
supported_devices: [deck, any]
danger: low
tags: [retroarch, shaders, crt, bezel, retro]
---

> **Prerequisites:** RetroArch must be installed, and the Mega Bezel shader pack must be downloaded first via RetroArch → Online Updater → Update Slang Shaders. Then run this command to add the Duimon preset pack. In RetroArch, enable **Simple Presets** (Settings → Video → Shaders) before saving any preset.
