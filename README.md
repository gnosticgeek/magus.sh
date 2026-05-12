# magus.sh

> A fresh SteamOS install, opinionated in ten minutes.

A curated, copy-paste-driven setup page for Steam Deck (and SteamOS in general). Open the page on the Deck, tick what you want, copy, paste into Konsole. Wraps the tools the community already trusts — CryoUtilities, Decky Loader, ProtonUp-Qt, Heroic, Lutris — with a consistent UX and no installer to trust blindly.

## Status

`v0.1` alpha. Ships:

- **Performance** — sudo password, CryoUtilities, manual VM tuning, Decky Loader.
- **App pack** — 10 essential Flatpaks (Heroic, Lutris, ProtonUp-Qt, Bitwarden, OBS, Spotify, Ludusavi, …).

Coming: emulation suite, theming, save sync / SteamOS-update resilience, productivity pack, Steam Machine support.

## Why exists

The right tools exist. The right *front door* doesn't. Tinkerer-class Deck owners currently bounce between five GitHub repos and a dozen Reddit threads. magus.sh collapses that into one page, with idempotent commands, transparent provenance, and zero installer to trust.

## Principles

- **Wrap, don't reinvent.** Every command points at an upstream tool that's already proven.
- **Idempotent always.** Running twice is a no-op. Every entry passes this bar before merge.
- **Static, no telemetry.** Pure HTML. No tracking, no analytics, no backend.
- **SteamOS-first, not Deck-only.** Commands tagged with form factor so Steam Machine support is a tag, not a fork.

## Develop

```bash
npm install
npm run dev
```

Opens at `http://localhost:4321`. The Deck's screen is **1280×800** — design at that viewport.

## Adding a command

Drop a Markdown file under `src/content/commands/<category>/`. Schema lives in `src/content.config.ts`. The bar to merge:

1. Idempotent.
2. Reversible (or honest about what it changes).
3. Links upstream.
4. Tagged with `supported_devices`.

## Stack

- Astro 6 + TypeScript (strict)
- Tailwind v4 (via Vite plugin)
- Static, hostable on Cloudflare Pages, GitHub Pages, anywhere

## License

MIT.
