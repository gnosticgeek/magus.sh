---
title: Mega Bezel
category: gaming
order: 150
group: retro
summary: HSM's reflection shader for RetroArch — CRT scanlines, glass reflections, and bezel art. Foundation for the Duimon preset pack below.
icon: lucide:frame
commands:
  - run: |
      mkdir -p ~/.var/app/org.libretro.RetroArch/config/retroarch/shaders/shaders_slang/bezel
      ASSET=$(curl -fsSL https://api.github.com/repos/HyperspaceMadness/Mega_Bezel/releases/latest | grep -oE '"browser_download_url": *"[^"]+\.zip"' | head -1 | cut -d'"' -f4)
      TMP=$(mktemp -d)
      curl -fL "$ASSET" -o "$TMP/Mega_Bezel.zip"
      unzip -q "$TMP/Mega_Bezel.zip" -d "$TMP"
      SRC=$(find "$TMP" -maxdepth 3 -type d -name 'Mega_Bezel' | head -1)
      rm -rf ~/.var/app/org.libretro.RetroArch/config/retroarch/shaders/shaders_slang/bezel/Mega_Bezel
      mv "$SRC" ~/.var/app/org.libretro.RetroArch/config/retroarch/shaders/shaders_slang/bezel/Mega_Bezel
      rm -rf "$TMP"
    description: Grabs the latest Mega Bezel release zip from GitHub and extracts it into RetroArch's slang shader folder. Equivalent to Online Updater → Update Slang Shaders, but pinned to the official release rather than the live nightly. Set the video driver to Vulkan (or GLCore) in RetroArch before loading a preset.
idempotent: true
reversible: true
undo: rm -rf ~/.var/app/org.libretro.RetroArch/config/retroarch/shaders/shaders_slang/bezel/Mega_Bezel
upstream:
  name: HyperspaceMadness/Mega_Bezel
  url: https://github.com/HyperspaceMadness/Mega_Bezel
supported_devices: [deck, any]
danger: low
tags: [retroarch, shaders, crt, bezel, retro]
---

> **Prerequisites:** RetroArch must be installed (Flatpak edition — the path above is `~/.var/app/org.libretro.RetroArch/…`). After install, in RetroArch set **Settings → Drivers → Video → vulkan** (recommended) or **glcore**, enable advanced settings, set aspect ratio to **Full**, and disable Integer Scale. Then load a preset from `shaders_slang/bezel/Mega_Bezel/Presets/`.
