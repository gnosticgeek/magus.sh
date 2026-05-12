---
title: Mission Center
category: install
order: 44
group: system
summary: Pretty, native system monitor. CPU, GPU, RAM, drives, processes.
icon: lucide:activity
commands:
  - run: flatpak install -y flathub io.missioncenter.MissionCenter
    description: Installs Mission Center from Flathub. Looks like a refined macOS Activity Monitor. Quicker to diagnose stutters than tail -f anything.
idempotent: true
reversible: true
undo: flatpak uninstall -y io.missioncenter.MissionCenter
upstream:
  name: Mission Center
  url: https://missioncenter.io
supported_devices: [any]
tags: [monitor, performance, diagnostics]
---
