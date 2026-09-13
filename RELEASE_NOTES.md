# Magus v0.4.8

This Mac alpha release expands the open-source catalogue and adds portable AI
skills shared by Codex and Claude Code. The manifest schema remains `0.4.0`.

## Highlights

- **Shared Agent Skills.** App setups now offers Frontend Design, Systematic
  Debugging, Test-Driven Development, Brainstorming, Writing Plans, Resolving
  Merge Conflicts, React Best Practices and Find Skills. Each source is linked
  and pinned to an immutable Git revision; confirmed installs fetch the whole
  skill folder for both `~/.agents/skills` and `~/.claude/skills`.
- **Safe skill lifecycle.** Preview stays offline, archive extraction is bounded
  and rejects links and path traversal, existing differing files are never
  overwritten, and restore removes only unchanged files Magus created.
- **More open-source Mac apps.** LocalSend, LibreOffice, Maccy, MonitorControl,
  UTM, CotEditor, Joplin, VSCodium, draw.io, Hammerspoon and Objective-See's
  BlockBlock, OverSight, TaskExplorer and What's Your Sign join the catalogue.
- **Media, OCR and developer tools.** YT-DLP, OCRmyPDF, FFmpeg, Tesseract,
  ImageMagick, Neovim and VimR are now selectable. Package notes explain codec,
  language-data and dependency boundaries.
- **Thaw replaces Ice.** The open-source Thaw cask is now the sole recommended
  menu-bar manager of the two, with its macOS and permission requirements shown
  before selection.

## Upgrade

Choose **Update Magus** from the Mac menu, or run:

```sh
curl -fsSL https://magus.sh/install | sh
```

Existing manifests continue to work without a schema change. An old Ice
selection is migrated automatically to Thaw when the manifest or profile loads.
