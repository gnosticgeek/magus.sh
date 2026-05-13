# magus.sh TUI Specification

**Status:** Draft (validated via interactive prototype at `/tui`, Go implementation in `magus/`)  
**Last Updated:** 2026-05-13  
**Target:** Static Go binary using [Bubble Tea](https://github.com/charmbracelet/bubbletea) + Lipgloss + Bubbles

---

## Overview

The magus.sh TUI is a five-stage terminal application that guides users through selecting and installing SteamOS setup commands. Two flows are available:

**Wizard path (primary):** Install stage-by-stage as you go.
```
splash → pick menu → pick stage → ⚡ install → (next stage) → …
```

**Script path (secondary):** Review all picks, write a script, run it in one shot.
```
splash → pick menu → review → write → run
```

Both paths share the same pick state — picks accumulate across stages and are only cleared on reset.

---

## Visual Design

### Terminal Dimensions
- **Viewport:** 80 columns × 24 rows (standard terminal)
- **Window chrome:** macOS-style (red/yellow/green dots, draggable title bar)
- **Font:** Monospace (Geist Mono, JetBrains Mono, or system fallback)
- **Height cap:** Content area capped at 32 rem with vertical scrolling

### Color Palette

| Name | Purpose | Example |
|------|---------|---------|
| `t-bright` | Focused/primary text | Command titles when focused |
| `t-text` | Normal/secondary text | Selected items, unselected titles |
| `t-muted` | Descriptive/dim text | Summaries, hints, action labels |
| `t-dim` | De-emphasized | Separators, counts, metadata |
| `t-accent` | Interactive/highlight | Cursors, checkboxes, focus state |
| `t-warn` | Warning/alert | Errors, empty state messages |

### Typography

- **Step titles:** accent color, bright weight
- **Stage labels:** padded to 20 chars, bright when focused. In the Pick Menu each stage label is prefixed with its alchemical sigil (see *Sigils*).
- **Command titles:** bright when focused, text if picked, muted if unpicked
- **Checkboxes:** filled (◉) when picked, hollow (◯) otherwise
- **Cursor indicator:** right-pointing chevron (❯) when focused
- **Group rows:** padded name + pick count + dim `›` arrow

### Sigils

Each stage carries a single Unicode glyph that runs through the splash legend, the menu rows, and the right-pane preview header. Defined once in `src/lib/stages.ts` as `STAGE_SIGILS`:

| Stage | Sigil | Element |
|-------|-------|---------|
| setup | 🜃 | Earth |
| install (Apps) | 🜁 | Air |
| optimise | 🜂 | Fire |
| customise | 🜄 | Water |
| gaming | ☉ | Sun |

The same constant is consumed by both `index.astro` and `tui.astro` — the wordmark on the homepage and the menu rows in the TUI never drift.

### Layout — Split Pane (Pick Menu, Pick Stage)

The Pick Menu and Pick Stage views render two columns inside the terminal body:

- **Left (~52% of width):** the row list — stages, presets, action rows, or commands
- **Right (~48% of width):** a *preview pane* that updates as you move the cursor

The split is implemented by wrapping rendered output in `<div class="split">` with a CSS-grid two-column layout, and a thin dashed left border on the right pane. The preview block always opens with a small section header — `── stage ──`, `── preset ──`, `── apps ──`, `── shortcut ──`, `── back ──`, `── quit ──`, `── review ──`, `── group ──`, or `── install ──` — so the user can tell what kind of row they're looking at.

The Splash, Pick Search, Pick Installing, Review, Write, and Run screens render single-column.

### Status Bar — Key Hierarchy

Hint keys at the bottom of the terminal are typed as `'primary' | 'normal' | 'system'`:

- **primary** — the action this view is centred on (`enter open`, `space toggle`, `y retry`, `↵ continue`). Accent-tinted background and border on the `<kbd>`.
- **normal** — supporting navigation (`← →`, `↑ ↓`, `esc`, `/`, `e edit`). Default styling.
- **system** — destructive / escape hatches (`r reset`, `q abort`, `q quit`). Reduced opacity, never primary.

`Hint` carries an optional `kind` field; the renderer maps it to a `data-kind` attribute on each `<kbd>`. Each `kind` should appear at most once per status bar.

### Separators & Structure

- Horizontal rule: `<span class="t-dim">────</span>` between menu sections
- Empty lines: `&nbsp;` to maintain spacing
- Summary text: Prefixed with `<span class="t-dim">›</span>`

---

## Navigation Model

### State Machine

```
splash ──[enter]──> pick(menu) ──[enter stage]──> pick(stage)
                        │                              │
                    [enter preset]              [enter group] ──> pick(group)
                        │                              │              │
                    (applies picks)             [⚡ install]    [esc back]
                                                      │
                                               pick(installing)
                                                      │
                                               [esc when done] ──> pick(menu)
                        │
                    [review] ──> review ──> write ──> run
```

### Keyboard Shortcuts (Global)

| Key | Action | Context |
|-----|--------|---------|
| `←` / `→` | Previous/next step | Any screen |
| `r` / `R` | Reset (return to splash, clear picks) | Any screen |

### Keyboard Shortcuts (Per-View)

#### Splash Screen
| Key | Action |
|-----|--------|
| `enter` | Begin ceremony (go to pick) |

#### Pick Menu (Stage Selection)
| Key | Action |
|-----|--------|
| `↑` / `↓` | Move cursor in menu |
| `enter` | Open focused stage, apply preset, or trigger action |
| `/` | Open fuzzy search across all commands |

#### Pick Stage — Group List (Apps stage)
| Key | Action |
|-----|--------|
| `↑` / `↓` | Move cursor |
| `enter` | Open focused group |
| `esc` | Back to stage menu |

#### Pick Stage — Command List
| Key | Action |
|-----|--------|
| `↑` / `↓` | Move cursor in command list |
| `space` | Toggle checkbox on focused command |
| `enter` | Activate focused row (toggle item, select-all, open install, or back) |
| `/` | Open fuzzy search scoped to current stage |
| `esc` | If inside a group: back to group list. Otherwise: back to menu. |

#### Pick Search
| Key | Action |
|-----|--------|
| Any alphanumeric / `-` / space | Append to search query |
| `backspace` | Remove last character from query |
| `↑` / `↓` | Move cursor in results |
| `space` | Toggle checkbox on focused result |
| `enter` | Toggle checkbox on focused result |
| `esc` | Back to prior view (menu or stage) |

#### Pick Installing — Running
| Key | Action |
|-----|--------|
| *(none — animation runs automatically)* | |

#### Pick Installing — Failed
| Key | Action |
|-----|--------|
| `y` / `Y` | Retry failed command |
| `s` / `S` | Skip failed command |
| `q` / `Q` | Abort install, return to stage |

#### Pick Installing — Done
| Key | Action |
|-----|--------|
| `esc` | Return to stage menu |

#### Review Screen
| Key | Action |
|-----|--------|
| `y` / `Y` | Write script (go to write) |
| `e` / `E` | Edit picks (return to pick) |
| `q` / `Q` | Quit and reset |
| `enter` | Write script |

#### Write Screen
| Key | Action |
|-----|--------|
| `enter` | Continue to run |

#### Run Screen — Prompt
| Key | Action |
|-----|--------|
| `y` / `Y` | Execute script with `bash /tmp/magus.sh` |
| `n` / `N` | Exit (script is saved; user can paste later) |

#### Run Screen — Failed
| Key | Action |
|-----|--------|
| `y` / `Y` | Retry failed command |
| `s` / `S` | Skip failed command |
| `q` / `Q` | Abort install |

---

## Screens

### Splash

**Purpose:** Welcome ritual, ASCII branding, set expectations.

**Content:**
- Block-letter `MAGUS` wordmark on the left, info box on the right (`magus.sh v0.1`, `spells N`, `stages 5`, `runtime ~10 min`)
- Tagline: `transmute your device`
- **Stage legend** (replaces the cryptic glyph row): `── 🜃 setup · 🜁 apps · 🜂 optimise · 🜄 customise · ☉ gaming ──` — pairs each sigil with the stage name once on splash, then the menu rows reuse the sigils as a learned shorthand. Built dynamically from `STAGE_SIGILS` so it stays in sync.
- Two `✓` bullet lines: `idempotent — safe to run again`, `no telemetry · one bash script · paste & walk away`
- Call-to-action: `press [enter] to begin the ceremony`

**No cursor, no movement.**

---

### Pick Menu (Stage Selection)

**Purpose:** Choose which stage(s) to explore, apply a preset, or go to review.

**Layout:** split — row list on the left, preview pane on the right.

**Header:**
```
? Where to next? (↑ ↓ move · enter open)
```

**Rows (in order):**

1. **Stage rows** — one per stage with commands. Each row carries the stage's sigil:
   - `  🜃 01  Setup           0/2` — unfocused, nothing picked (sigil dim)
   - `❯ 🜁 02  Apps            3/11 ✓` — focused, all picked (sigil + count in accent; ✓ in accent)
   - `  🜂 03  Optimise        0/8 ⚡` — installed via wizard (⚡ replaces ✓)
   - Label (`NN  Short`) padded to 20 chars; count in accent if any picked, dim otherwise.

2. **Separator:** `────`

3. **Preset rows** — curated bundles. The menu currently ships three:
   - `✦ Magnum Opus · the full proven kit  12`
   - `✦ Retro Operator · shaders & bezels  8`
   - `✦ Hush Mode · a calm Deck            5`
   - Trailing number is the de-duplicated count of commands the preset would apply.
   - Entering a preset instantly marks all matching commands as picked. Presets *only add*; they never deselect.
   - `✦` icon dim when unfocused, accent when focused.

4. **Separator:** `────`

5. **Action rows:**
   - `Review my picks (N)` — N is total picked count; leads to Review screen
   - `Quit`

**Right pane (preview):** updated on every cursor move.

- **Stage focused** — `── stage ──` header, sigil + `NN Short` title, tagline, `includes` block (groups + per-group counts) for stages that have groups, otherwise a `commands` list (first 8 + `… +N more`), and a footer `N/M picked · ~K min if all`.
- **Preset focused** — `── preset ──` header, `✦ Name`, tagline, `applies N commands` followed by the `+ Title` list (first 12 + `… +N more`), and `enter to apply · won't deselect anything`.
- **Action focused (review / quit)** — short blurb explaining what the row does and a current-state line (e.g., `N commands queued so far`).

**Footer:**
```
{total_picked} of {total_commands} commands selected total
```

**Behavior:**
- Cursor restores to prior position when returning from a stage
- All picks are preserved across navigation
- Installed stages show `⚡` (supersedes `✓`)

---

### Pick Stage — Group List (Apps and Gaming)

**Purpose:** Sub-navigate the larger stages (Apps · 11 commands, Gaming · 23 commands) via logical groups.

**Layout:** split — group list on the left, preview pane on the right.

**Header:**
```
? 02 · Apps (space toggle · enter open)
```

**Rows (Apps):**
- `❯ Capture & Chat       0/2 ›`
- `  System Tools         0/3 ›`
- `  Browsers & Comms     0/4 ›`
- `  Media                0/2 ›`
- `────`
- `⚡ Install N commands` (or `⚡ pick at least one to install` when 0)
- `← back to stages`

**Groups (Apps stage):**

| Group | Items |
|-------|-------|
| Capture & Chat | OBS Studio, Discover Overlay |
| System Tools | Flatseal, Warehouse, Mission Center |
| Browsers & Comms | Bitwarden, Brave, Firefox, Vesktop |
| Media | VLC, Spotify |

**Groups (Gaming stage):**

| Group | Items |
|-------|-------|
| Retro & Emulation | RetroArch, Dolphin Emulator, DuckStation, EmuDeck, RetroDeck, Mega Bezel, Duimon Mega Bezel Shaders |
| Launchers & Compat | ProtonUp-Qt, Wine Cellar, Bottles, Cartridges, Heroic Games Launcher, Lutris, Waydroid (Android container) |
| Streaming & Remote Play | GeForce Now, PS Remote Play (Chiaki4deck), Moonlight, Sunshine |
| Tools & Overlays | GOverlay, Ludusavi |
| Source Ports | GZDoom, OpenMW, DevilutionX |

**Right pane (preview):** for the focused group — `── group ──` header, group name, `N commands · K picked`, then the first 10 items as `◯ Title` / `◉ Title`. For the install / back rows, the same blurbs as in the Command List view (see below).

**Behavior:**
- Enter opens the focused group → command list scoped to that group
- Escape returns to Pick Menu
- Group pick count reflects only items in that group

---

### Pick Stage — Command List

**Purpose:** Select commands within a single stage (or group).

**Layout:** split — command list on the left, preview pane on the right.

**Header (stage root):**
```
? 02 · Apps (space toggle · enter open)
```

**Header (inside a group):**
```
? 02 · Apps › Browsers & Comms (space toggle · enter open)
  Capture & Chat  System Tools  [Browsers & Comms]  Media
```

The second line is the **group tab strip**: every group in the current stage rendered inline, the active group bracketed and accent-coloured. It's a visual breadcrumb only — there is no keyboard shortcut to hop between groups; users escape to the group list and re-enter. (Future: bind `[` / `]` for direct hop.)

**Rows:**
1. Commands (sorted by order):
   - `❯ ◉ Command Name` — focused, picked
   - `  ◯ Another Command` — unfocused, unpicked
   - Checkbox: ◉ if picked, ◯ if unpicked
   - Color: bright if focused, text if picked, muted if unpicked

2. Separator: `────`

3. Special rows:
   - `select all in this stage` — toggles all items
   - `← back to Apps` (if inside a group) — returns to group list
   - `⚡ Install N commands` (if at stage root, with picks) — enters installing view
   - `⚡ pick at least one to install` (if at stage root, no picks) — dim, non-focusable
   - `← back to stages` (if at stage root) — returns to menu

**Footer:**
```
N of M selected in this stage
```

The single-line summary that used to sit above this footer is gone — its content now lives in the right pane.

**Right pane (preview):** updates on every cursor move.

- **Command focused** — `── <stage> ──` header (e.g., `── apps ──`), `◯/◉ Title`, full `summary`, then a `will run` block listing the first 3 `run` strings as `$ <command>` (with `… +N more` if the command bundles more than 3 lines), and a final hint line `space toggles · enter also toggles`.
- **Group focused (group list)** — `── group ──`, name, `N commands · K picked`, then the first 10 items as `◯/◉ Title`.
- **Install focused** — `── install ──`, either `nothing picked yet · space toggles a command…` or `⚡ N commands ready · est. ~M min · <stage> stage only` followed by `✓ Title` rows for the picked items.
- **select-all focused** — `── shortcut ──`, `Select all in this stage` / `Deselect all in this stage` (toggle-aware), and a one-line `toggles every command in <stage> (M total)` blurb.
- **back / back-group focused** — `── back ──`, the row label, and a one-liner reassuring the user that picks carry across.

**Behavior:**
- Space toggles focused command
- Enter on command: same as space
- Enter on select-all: toggles all in current scope (stage or group)
- Enter on install: starts installing if picks > 0, disabled if nothing selected
- Escape: if inside group → group list; if at stage root → menu

---

### Pick Installing

**Purpose:** Animate per-stage installation with live feedback.

**Running state:**
```
installing 2/5  [████████░░░░░░░░░░░░░░] 40%

  ✓ ProtonUp-Qt
  ✓ Heroic Games Launcher
  › Bottles ▌
    3 more queued
```

**Failed state:**
```
installing 2/5  [████████░░░░░░░░░░░░░░] 40%

  ✓ ProtonUp-Qt
  ✓ Heroic Games Launcher
  ✗ Bottles

  error · connection timed out — could not reach download server (simulated)
  exit code 1

  [y] retry   [s] skip   [q] abort
```

The trailing `(simulated)` chip is rendered in dim text. It clarifies that the failure is a deterministic prototype demo (triggered at index 2 when `picked.length ≥ 4`), not a real network error — reviewers were assuming the prototype was broken. Strip it out when the TUI runs against a real installer.

**Done state:**
```
Apps done ✓

  3 installed   1 skipped

  ✓ ProtonUp-Qt
  ✓ Heroic Games Launcher
  — Bottles
  ✓ Cartridges

press [esc] to continue to next stage
```

**Behavior:**
- Each command simulates ~800ms of install time
- Progress bar: `█` filled, `░` unfilled, 22 chars wide
- `installedStageIds` set is updated when install completes (stage shows ⚡ in menu)
- Abort (`q`) returns to stage view without marking stage as installed

---

### Review

**Purpose:** Cross-stage summary of all picks before writing a script.

**Header (if no picks):**
```
nothing picked.

press [←] to go back, or [r] to restart.
```

**Header (with picks):**
```
{N} commands ready · est. ~{mins} min
```

**Body (grouped by stage):**
```
  02 Apps
     ✓ Heroic Games Launcher
     ✓ ProtonUp-Qt

  03 Optimise
     ✓ CryoUtilities
```

**Footer:**
```
―
[y] write to disk     [e] edit picks     [q] quit
```

**Time Estimate:** `max(2, round(picked.length * 1.4))` minutes

---

### Write

**Purpose:** Animate script generation and file I/O.

**Sequence (with delays):**
1. `→ rendering magus.sh ................ done (X.X KB)` — 480ms
2. `→ writing /tmp/magus.sh ............. done` — 480ms
3. `→ chmod +x .......................... done` — 380ms
4. Blank line — 220ms
5. Done:
   ```
   saved to: /tmp/magus.sh

   N commands · X.X KB · idempotent

   › press [enter] to continue ▌
   ```

**Script Size Estimate:** `max(0.4, picked.length * 0.18).toFixed(1)` KB

**File Output:**
- Path: `/tmp/magus.sh`
- Header: `#!/usr/bin/env bash` + `# magus.sh — selected commands`
- Per-command: comment with title/summary, then `run` lines
- Ordering: by stage order, then by command order

---

### Run

**Purpose:** Execute the script or save for later.

**Prompt state:**
```
script ready.

run now?  [y] yes, install     [n] save & paste later

or later in Konsole:
  $ bash /tmp/magus.sh

── preview ────────────────────────────────────
#!/usr/bin/env bash
# magus.sh — N commands

# Command 1
bash_command_here

… +N more
```

**Running / failed / done states:** Same UI pattern as Pick Installing (see above), applied to the full picked set across all stages.

**Behavior:**
- `y`: Simulate executing `/tmp/magus.sh`
- `n`: Exit; script remains at `/tmp/magus.sh`
- Script preview shows first 4 commands

---

## Data Structures

### Command (Cmd)

```typescript
type Cmd = {
  id: string;              // e.g., "setup/01-set-sudo-password"
  title: string;           // Display name
  summary?: string;        // One-line description
  commands: Array<{
    run: string;           // Bash command to execute
    description?: string;  // Inline comment for script
  }>;
};
```

### Stage

```typescript
type Stage = {
  id: string;     // e.g., "setup", "install"
  num: string;    // Display number (e.g., "01")
  short: string;  // Display name (e.g., "Apps")
  tagline: string; // One-line description; surfaced in the Pick Menu preview pane
  sigil: string;  // Alchemical glyph from STAGE_SIGILS (e.g., "🜁")
  items: Cmd[];   // Commands in this stage
};
```

`tagline` and `sigil` are populated server-side in `tui.astro` from `STAGES` and `STAGE_SIGILS` (both exported from `src/lib/stages.ts`).

### AppGroup

```typescript
type AppGroup = {
  id: string;          // e.g., "retro"
  name: string;        // Display name, e.g., "Retro & Emulation"
  patterns: RegExp[];  // Matched against cmd.id to assign commands to group
};
```

Groups are defined in `STAGE_GROUPS: Record<string, AppGroup[]>`, currently for the `install` (Apps, 4 groups) and `gaming` (Gaming, 5 groups) stages. Pattern matching is order-independent; each command must match exactly one group.

### Preset

```typescript
type Preset = {
  id: string;          // e.g., "magnum-opus"
  name: string;        // Display name
  tagline: string;     // Short description shown in menu
  patterns: RegExp[];  // Matched against cmd.id to select commands
};
```

**Current presets:**

| Preset | Tagline | Commands |
|--------|---------|----------|
| Magnum Opus | the full proven kit | Set Sudo Password, Install Dependencies, ProtonUp-Qt, Heroic, CryoUtilities, Wi-Fi Powersave, Tablet Mode, Decky Loader, CSS Loader, SteamGridDB, Flatseal, Brave (12) |
| Retro Operator | shaders & bezels | Set Sudo Password, Install Dependencies, RetroArch, Dolphin, DuckStation, RetroDECK, Duimon Mega Bezel, ProtonUp-Qt (8) |
| Hush Mode | a calm Deck | Set Sudo Password, Wi-Fi Powersave, Tablet Mode, Flatseal, Brave (5) |

The de-duplicated match count for each preset is computed by the `presetMatches(preset): Cmd[]` helper — patterns may overlap without inflating the displayed count.

### Row Types

```typescript
type MenuRow =
  | { kind: 'stage'; stage: Stage }
  | { kind: 'preset'; preset: Preset }
  | { kind: 'action'; id: 'review' | 'quit'; label: string };

type StageRow =
  | { kind: 'item'; cmd: Cmd }
  | { kind: 'group'; group: AppGroup }
  | { kind: 'select-all' }
  | { kind: 'install' }
  | { kind: 'back-group' }
  | { kind: 'back' };
```

### State (Internal)

```typescript
type PickView = 'menu' | 'stage' | 'search' | 'installing';
type InstallPhase = 'prompt' | 'running' | 'failed' | 'done';

type PickState = {
  step: 'splash' | 'pick' | 'review' | 'write' | 'run';
  cursor: number;
  picked: Set<string>;
  pickView: PickView;
  currentStageId: string | null;
  currentGroupId: string | null;       // Open group within a stage (null = group list)
  menuCursor: number;                  // Saved menu position (restored on back)
  searchQuery: string;
  searchResults: Array<{ stage: Stage; cmd: Cmd }>;
  priorView: 'menu' | 'stage';
  installedStageIds: Set<string>;      // Stages completed via wizard path
  installPhase: InstallPhase;
  installIndex: number;                // Current command index during install
  installLog: Array<{ title: string; result: 'done' | 'skipped' }>;
};
```

---

## Behavior Rules

### Selection & Persistence

- Picks are preserved when navigating between stages and groups
- Picks are cleared only on reset (`r` key or Quit action)
- All stages can be visited in any order
- Presets apply immediately on enter; they OR into the existing pick set (no deselect)

### Install Wizard

- "Install N commands" is disabled (dim, non-focusable) if stage picks = 0
- Each command simulates ~800ms; failure is triggered at install index 2 when ≥ 4 items
- `installRunId` counter cancels stale async callbacks when navigating away
- On completion, stage ID is added to `installedStageIds` (shows ⚡ in menu)
- Abort returns to stage view; stage is NOT added to `installedStageIds`

### Group Navigation

- Escape from within a group → group list (not all the way to menu)
- Escape from group list → menu
- Header shows breadcrumb: `02 · Apps › Gaming` when inside a group
- "select all" in a group scopes only to that group's commands

### Scrolling & Focus

- Focused row is highlighted with background color
- Focused row scrolls into view (nearest edge)
- Cursor stops at first/last row — no wraparound

### Search Scoping

- From menu: search all commands across all stages
- From stage or group: search only commands in current stage
- Prior view is saved; Escape returns to the correct location

### Select All Behavior

- Toggle: if all items in scope are picked → deselect all; otherwise → select all
- Applies only to items in current scope (stage root or open group)
- Cursor remains on select-all row after action

### Validation

- Review is accessible regardless of pick count (shows warning if empty)
- Install button is disabled if no picks in that stage
- Write step always succeeds (no validation)

---

## Command Catalogue

**Total: 51 commands across 5 stages.** Stage IDs (`setup`, `install`, `optimise`, `customise`, `gaming`) are the source of truth — display names live in `STAGES`, sigils in `STAGE_SIGILS`, both exported from `src/lib/stages.ts`.

### Setup (01) — 2 commands · sigil 🜃
| Order | Command | Notes |
|-------|---------|-------|
| 10 | Set a sudo password | Prerequisite for anything that touches `/etc`, `/var`, or kernel modules. |
| 20 | Install Dependencies | nvm + Node.js 22 + tsx + @inquirer/prompts, installs to `~/.nvm` (survives SteamOS updates). The TUI itself runs on this stack. |

### Apps (02) — 11 commands across 4 groups · sigil 🜁
Managed via `STAGE_GROUPS['install']`. All Flatpak IDs verified against Flathub. Notable:
- `io.github.trigg.discover_overlay` — lowercase `discover_overlay` (not `Discover_overlay`)

| Order | Command | Group |
|-------|---------|-------|
| 34 | OBS Studio | Capture & Chat |
| 36 | Discover Overlay | Capture & Chat |
| 40 | Flatseal | System Tools |
| 42 | Warehouse | System Tools |
| 44 | Mission Center | System Tools |
| 60 | Brave | Browsers & Comms |
| 62 | Firefox | Browsers & Comms |
| 64 | Bitwarden | Browsers & Comms |
| 66 | Spotify | Media |
| 68 | Vesktop | Browsers & Comms |
| 70 | VLC | Media |

### Optimise (03) — 8 commands · sigil 🜂
Mostly small `kwriteconfig5` / `qdbus` invocations and `/etc` drop-ins.

| Order | Command | Notes |
|-------|---------|-------|
| 10 | Install CryoUtilities | Performance + filesystem tweaks (community installer). |
| 20 | Force touch mode | `kwriteconfig5` — KWin tablet-mode. |
| 30 | Larger cursor for touch | KCM cursor-size override. |
| 40 | Double-click to open | KDE single-click → double-click. |
| 50 | Disable Baloo (file indexer) | Stops background indexing on `~`. |
| 60 | Persistent Wi-Fi power save | Drops a NetworkManager `conf.d` snippet. |
| 70 | Btrfs `/home` conversion | Filesystem migration; high-impact, marked `danger: medium`. |
| 80 | Nested Desktop in Game Mode | Writes `~/.local/bin/PlasmaNested.sh` (davidedmundson script). User finishes by adding it as a non-Steam game once. |

### Customise (04) — 7 commands · sigil 🜄
All Decky plugins — installed via Decky Loader (the Setup stage covers Loader installation through the Magnum Opus / Hush Mode presets and stage 01).

| Order | Command | Notes |
|-------|---------|-------|
| 1 | Install Decky Loader | Plugin runtime — same idempotent script as Setup. Safe to re-run. |
| 2 | CSS Loader | Theme engine for the Steam UI. |
| 3 | SteamGridDB | Grid art for non-Steam games. |
| 4 | ProtonDB Badges | Compatibility badges in the library. |
| 5 | HLTB for Deck | "How long to beat" overlay. |
| 6 | PlayTime | Per-title playtime tracking. |
| 7 | AutoFlatpaks | Background Flatpak updater. |

### Gaming (05) — 23 commands across 5 groups · sigil ☉
Managed via `STAGE_GROUPS['gaming']`. Flatpak-first where Flathub coverage exists; `git clone` for source-port content (Mega Bezel) and AppImage for orphans (DuckStation).

| Order | Command | Group |
|-------|---------|-------|
| 100 | RetroArch | Retro & Emulation (Flatpak `org.libretro.RetroArch`) |
| 110 | Dolphin Emulator | Retro & Emulation (Flatpak `org.DolphinEmu.dolphin-emu`) |
| 120 | DuckStation | Retro & Emulation (AppImage from GitHub — removed from Flathub) |
| 130 | EmuDeck | Retro & Emulation |
| 140 | RetroDeck | Retro & Emulation (Flatpak `net.retrodeck.retrodeck`) |
| 150 | Mega Bezel | Retro & Emulation (shader pack via RetroArch Online Updater) |
| 160 | Duimon Mega Bezel Shaders | Retro & Emulation — `git clone` into `~/.var/app/org.libretro.RetroArch/config/retroarch/shaders/Mega_Bezel_Packs/` |
| 200 | ProtonUp-Qt | Launchers & Compat |
| 210 | Wine Cellar | Launchers & Compat |
| 220 | Bottles | Launchers & Compat |
| 230 | Cartridges | Launchers & Compat |
| 240 | Heroic Games Launcher | Launchers & Compat |
| 250 | Lutris | Launchers & Compat |
| 260 | Waydroid (Android container) | Launchers & Compat — clones ryanrudolfoba/SteamOS-Waydroid-Installer; user finishes from a terminal (interactive). `deck_only: true`. |
| 300 | GeForce Now | Streaming & Remote Play |
| 310 | PS Remote Play (Chiaki4deck) | Streaming & Remote Play |
| 320 | Moonlight | Streaming & Remote Play |
| 330 | Sunshine | Streaming & Remote Play |
| 400 | GOverlay | Tools & Overlays |
| 410 | Ludusavi | Tools & Overlays |
| 500 | GZDoom | Source Ports |
| 510 | OpenMW | Source Ports |
| 520 | DevilutionX | Source Ports |

---

## Edge Cases & Notes

### Empty Stages
- Stages with no commands are not displayed in the menu
- Search results can be empty — show: `no spells match "query"`

### Very Long Command Names
- Truncate to 30 chars in search results
- Pad to fixed width for alignment
- Full name shown in stage/group view

### Network / External State
- Pre-flight checks: detect Steam Deck before offering Deck-only commands (future)
- Script output: `/tmp/magus.sh`

### Performance
- Substring search filtered in-memory (acceptable for ~50–100 commands)
- Show up to 12 results with `… +N more` indicator

---

## Implementation Notes

### Technology Stack
- **Runtime:** Statically-linked Go binary — no runtime, no package manager, no `$PATH` surgery
- **UI library:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm-style update/view/model) + [Lipgloss](https://github.com/charmbracelet/lipgloss) (styles) + [Bubbles](https://github.com/charmbracelet/bubbles) (progress/spinner primitives)
- **Scaffold:** `magus/` directory
- **Distribution:** prebuilt `linux/amd64` binary from GitHub releases; ~5 MB, drops into `~/.local/bin/magus` and survives SteamOS immutable updates
- **Data:** `magus/commands.json` is embedded at compile time via `//go:embed`; regenerated from `src/content/commands/**/*.md` by `magus/scripts/gen-commands.mjs`

### File Structure
```
magus/
├── go.mod               # module: magus
├── commands.json        # generated catalogue, embedded via //go:embed
├── scripts/
│   └── gen-commands.mjs # walks markdown frontmatter → commands.json
├── data.go              # Catalogue + Cmd/Stage/Group/Preset + lookups
├── styles.go            # lipgloss palette + status-bar key hierarchy
├── main.go              # Model, Update, View dispatcher
├── splash.go            # splash screen
├── menu.go              # pick menu (split pane)
├── stage.go             # pick stage (groups + commands + select-all)
├── search.go            # fuzzy search across catalogue
├── install.go           # per-stage install animation
├── review.go            # cross-stage review
├── write.go             # script-write animation
└── run.go               # final-run prompt + execution
```

The Astro `/tui` page remains the visual design source of truth — both
implementations consume the same content collection and `STAGE_SIGILS`.

### Testing Strategy
1. Unit tests for state transitions (pick → install → done, escape chains)
2. Integration tests for keyboard input → state changes
3. Manual testing of all keyboard shortcuts
4. Visual regression tests (ASCII output comparison)

### Future Enhancements
1. **`?` help overlay** — context-sensitive keyboard shortcut reference
2. **Undo last action** — pop last toggle from a history stack
3. **History tracking** — remember user's last selections across runs
4. ~~**Smart preview pane** — show command details as you navigate~~ ✓ shipped 2026-05-13 (split layout in Pick Menu and Pick Stage)
5. **Command filtering** — filter by tag or danger level
6. **Pre-flight checks** — detect real Steam Deck, skip Deck-only commands otherwise (relevant to `deck_only: true` items like Waydroid)
7. **Custom output path** — let user choose script destination
8. **Group hop shortcuts** — `[` / `]` to cycle through the group tab strip without bouncing back to the group list
9. **Live syntax-tinted Run preview** — line numbers in a gutter, fade-out instead of `… +N more`, comments dim / `bash`-keywords accent / `$` calls bright
10. **Live script-assembly Write screen** — stream the generated `magus.sh` into a small code window as it "writes," replacing the dot-padded animation
11. **In-frame status bar** — pull the keyboard-hint rail inside the terminal chrome (vim/btop convention) so the prototype stops reading like dev-tools chrome
12. **Preset-applied confirmation** — brief in-pane band (`✦ Magnum Opus applied · +5 new, 7 already picked`) so the user knows the action landed

---

## Appendix: Terminal Capabilities

### Assumed Features
- ANSI color support (256 colors or 24-bit)
- Unicode support (◉ ◯ ❯ ← █ ░ ▌ etc.)
- Raw mode input (keystroke capture)
- Cursor positioning (focused row highlighting)

### Fallback Behavior
- If ANSI unavailable: ASCII alternatives (`[x]` instead of ◉, `>` instead of ❯)
- If Unicode unavailable: ASCII chars only
- If raw mode unavailable: line-buffered input (less responsive)

---

## Changelog

| Date | Change |
|------|--------|
| 2026-05-12 | Initial specification from interactive prototype |
| 2026-05-12 | Added select-all feature for stages |
| 2026-05-12 | Locked down keyboard shortcuts and state machine |
| 2026-05-12 | Redesigned install flow: per-stage wizard path added alongside script path |
| 2026-05-12 | Removed "Done — write the script" from menu; primary CTA is now per-stage ⚡ install |
| 2026-05-12 | Added preset system (Magnum Opus — 12 curated commands) |
| 2026-05-12 | Apps stage sub-grouped into 5 logical groups (Gaming, Streaming, System, Comms, Media) |
| 2026-05-12 | Install Dependencies replaces Install Decky Loader in Setup stage |
| 2026-05-12 | DuckStation switched to AppImage install (removed from Flathub) |
| 2026-05-12 | Duimon Mega Bezel added to Retro stage |
| 2026-05-12 | Fixed Discover Overlay Flatpak ID casing (`discover_overlay`) |
| 2026-05-13 | Gaming extracted from Apps into its own 5th stage with 5 groups (Retro & Emulation, Launchers & Compat, Streaming & Remote Play, Tools & Overlays, Source Ports); Apps reduced to 11 commands across 4 groups |
| 2026-05-13 | `STAGE_SIGILS` extracted to `src/lib/stages.ts` as the single source of truth; `Stage` type gained `sigil` and `tagline` fields |
| 2026-05-13 | Splash: cryptic glyph row replaced with labeled stage legend (`🜃 setup · 🜁 apps …`); `[+]` bullets switched to `✓` |
| 2026-05-13 | Pick Menu stage rows now prepend the stage's alchemical sigil |
| 2026-05-13 | Split-pane preview added to Pick Menu (right pane shows focused stage's groups / preset's commands / action blurb) and Pick Stage (focused command's full summary + the bash that'll run, focused group's items, etc.) |
| 2026-05-13 | Group tab strip rendered above the command list when inside an Apps or Gaming group |
| 2026-05-13 | Status bar key hierarchy: hints carry `kind: 'primary' \| 'normal' \| 'system'`; primary kbds are accent-tinted, system kbds dimmed |
| 2026-05-13 | Install button copy returned to spec (`⚡ Install N commands`); empty state reads `⚡ pick at least one to install` (replaces the "summon spirits" interim copy) |
| 2026-05-13 | Failure messages annotate with a dim `(simulated)` chip so prototype reviewers don't think the demo is broken |
| 2026-05-13 | Two new presets: Retro Operator (8 commands · shaders & bezels) and Hush Mode (5 commands · a calm Deck) |
| 2026-05-13 | New Optimise command: Nested Desktop in Game Mode (order 80) — writes `~/.local/bin/PlasmaNested.sh` |
| 2026-05-13 | New Gaming command: Waydroid (Android container) (order 260, Launchers & Compat) — clones the SteamOS-Waydroid-Installer; `deck_only: true` |
| 2026-05-13 | In-terminal CSS (`.t-*`, `.cursor-blink`, `.line.is-focused`, `.split*`) moved to a `<style is:global>` block — these elements are created at runtime by the page script and don't carry Astro's per-component scope attribute. Side benefit: `.t-accent` now actually renders accent-orange (was silently broken) |
| 2026-05-13 | Technology stack switched from Node.js + @inquirer/prompts to Go + Bubble Tea + Lipgloss. `@inquirer/prompts` is built for one-prompt-at-a-time inquiry — the spec's full-screen stateful app (split panes, animated progress, state machine) maps onto Bubble Tea's update/view/model 1:1. Distribution is now a single static binary instead of nvm + Node + tsx + npm install. Mockup lives in `magus/`. |
