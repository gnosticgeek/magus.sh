# magus — Go TUI (Bubble Tea mockup)

A statically-linked terminal app that walks the five magus.sh stages and writes
`/tmp/magus.sh`. Replaces the @inquirer/prompts TUI proposed in
[`TUI_SPECIFICATION.md`](../TUI_SPECIFICATION.md) — see the *Implementation
Notes* section there for the visual + navigation contract.

## Quick start

```bash
# from the repo root
brew install go              # one-time, only needed to build
cd magus
go run .                     # iterate
go build -o magus .          # produce the shipped binary (~5 MB, single file)
./magus
```

Or from the repo root: `npm run magus` (see `package.json`).

## Layout

```
magus/
├── go.mod          # module name: magus
├── commands.json   # embedded catalogue (regenerate via scripts/gen-commands.mjs)
├── scripts/
│   └── gen-commands.mjs  # walks src/content/commands → commands.json
├── data.go         # types + JSON load + catalogue lookups
├── styles.go       # lipgloss palette + status-bar key hierarchy
├── main.go         # Model, Update, View dispatcher
├── splash.go       # splash screen
├── menu.go         # pick menu (split pane)
├── stage.go        # pick stage (groups + commands + select-all)
├── search.go       # fuzzy search across catalogue
├── install.go      # per-stage install animation
├── review.go       # cross-stage review
├── write.go        # script-write animation
└── run.go          # final-run prompt + execution
```

## Refreshing the catalogue

After editing any `src/content/commands/**/*.md` frontmatter, regenerate:

```bash
node magus/scripts/gen-commands.mjs
```

The Go binary embeds `commands.json` at compile time via `//go:embed`, so
remember to rebuild after regenerating.

## Distribution

Target is `linux/amd64` (Steam Deck). To produce a release artefact:

```bash
GOOS=linux GOARCH=amd64 go build -o magus-linux-amd64 .
```

Setup stage's "Install Dependencies" downloads this binary from the GitHub
release rather than installing Node + tsx — single file, no runtime, survives
SteamOS immutable updates because it sits in `~/.local/bin/`.

## Why Bubble Tea over @inquirer/prompts

- The spec describes a full-screen stateful app with split-pane previews,
  animated progress bars, a documented state machine, and key-hierarchy hints.
  Bubble Tea's Elm-style update/view/model maps onto `PickState` 1:1; Lipgloss
  handles the `t-bright`/`t-text`/`t-muted` palette as composable styles.
- `@inquirer/prompts` is built for one-question-at-a-time inquiry — full-screen
  TUIs are out of scope, you'd reimplement most of the rendering by hand.
- Static binary distribution beats the nvm/Node/tsx bootstrap on every axis:
  fewer steps, smaller surface, faster startup.
