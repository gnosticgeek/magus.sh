# magus.sh

An open-source terminal setup tool for SteamOS devices and a Mac-first local alpha, with a website that guides users into setup.

## Product status

- **Steam devices:** Steam Deck and Steam Machine paths are available in alpha. Steam Machine hardware detection is unverified on real hardware; HDMI colour range, HDMI-CEC, and performance power profile preferences are not applied yet.
- **Mac:** alpha available: 82 Homebrew apps/tools/fonts with 13 app categories, installed markers, one review basket, Homebrew updates, a built-in checker and verified updater for Magus releases, six reversible Finder preferences, and portable setups for Ghostty, Zed, Firefox, Zsh, and Raycast. See [build and test instructions](magus/MAC.md). The installer supports Apple silicon and Intel release assets.
- **Linux (Flatpaks):** coming soon. General Linux remains unreleased.
- The separate desktop app and website Labs are experimental, not the primary product journey.

## Website

Astro + TypeScript + Tailwind. Pages share a single dark palette, Geist typography, and reusable product components. Platform names, status, and destinations live in `src/lib/platforms.ts`.

| Route | Purpose |
| --- | --- |
| `/` | Product introduction and illustrative interactive workflow |
| `/start` | Platform chooser |
| `/steam` | Installation, launch, device limitations, troubleshooting |
| `/setup` | Advanced Steam Deck command picker |
| `/linux`, `/mac` | Coming-soon pages |
| `/tui`, `/test` | Experimental Labs linked from the footer |
| `/install`, `/run` | Plain-text installer; both use the same source script |

On Mac, the install command verifies Magus and opens the menu when a terminal is available; changes require review inside the menu. On Linux, run `magus run` separately to begin setup. See [the terminal tool documentation](magus/README.md) for the manifest and reconciler contracts.

## Develop and verify

Requires Node 22.12 or later.

```sh
npm install
npm run dev
npm test
npm run check
npm run build
```

The development site opens at `http://localhost:4321`. Review at mobile widths, Steam Deck's **1280×800** display, and a wider desktop viewport. The existing Cloudflare build/deployment setup is retained. Deployment is separate from local review.

Before publishing, verify GitHub Releases contains `magus-linux-amd64`, `magus-linux-arm64`, and `checksums.txt` for the advertised latest release. Browser testing does not verify Steam Machine hardware support.

## Command catalogue

Add Markdown under `src/content/commands/<category>/`; validation lives in `src/content.config.ts`. Commands should be repeatable, link to their upstream source, declare device support, and explain changes or reversal limitations. The web picker retains presets, search, selection, inspection, and generated script review.

## Principles

- Build on established upstream tools.
- Keep commands and configuration inspectable.
- Keep availability and alpha limitations explicit.
- No telemetry, account system, or waitlist backend.

MIT licensed.
