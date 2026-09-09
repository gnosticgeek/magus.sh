# Terminal setup proposal

Status: Ghostty setup is implemented as menu item 8. Choose Catppuccin,
TokyoNight or Rose Pine, then use Review & install. The profile is saved in the
Mac manifest and runs after package installation. Export and restore are available
inside Terminal setup. Preview mode makes no changes.

The remaining text records the broader design direction. iTerm2 configuration,
font/size controls, a custom shell prompt and an interactive diff remain future work.
The current review shows the complete proposed Ghostty file in its details pane.
Conflicting legacy/macOS configs and symlinked or manually edited files are rejected
with an explanation; the user can export instead.

## User flow

Add a “Terminal setup” entry. Choose a terminal, choose an appearance, review the
app, font and configuration changes together, then install. Ghostty is the default
choice for this project. Offer iTerm2 as another choice; its configuration adapter
needs separate implementation and verification before promising equivalent setup.
Keep plain app installation available for people who already manage their dotfiles.

Start with Ghostty and three appearance choices: Catppuccin, TokyoNight and Rose
Pine. Use each theme's light/dark variants. Start at 14pt with generous padding and
92% opacity with a soft blur for readability. Offer font size and transparency as optional
preferences rather than a long configuration questionnaire.

The example in dotfiles/ghostty/config.ghostty uses JetBrains Mono, Catppuccin
Mocha/Latte, balanced padding, a steady block cursor and native tabbed titlebars.
Ghostty already embeds JetBrains Mono and Nerd Font symbols. Installing the
existing font-jetbrains-mono cask additionally makes the font available to other
apps; the review should make that distinction clear.

## Persistent, portable configuration

Save the terminal choice, theme and font as a separate terminal profile in the
Magus manifest. Packages alone cannot represent this: installed packages are
pruned from the basket, but an installed terminal may still need configuration.
Add a dedicated reconciliation step for the profile, after package installation.
List that step independently in review, preview, progress and retry handling.

Write ~/.config/ghostty/config.ghostty, respecting an explicit XDG_CONFIG_HOME.
Keep the generated file portable: no machine-specific absolute paths. Provide
“Export dotfiles” to save the same content in a directory the user can put in Git.
A single Ghostty file is sufficient for terminal appearance. A customised shell
prompt would also require a separate prompt configuration and shell integration.

Inspect both XDG and macOS Application Support locations, including legacy config
filenames. Ghostty loads macOS-specific files after XDG files, so those can override
the exported configuration. Detect and explain conflicts before applying.

Back up existing content and permissions before the first change. Show the exact
diff in review. Treat symlinks as externally managed dotfiles: offer an export
rather than replacing the link. Write atomically and track the applied content.
On subsequent runs, leave manually edited files alone until the user chooses how
to reconcile them. Restore only when the current content still matches what Magus
wrote. A repeated run with the same profile should make no change.

Validate the generated file with Ghostty's +validate-config before applying it.
Keep preview entirely free of installation/configuration writes. Test fresh setup,
existing configuration, symlinks, external edits, failed package installation,
reruns, export and restore in isolated home directories.

## Shell customisation

A terminal theme styles the window and terminal colours; the shell owns the prompt.
Make a coordinated prompt an explicit additional feature. Keep the user's current
shell. A later Starship integration can provide directory, Git and command status,
with a separately exported starship.toml and an idempotent, backed-up shell hook.
Do not present this as implemented by the Ghostty example.

## Sources and validation

- https://ghostty.org/docs/config — file locations, precedence, includes and reload.
- https://ghostty.org/docs/config/reference — fonts, padding, cursor and window options.
- https://ghostty.org/docs/features/theme — bundled themes and adaptive appearance.

The local Ghostty installation contains Catppuccin, TokyoNight and Rose Pine
variants. Its configuration validator accepted the included example on 2026-09-09.
The example has not been applied to the user's live configuration or visually
checked in a new Ghostty window.

## Modern commands

In Magus, open **Terminal setup → Modern commands** to select Zsh integrations,
then review and apply. The master switch is off by default. Each integration can
be toggled separately, and **Undo modern terminal commands** queues removal of
the managed shell block. Open a new terminal for changes to take effect.
See [the Mac guide](MAC.md#modern-terminal-commands) for supported commands,
backup behaviour, custom ZDOTDIR and headless configuration.
