# magus — Steam Machine path: build brief

**Self-contained handoff.** Everything needed to start building, with no prior
conversation context.

Status: draft, August 2026. Nothing here has been run on real Steam Machine
hardware yet — §10 lists what must be verified before anything in §4 is promised.

---

## 1. Decisions already made — don't re-litigate

These were settled deliberately. Reopen them only with new information.

| Decision | Rationale |
|---|---|
| **Magus grows a Steam Machine path** — not a new project | One install command, shared installer core and manifest format. Device detection branches Deck vs Machine. |
| **It's a reconciler, not an installer** | Re-running is the normal case. See §3 — this is the load-bearing architectural choice. |
| **TUI now, GUI later**, over the same manifest | Keeps `curl \| bash` viable and testable headlessly; the GUI becomes a skin, not a rewrite. |
| **Userland only** — `~/.local`, no root, no pacman | The only way to survive SteamOS atomic updates. |
| **Bundles, not a fifty-item checklist** | Presenting a console buyer with a checklist defeats the point of being opinionated. |
| **We do not own the desktop session** | No global shortcuts, no panel replacement, no WM wars. See §12 — that path costs ~170 lines of fighting kglobalaccel and buys nothing. |
| **Home services (Jellyfin/RomM) deferred** | Interesting, not first. Parked in §11 with the work already done. |
| **No telemetry** | Carried from magus v0.1. Non-negotiable. |

---

## 2. The project

### One line
The first hour of owning a Steam Machine, done for you in ten minutes — and
re-runnable forever after.

### The problem
A Steam Machine arrives as a console and turns out to be a Linux PC. The owner
hits a wall the moment they leave Game Mode: no terminal they like, no browser,
no productivity apps, a stock KDE desktop, and a set of TV-specific display and
audio defaults that are quietly *wrong* — washed-out blacks from limited-range
HDMI, no CEC, a balanced power profile on a machine permanently plugged into the
wall.

Every one of those has a known fix. None is discoverable. All are lost the next
time SteamOS pushes an atomic update.

### Who it's for
The console buyer who discovers they own a PC. Not a Linux hobbyist — if the
wizard requires knowing what a display server is, it has failed.

---

## 3. The core insight: reconciler, not installer

Everything else follows from this.

The wizard's only job is to produce a **manifest** — a declarative record of what
this machine should look like. Every subsequent run reads that manifest and
converges the machine to it, at current versions, repairing whatever has drifted.
Re-running is the normal case:

- SteamOS shipped an atomic update and blew something away → re-run repairs it.
- magus itself improved → re-run pulls the better version of a step.
- An app has an update → re-run brings it forward.
- The user changed their mind → edit the manifest, converge.

**The practical consequence:** every step must be idempotent and must derive its
state from the *filesystem*, not from a record of what it did last time. A step
checks for the artifact it owns and returns early if it's already correct. There
is no "installed version" file to fall out of sync with reality — which matters
enormously on an OS that periodically deletes parts of your install behind your
back.

Recording a `version` in the manifest lets later releases run migrations,
including retiring things an earlier version installed. **"Idempotent" must mean
*converge to the current intended state*, not merely *don't crash on re-run*.**

---

## 4. The wizard

Five decisions, each with a default strong enough that holding Enter produces a
well-configured machine. `--defaults` skips all of them.

### Step 1 — Terminal
Ghostty · kitty · Alacritty · keep Konsole.
**Default: kitty** — userland install via the official installer, GPU-accelerated,
themes cleanly, proven on SteamOS.

### Step 2 — Browser
Firefox · Brave · Chrome · Vivaldi.
**Default: Firefox (flatpak).** Flatpak throughout for browsers — sandboxed,
survives updates, no root.

### Step 3 — Apps, as bundles
- **Essentials** *(on)* — archive tool, text editor, media player (mpv/VLC), image viewer.
- **Gaming extras** *(on)* — ProtonUp-Qt, Decky Loader, Heroic, Lutris, MangoHud.
  Highest-value bundle on the platform.
- **Creative** — GIMP, Krita, Inkscape, OBS.
- **Dev** — VS Code or Neovim, git, distrobox.
- **Comms** — Discord, Signal, Element.
- **Home services** — deferred, see §11.

### Step 4 — Optimisations
Mostly not questions — things we just do, with a summary shown and a
`--no-optimise` escape. **These are the differentiation.** They're high-impact,
undiscoverable, and no existing tool does them because everything in this space
was written for a handheld.

- **HDMI colour range → full RGB.** Highest-impact single fix. Limited range on a
  TV makes every black grey and nobody knows to change it.
- **HDMI-CEC.** Turn the TV on and switch input on wake. This is what makes it
  feel like a console rather than a PC under the telly.
- **Performance power profile.** A Deck balances for battery; a mains-powered
  Machine should not.
- **Audio passthrough** for AV receivers, correct sample rate.
- **Refresh rate, HDR, overscan** sanity-checked against the connected display.
- **Proton-GE** installed and selected; shader pre-caching enabled.
- **External storage** mounted at a stable path (*not* label-dependent) and
  registered as a Steam library.
- **Wired-network preference; Wi-Fi power saving off.**
- **Wake-on-LAN**, for the streaming-host case.

If we shipped only step 4, it would still be worth installing.

### Step 5 — Theming
One chosen palette drives Plasma colour scheme, icons, cursors, wallpaper, GTK
consistency (so flatpaks match), and terminal theme.

Two Steam-Machine-only touches:
- **Game Mode theming** via CSS Loader, so the console side matches the desktop.
- **The front LED bar set to the theme accent.** Nothing else on the platform can
  do this, and it's the screenshot that sells the project. Gated on §10.

---

## 5. Architecture

    magus run              # wizard (or --defaults), then converge
    magus run --defaults   # no questions, full opinionated set
    magus reconcile        # converge to existing manifest, no questions
    magus doctor           # report drift and breakage, change nothing
    magus uninstall        # reverse, restoring backups

**The wizard is a manifest builder and nothing more.** It writes
`~/.config/magus/manifest.toml`, then hands off to the reconciler. The
non-interactive path writes the same manifest from defaults. This keeps one-paste
install viable, makes the system testable without a TTY, and means the future GUI
reads and writes the same file rather than being a second implementation.

    [magus]
    version = "0.2.0"
    device  = "steam-machine"

    [choices]
    terminal = "kitty"
    browser  = "firefox"
    bundles  = ["essentials", "gaming"]
    theme    = "tokyo-night"

    [optimisations]
    hdmi_full_range = true
    cec             = true
    power_profile   = "performance"

---

## 6. Platform facts: the immutable filesystem

SteamOS 3 is A/B partitioned with a read-only root. An update writes a whole new
image to the *other* slot and reboots into it. The question is never "can I write
to `/usr`" (you can, with `steamos-readonly disable`) but "does my write exist in
the image I boot into next month." For `/usr`, no — silently, at the worst
possible time. The pacman keyring also ships unpopulated, so the package path is
hostile even before the update wipes it.

| Location | Survives update | Notes |
|---|---|---|
| `/home` | Yes | Separate partition. The safe harbour. |
| `/etc` | Mostly | 3-way merged against the new image. **Adding** new files is reliable; **modifying** files upstream also changed can lose your version. |
| `/var` | Usually | Less well-documented. Verify before depending on it. |
| `/usr`, `/opt` | **No** | Replaced wholesale. |
| Flatpak (user) | Yes | Lives in `~/.local/share/flatpak`. |

**Rule: only ever add new files to `/etc`, never edit existing ones.**

### The canonical userland layout

Proven by omasteam (§7). Copy verbatim:

| Path | Holds |
|---|---|
| `~/.local/bin/` | executables and symlinks into app dirs |
| `~/.local/<app>.app/` | self-contained third-party installs |
| `~/.local/share/magus/` | static assets, templates |
| `~/.local/state/magus/` | runtime state, PID files, temp renders |
| `~/.config/magus/` | manifest, user config, themes |
| `~/.config/autostart/*.desktop` | session autostart (desktop-bound only) |
| `~/.local/share/applications/*.desktop` | launcher entries |

Derive the state dir properly:

    STATE_DIR="${XDG_STATE_HOME:-$HOME/.local/state}/magus"

**For anything that must run outside a desktop session** (background services,
anything needed in Game Mode), use **systemd user units + `loginctl
enable-linger`**, not `~/.config/autostart`. You get restart policy, dependency
ordering, and `journalctl --user` logging, and it survives session switches.

---

## 7. Implementation conventions

Distilled from [28allday/omasteam](https://github.com/28allday/omasteam) — a
working, non-trivial userland install layer for SteamOS. Its `NOTES.md` is a
catalogue of the specific ways this goes wrong. Different product (it's a desktop
shell), but the platform mechanics transfer directly.

### Installer patterns

**Idempotency derived from the filesystem:**

```bash
if [ "$FORCE_KITTY" = 0 ] && [ -x "$HOME/.local/kitty.app/bin/kitty" ]; then
  ok "kitty already installed — skipping (use --force-kitty to reinstall)"
fi
```

```bash
if fc-match "JetBrainsMono Nerd Font Mono" | grep -qi "JetBrainsMono"; then
  ok "font already present — skipping download"; return
fi
```

**Preflight dies early on named tools; environment mismatches only warn:**

```bash
for t in curl unzip fc-cache kwriteconfig6 qdbus6 uuidgen; do
  have "$t" || die "missing required tool: $t"
done
```

**Feature flags guard each function at the top:**

```bash
install_bar() {
  [ "$WITH_BAR" = 1 ] || { ok "bar not requested — skipping"; return; }
```

**`--help` generated from the header comment**, so it can't drift:

```bash
-h|--help) awk 'NR>=3{ if($0 ~ /^#/){sub(/^# ?/,"");print} else exit }' "${BASH_SOURCE[0]}"; exit 0 ;;
```

**Four log verbs and nothing else:** `log` (section), `ok`, `warn`, `die`, plus
`have()` for command existence. 864 lines used only these.

**Back up before destroying; only destroy what you can verify is stock.**
omasteam copies the Plasma applet config aside before removing panels, and only
deletes the two desktop icons after proving they're the shipped ones —
`Return.desktop` matched on its `switch-to-game-mode` Exec, `steam.desktop` only
while it's still a symlink. A real file at either path is someone's own launcher.

**Optional components degrade instead of aborting** — check for the dependency,
`warn` and return, let the rest of the run finish.

### Runtime patterns

**Atomic state writes with a unique temp per write:**

```bash
tmp=$(mktemp "$STATE.XXXXXX") || return
{ ...build... } > "$tmp" && mv -f "$tmp" "$STATE" || rm -f "$tmp"
```

Not `"$STATE.$$"` — inside a background subshell `$$` is still the *main* shell's
PID, so concurrent writers collide. Sweep orphaned temps at startup:

```bash
find "$STATE_DIR" -maxdepth 1 -name 'state.json.*' -mmin +1 -delete
```

**JSON-escape every string you didn't generate.** SSIDs, device names, folder
paths are user-controlled; one unescaped `"` makes the state file invalid and the
UI silently freezes on stale data forever.

**Templating: `awk` + `ENVIRON`, never `sed`, never `awk -v`.** `sed` treats `&`
and `\` specially in the replacement and its delimiter collides with paths;
`awk -v` escape-processes its values. `ENVIRON` is verbatim. This bites the moment
you substitute a user-chosen path into a template.

**Subshell state goes in a file.** A function called via `$(...)` can't keep
globals — they die with the subshell.

---

## 8. The three traps that cost omasteam the most

### `pkill -f` / `pgrep -f` self-match — bit eight times

`pkill -f foo` matches *any* process whose command line contains `foo`, including
**the shell you just typed it in**. Symptom: a mysterious exit 144 as your own
shell dies.

Check argv **by position** via `/proc/<pid>/cmdline`:

```bash
readargv() { ARGV=(); [ -r "/proc/$1/cmdline" ] || return 1; mapfile -d '' -t ARGV < "/proc/$1/cmdline"; [ ${#ARGV[@]} -gt 0 ]; }

is_daemon() {
  readargv "$1" || return 1
  [ "${ARGV[0]##*/}" = "magus-daemon" ] && return 0
  [ "${ARGV[1]:-}" ] && [ "${ARGV[1]##*/}" = "magus-daemon" ]
}
```

|  | argv[0] | argv[1] |
|---|---|---|
| real daemon | `bash` | `…/magus-daemon` |
| a shell that merely *mentions* it | `bash` | `-c` |

That last row is the trick. Two follow-ons: the predicate runs against every
process on the box, so every `${ARGV[n]}` needs `:-` or `set -u` turns a
one-element command line into a fatal mid-sweep; and never relax it back to a
substring test. PID files first, `/proc` sweep second for orphans.

### The graphical session's PATH has no `~/.local/bin`

That's a `.bashrc` addition, and autostart never sources `.bashrc`. So:

- Every `.desktop` `Exec=` must be absolute: `Exec=$HOME/.local/bin/magus`,
  written via heredoc at install time.
- Any daemon shelling out to its siblings must fix its own PATH:
  `export PATH="$HOME/.local/bin:$PATH"`.
- A `TryExec` that misses makes the entry vanish **silently** rather than erroring.

### Replaying a stale command queue

If a UI queues commands for a backend to drain: use per-surface table names (a
shared name lets one drainer swallow another's commands); delete only rows you
actually ran (`WHERE id <= $last` — a blanket delete drops anything inserted
between the select and it, and a lost lock race replays everything); and **purge
the queue at daemon startup**, or a leftover `power` command from a dead session
fires the instant the backend returns.

*Note: omasteam's SQLite outbox exists because `qmlscene` can't spawn processes.
If our UI can, we don't need it — but keep the underlying idea that every action
is a named command the backend dispatches, because that's what makes §9 possible.*

### One more, from `set -e`

Under `set -e` an unguarded lookup kills the script mid-apply — omasteam's theme
tool needed `|| true` on palette queries because a theme missing one key silently
killed the whole apply. Long-running loops there use `set -uo pipefail` without
`-e` deliberately. Also: `grep -q` on a big producer under `pipefail` gives a
false negative (grep exits early → SIGPIPE → pipeline "fails"); use a count test.

---

## 9. Testing without a seat at the machine

Live-session env for anything run over SSH/tmux/cron:

    WAYLAND_DISPLAY=wayland-0 XDG_RUNTIME_DIR=/run/user/1000 DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus

**Make every wizard action expressible without the wizard** and you get a headless
test harness for free. This is the practical reason the manifest-builder split in
§5 matters.

**Know where your logs actually go before concluding "no errors."** Qt sends QML
warnings to journald, not stderr — every `2>&1` looked clean while real errors
existed, which invalidated an entire first pass of an omasteam bug sweep.

Screenshots without PIL or ImageMagick (neither is installed):

    spectacle -b -n -f -o shot.png
    ffmpeg -i shot.png -vf "crop=W:H:X:Y" crop.png

---

## 10. Verify on hardware before promising anything

Nothing below has been confirmed on a real Steam Machine. Each gates a feature in
§4.

| Unknown | Check | Gates |
|---|---|---|
| Front LED controllable from the host? | `ls /sys/class/leds/`, `ls -l /dev/hidraw*`, `lsusb -v \| grep -A5 -i valve` | LED theming |
| HDMI-CEC exposed? | `ls /dev/cec*`, check for `cec-ctl` | CEC optimisation |
| Colour range settable? | `kscreen-doctor -o`, DRM properties | Full RGB fix |
| Power profile control? | `powerprofilesctl list`, `/sys/firmware/acpi/platform_profile` | Performance profile |
| Container runtime present? | `which podman docker distrobox` | Home services (§11) |
| External storage mount stability | `lsblk -o NAME,SIZE,MOUNTPOINT,LABEL` across reboots | Steam library registration |
| `/var` persistence across A/B update | Write a marker, force an update, re-check | Where state may live |

**Fallbacks worth pre-planning:** if the LED is firmware-locked, target the Steam
Controller's LED plus an on-screen ambient overlay — same event plumbing,
different output sink. Structure that code as event sources → pluggable sinks
from day one so a locked LED costs a backend, not the feature.

---

## 11. Deferred: home services

Parked, not cancelled. The design and the step-1 artifacts already exist:

- `services/jellyfin/compose.yaml`, `services/romm/compose.yaml`,
  `services/README.md` — runnable by hand, adapted for rootless podman
  (ports above 1024, fully-qualified `docker.io/` image names, secrets in `.env`).
- Shape agreed: background containers, a very simple configurator used rarely,
  and **Steam library shortcuts straight to each app** as the daily-driver entry
  points. Jellyfin has a native TV client on Flathub; RomM is web-only, so a
  kiosk browser shortcut first, ES-DE/Steam ROM Manager as the better later path.
- Don't invent a container spec — ship a stock compose file per service and let a
  manifest describe only the settings that patch it.
- Whatever runs must yield to games: one low-priority cgroup slice, throttled
  harder while a game is running.
- Steam shortcuts are written to `shortcuts.vdf` (binary VDF, under
  `~/.steam/steam/userdata/<id>/config/`; Python's `vdf` package handles it).
  **Steam only reads it at startup**, so adding a shortcut ends with "restart
  Steam" — design the UX around that. Drop artwork into the adjacent `grid/`
  folder or entries appear as grey placeholders.

---

## 12. Non-goals

- Not a distro, image, or SteamOS replacement.
- Not a package manager — we orchestrate flatpak and userland installers.
- Not a general Linux ricing tool. Steam hardware only; that focus is the moat.
- **Does not own the desktop session.** omasteam's `configure_keybindings()` is
  ~170 of its 864 lines, almost entirely fighting kglobalaccel: shortcuts only
  grab at session start; file writes get clobbered by the running daemon's
  settings sync; keys owned by a component's `.desktop` default must be re-cleared
  *every* login via an autostart rebind. That's the tax of owning the session. If
  we ever find ourselves writing to `kglobalshortcutsrc`, we've taken a wrong turn.
- Does not modify the base OS or require root, ever.

---

## 13. Roadmap

- **v0.2 — the path exists.** Device detection, manifest, reconciler, terminal +
  browser + essentials. Proves the architecture.
- **v0.3 — the optimisations.** HDMI full range, CEC, power profile, storage.
  Where a user first says "oh, that's *better*."
- **v0.4 — theming**, including Game Mode and (if §10 allows) the LED bar.
- **v1.0 — survives an OS update in the wild**, verified on hardware, with
  `doctor` and `uninstall` both trustworthy.
- **Later —** home services bundle (§11).

## 14. Success criteria

1. Fresh Steam Machine, one paste, ten minutes, one reboot → a machine that looks
   and behaves like a considered product.
2. SteamOS ships an atomic update and breaks things → one re-run repairs it, with
   no user decisions required.
3. A user who wants none of our opinions still gets value from the optimisations
   alone.

## 15. Risks

- **Scope creep into a desktop environment.** The pull toward "and then a bar, and
  then keybindings" is strong, omasteam already does it well, and staying a
  *setup tool* is the discipline this project most needs.
- **Hardware unknowns** (§10) — don't promise what isn't verified.
- **Flatpak vs userland binaries** needs one consistent rule, not case-by-case.
- **Trust.** `curl | bash` asks for a lot. Source readable, run dry-runnable,
  every destructive act backed up first.

---

## Sources

- [28allday/omasteam](https://github.com/28allday/omasteam) — userland SteamOS
  install layer. Read `install.sh` and especially `NOTES.md`.
- [rommapp/romm](https://github.com/rommapp/romm) — `examples/docker-compose.example.yml`.
- Prior working docs: `VISION.md`, `omasteam-conventions.md`, `services/` — these
  did not travel with the brief and are not in this repo. This brief supersedes
  them; they remain as longer-form reference wherever they live.
