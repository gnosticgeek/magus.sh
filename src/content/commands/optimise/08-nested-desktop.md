---
title: Nested Desktop in Game Mode
category: optimise
order: 80
summary: A pocketable Plasma session inside Game Mode. Skip the full "Switch to Desktop" dance — pop a 1280×800 KDE window over your library when you just need to tweak something.
icon: lucide:picture-in-picture-2
commands:
  - run: |
      install -d ~/.local/bin
      cat > ~/.local/bin/PlasmaNested.sh <<'EOF'
      #!/bin/sh
      # PlasmaNested.sh — runs a nested Plasma Wayland session.
      # Add this script to Steam as a non-Steam game; launching it from Game
      # Mode opens KDE in a 1280×800 window. Closing the window returns you
      # to the library.

      # Performance overlay's LD_PRELOAD interferes with kwin_wayland.
      unset LD_PRELOAD

      # Shadow kwin_wayland_wrapper so we can pass --width/--height through
      # plasma-session, which otherwise calls the wrapper with no args.
      mkdir -p "$XDG_RUNTIME_DIR/nested_plasma"
      cat > "$XDG_RUNTIME_DIR/nested_plasma/kwin_wayland_wrapper" <<'WRAP'
      #!/bin/sh
      /usr/bin/kwin_wayland_wrapper --width 1280 --height 800 --no-lockscreen "$@"
      WRAP
      chmod a+x "$XDG_RUNTIME_DIR/nested_plasma/kwin_wayland_wrapper"
      export PATH="$XDG_RUNTIME_DIR/nested_plasma:$PATH"

      dbus-run-session startplasma-wayland

      rm -f "$XDG_RUNTIME_DIR/nested_plasma/kwin_wayland_wrapper"
      EOF
      chmod +x ~/.local/bin/PlasmaNested.sh
    description: Writes the launcher to `~/.local/bin/PlasmaNested.sh` and makes it executable. To finish setup — once, in Desktop Mode — open Steam, **Add a Non-Steam Game**, browse to `~/.local/bin/PlasmaNested.sh`, rename it "Nested Desktop". Re-running this command overwrites the same file with the same contents.
idempotent: true
reversible: true
undo: rm -f ~/.local/bin/PlasmaNested.sh
upstream:
  name: davidedmundson/PlasmaNested gist
  url: https://gist.github.com/davidedmundson/8e1732b2c8b539fd3e6ab41a65bcab74
danger: low
supported_devices: [deck, steam-machine, any]
tags: [kde, plasma, gamescope, kwin, wayland]
---
