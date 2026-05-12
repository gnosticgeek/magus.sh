# magus.sh TUI Specification

**Status:** Draft (validated via interactive prototype at `/tui`)  
**Last Updated:** 2026-05-12  
**Target:** Node.js CLI using @inquirer/prompts

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
- **Stage labels:** padded to 22 chars, bright when focused
- **Command titles:** bright when focused, text if picked, muted if unpicked
- **Checkboxes:** filled (◉) when picked, hollow (◯) otherwise
- **Cursor indicator:** right-pointing chevron (❯) when focused
- **Group rows:** padded name + pick count + dim `›` arrow

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
- Centered ASCII logo (geometric diamond with "magus" centered)
- Tagline: "transmute your deck"
- Alchemical symbols (5 stages): 🜃🜁🜂🜄☉
- Welcome message (ritual-themed language)
- Stats: number of spells (commands), number of stages, estimated time
- Call-to-action: `press [enter] to begin the ceremony`

**No cursor, no movement.**

---

### Pick Menu (Stage Selection)

**Purpose:** Choose which stage(s) to explore, apply a preset, or go to review.

**Header:**
```
? Where to next? (↑ ↓ move · enter open)
```

**Rows (in order):**

1. **Stage rows** — one per stage with commands:
   - `❯ 01  Setup             0/5` — unfocused, nothing picked
   - `❯ 02  Apps              3/23 ✓` — all picked (✓ in accent)
   - `❯ 02  Apps              3/23 ⚡` — installed via wizard (⚡ replaces ✓)
   - Label padded to 22 chars; count in accent if any picked, dim otherwise

2. **Separator:** `────`

3. **Preset rows** — curated bundles:
   - `✦ Magnum Opus · the full proven kit  12 commands`
   - Entering a preset instantly marks all matching commands as picked
   - `✦` icon dim when unfocused, accent when focused

4. **Separator:** `────`

5. **Action rows:**
   - `Review my picks (N)` — N is total picked count; leads to Review screen
   - `Quit`

**Footer:**
```
{total_picked} of {total_commands} commands selected total
```

**Behavior:**
- Cursor restores to prior position when returning from a stage
- All picks are preserved across navigation
- Installed stages show `⚡` (supersedes `✓`)

---

### Pick Stage — Group List (Apps stage only)

**Purpose:** Sub-navigate the 23-command Apps stage via logical groups.

**Header:**
```
? 02 · Apps (space toggle · enter open)
```

**Rows:**
- `❯ Gaming            2/6 ›`
- `  Streaming & Remote 0/5 ›`
- `  System Tools       0/5 ›`
- `  Browsers & Comms   0/4 ›`
- `  Media              0/3 ›`
- `────`
- `⚡ Install N commands`
- `← back to stages`

**Groups (Apps stage):**

| Group | Items |
|-------|-------|
| Gaming | ProtonUp-Qt, Wine Cellar, Bottles, Cartridges, Heroic, Lutris |
| Streaming & Remote | GeForce NOW, Chiaki4deck, Moonlight, Sunshine, OBS Studio |
| System Tools | Discover Overlay, Flatseal, Warehouse, Mission Center, GOverlay |
| Browsers & Comms | Bitwarden, Brave, Firefox, Vesktop |
| Media | VLC, Spotify, Ludusavi |

**Behavior:**
- Enter opens the focused group → command list scoped to that group
- Escape returns to Pick Menu
- Group pick count reflects only items in that group

---

### Pick Stage — Command List

**Purpose:** Select commands within a single stage (or group).

**Header (stage root):**
```
? 02 · Apps (space toggle · enter open)
```

**Header (inside a group):**
```
? 02 · Apps › Gaming (space toggle · enter open)
```

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
   - `⚡ Install N commands` (if at stage root) — enters installing view
   - `← back to stages` (if at stage root) — returns to menu

**Summary & Stats:**
```
› Command summary text here

N of M selected in this stage
```

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

  error · connection timed out — could not reach download server
  exit code 1

  [y] retry   [s] skip   [q] abort
```

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
  items: Cmd[];   // Commands in this stage
};
```

### AppGroup

```typescript
type AppGroup = {
  id: string;          // e.g., "gaming"
  name: string;        // Display name, e.g., "Gaming"
  patterns: RegExp[];  // Matched against cmd.id to assign commands to group
};
```

Groups are defined in `STAGE_GROUPS: Record<string, AppGroup[]>`, currently only for the `install` stage. Pattern matching is order-independent; each command must match exactly one group.

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

| Preset | Commands |
|--------|----------|
| Magnum Opus | Set Sudo Password, Install Dependencies, ProtonUp-Qt, Heroic, CryoUtilities, Wi-Fi Powersave, Tablet Mode, Decky Loader, CSS Loader, SteamGridDB, Flatseal, Brave |

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

### Setup (01)
| Order | Command | Notes |
|-------|---------|-------|
| 10 | Set Sudo Password | Required first step |
| 20 | Install Dependencies | nvm + Node.js 22 + tsx + @inquirer/prompts, installs to `~/.nvm` (survives SteamOS updates) |
| 30 | Install Decky Loader | Plugin framework |
| … | … | … |

### Apps (02) — 23 commands across 5 groups
Managed via `STAGE_GROUPS['install']`. All Flatpak IDs verified against Flathub. Notable:
- `io.github.trigg.discover_overlay` — lowercase `discover_overlay` (not `Discover_overlay`)

### Retro (05)
| Order | Command | Notes |
|-------|---------|-------|
| 10 | RetroArch | Flatpak `org.libretro.RetroArch` |
| 20 | Dolphin | Flatpak `org.DolphinEmu.dolphin-emu` |
| 30 | DuckStation | AppImage from GitHub releases (removed from Flathub) |
| 40 | EmuDeck | |
| 50 | RetroDECK | Flatpak `net.retrodeck.retrodeck` |
| 60 | Duimon Mega Bezel | `git clone` into `~/.var/app/org.libretro.RetroArch/config/retroarch/shaders/Mega_Bezel_Packs/`; requires Mega Bezel shader pack via RetroArch Online Updater |

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
- **Runtime:** Node.js (via nvm) + tsx
- **UI library:** @inquirer/prompts
- **Scaffold:** `tui/` directory
- **Dependencies installed by:** the "Install Dependencies" command in the Setup stage

### File Structure
```
tui/
├── index.ts           # Entry point + state machine
├── load-commands.ts   # Load from content collection JSON
├── screens/
│   ├── splash.ts
│   ├── pick.ts        # Menu, stage, group, search, installing
│   ├── review.ts
│   ├── write.ts
│   └── run.ts
└── types.ts           # Cmd, Stage, AppGroup, Preset, PickState
```

### Testing Strategy
1. Unit tests for state transitions (pick → install → done, escape chains)
2. Integration tests for keyboard input → state changes
3. Manual testing of all keyboard shortcuts
4. Visual regression tests (ASCII output comparison)

### Future Enhancements
1. **`?` help overlay** — context-sensitive keyboard shortcut reference
2. **Undo last action** — pop last toggle from a history stack
3. **History tracking** — remember user's last selections across runs
4. **Smart preview pane** — show command details as you navigate
5. **Command filtering** — filter by tag or danger level
6. **Pre-flight checks** — detect real Steam Deck, skip Deck-only commands otherwise
7. **Custom output path** — let user choose script destination

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
