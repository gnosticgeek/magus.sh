---
title: Install Dependencies
category: setup
order: 20
summary: Installs nvm and Node.js 22 to ~/. Required for the magus TUI — safe on SteamOS's immutable file system.
icon: lucide:package
commands:
  - run: curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.3/install.sh | bash
    description: Installs nvm (Node Version Manager) to ~/.nvm. No root required — survives SteamOS updates because it lives entirely in $HOME.
  - run: |
      export NVM_DIR="$HOME/.nvm"
      [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
      nvm install 22
    description: Installs Node.js 22 LTS via nvm. Binaries land in ~/.nvm/versions/node/, not /usr — so they persist across firmware updates.
  - run: |
      export NVM_DIR="$HOME/.nvm"
      [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
      npm install -g tsx @inquirer/prompts
    description: Installs the packages magus needs to run — tsx (TypeScript runner) and @inquirer/prompts (interactive menus).
idempotent: true
reversible: true
undo: rm -rf ~/.nvm && sed -i '/NVM_DIR/d' ~/.bashrc ~/.bash_profile ~/.zshrc 2>/dev/null || true
supported_devices: [deck, steam-machine, any]
danger: low
tags: [setup, node, nvm, dependencies, tui]
---
