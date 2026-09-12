# Magus v0.4.4

This point release makes the Mac catalogue easier to scan, expands its font
selection, and gives clearer guidance around broader Mac utilities. The
review-first flow, preview safety, and reversible Finder preferences remain
unchanged.

## Highlights

- **Terminal tools, in plain language.** The former Command-line tools section
  is now called **Terminal tools**, and Modern CLI is now **Modern commands**
  throughout the menu, website, and documentation.
- **Alphabetical catalogue menus.** App categories, apps, fonts, terminal tools,
  Mac settings, and presets are alphabetized while keeping selection identity
  stable.
- **Two additional fonts.** Atkinson Hyperlegible Next adds an accessibility-led
  everyday sans serif, while Cascadia Code adds another comfortable option for
  terminals and editors. The Mac catalogue now contains eight fonts and 82
  Homebrew items in total.
- **Clearer Vorssaint guidance.** Vorssaint remains available under **Apps > Menu
  Bar**, with accurate compatibility and privacy notes. The Mac settings screen
  recommends it for broader live utilities while Magus keeps ownership of its
  six small, reversible Finder preferences.

## Notes

Mac remains an alpha. Vorssaint currently requires Apple Silicon and macOS 14 or
later; its optional features may request additional macOS permissions. Installing
it through Magus does not enable those features or grant permissions.

## Upgrade

Choose **Update Magus** from the Mac menu, or run the usual installer:

```sh
curl -fsSL https://magus.sh/install | sh
```

Then open a new terminal and run `magus run`. Existing manifests remain on
schema `0.4.0`; v0.4.4 does not require a manifest migration.
