package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type modernShell struct {
	Enabled  bool     `toml:"enabled"`
	Commands []string `toml:"commands"`
}
type shellOption struct{ ID, Name, Summary, Package, Script string }

var shellOptions = []shellOption{
	{"z", "z → zoxide", "Jump to frequent directories without replacing cd.", "zoxide", `eval "$(zoxide init zsh)"`},
	{"fzf-preview", "Fuzzy search with file previews", "Use fd for file search and bat for previews.", "fzf", `if (( $+commands[fd] && $+commands[bat] )); then
      export FZF_DEFAULT_COMMAND='fd --type f --hidden --exclude .git'
      export FZF_CTRL_T_COMMAND="$FZF_DEFAULT_COMMAND"
      export FZF_CTRL_T_OPTS="--preview 'bat -n --color=always {}'"
    fi`},
	{"bat-man", "Colourised manual pages", "Render man pages through bat.", "bat", `export MANPAGER="sh -c 'col -bx | bat -l man -p'"
    export MANROFFOPT="-c"`},
	{"cat", "cat → bat", "Syntax highlighting with paging disabled; some cat flags differ.", "bat", "alias cat='bat --paging=never'"},
	{"thefuck", "Correct mistyped commands", "Enable thefuck's correction command; review suggestions before running them.", "thefuck", `eval "$(thefuck --alias)"`},
	{"ls", "ls → eza", "Colourful file listings. Some ls flags differ.", "eza", "alias ls='eza --group-directories-first'"},
	{"tree", "tree → eza", "Show directory trees.", "eza", "alias tree='eza --tree'"},
	{"top", "top → btop", "Open the resource dashboard. top flags differ.", "btop", "alias top='btop'"},
	{"cd", "cd → zoxide", "Learn visited folders and jump by name.", "zoxide", `eval "$(zoxide init zsh --cmd cd)"`},
	{"fzf", "Fuzzy file and history search", "Enable fzf shortcuts and completion.", "fzf", `eval "$(fzf --zsh)"`},
	{"atuin", "Searchable shell history", "Record commands locally. Atuin takes Ctrl+R after fzf; account and sync setup are separate.", "atuin", `eval "$(atuin init zsh)"`},
	{"delta", "Readable Git diffs", "Use delta as the pager in this terminal; existing GIT_PAGER takes priority.", "git-delta", `export GIT_PAGER="${GIT_PAGER:-delta}"`},
}

func (s modernShell) validate() error {
	seen := map[string]bool{}
	for _, id := range s.Commands {
		found := false
		for _, o := range shellOptions {
			if o.ID == id {
				found = true
			}
		}
		if !found || seen[id] {
			return fmt.Errorf("unknown or duplicate modern command %q", id)
		}
		seen[id] = true
	}
	return nil
}
func (s modernShell) packages() []string {
	var ids []string
	if s.Enabled {
		for _, o := range shellOptions {
			if oneOf(o.ID, s.Commands) && !oneOf(o.Package, ids) {
				ids = append(ids, o.Package)
			}
		}
	}
	if s.Enabled && oneOf("fzf-preview", s.Commands) {
		for _, id := range []string{"fd", "bat"} {
			if !oneOf(id, ids) {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

const shellStart = "# >>> Magus modern commands >>>"
const shellEnd = "# <<< Magus modern commands <<<"

func (s modernShell) block() string {
	b := "\n" + shellStart + "\nif [[ -o interactive ]]; then\n"
	for _, o := range shellOptions {
		if !oneOf(o.ID, s.Commands) {
			continue
		}
		binary := o.Package
		if binary == "git-delta" {
			binary = "delta"
		}
		b += "  if (( $+commands[" + binary + "] )); then\n    " + o.Script + "\n  fi\n"
	}
	return b + "fi\n" + shellEnd + "\n"
}
func shellPath(p Paths) (string, error) {
	dir := os.Getenv("ZDOTDIR")
	if dir == "" {
		dir = p.Home
	}
	if !filepath.IsAbs(dir) {
		return "", fmt.Errorf("ZDOTDIR must be an absolute path")
	}
	path := filepath.Join(dir, ".zshrc")
	return path, regularTerminalPath(path)
}
func shellBackup(p Paths) string { return filepath.Join(p.State, "modern-shell-backup.json") }

type shellSnapshot struct {
	Path     string
	Original []byte
	Existed  bool
	Mode     os.FileMode
	Applied  string
	Pending  string
}

// Pending allows retry after either file write succeeds but a later write fails.
func readShell(p Paths) (string, []byte, shellSnapshot, bool, error) {
	path, err := shellPath(p)
	if err != nil {
		return "", nil, shellSnapshot{}, false, err
	}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", nil, shellSnapshot{}, false, err
	}
	snap := shellSnapshot{Path: path, Original: data, Existed: err == nil, Mode: 0600}
	if info, e := os.Stat(path); e == nil {
		snap.Mode = info.Mode().Perm()
	}
	raw, err := os.ReadFile(shellBackup(p))
	exists := err == nil
	if exists {
		err = json.Unmarshal(raw, &snap)
	}
	if err != nil && !os.IsNotExist(err) {
		return "", nil, snap, exists, err
	}
	if snap.Path != path {
		return "", nil, snap, exists, fmt.Errorf("Zsh configuration path changed; restore using the original ZDOTDIR")
	}
	current := string(data)
	if strings.Contains(current, shellStart) || strings.Contains(current, shellEnd) {
		matched := ""
		for _, block := range []string{snap.Pending, snap.Applied} {
			if block != "" && strings.Count(current, block) == 1 && strings.Count(current, shellStart) == 1 && strings.Count(current, shellEnd) == 1 {
				matched = block
				break
			}
		}
		if matched == "" {
			return "", nil, snap, exists, fmt.Errorf("Magus shell block was edited or is unrecognised; leaving it unchanged")
		}
		snap.Applied = matched
	} else {
		snap.Applied = ""
	}
	return path, data, snap, exists, nil
}

type shellStep struct{ Settings modernShell }

func (s shellStep) ID() string { return "shell:configure" }
func (s shellStep) Describe() string {
	if !s.Settings.Enabled {
		return "Disable modern terminal commands (open a new terminal afterwards)"
	}
	return "Configure modern Zsh commands (open a new terminal afterwards)"
}
func (s shellStep) Check(c *Context) (State, error) {
	if err := s.Settings.validate(); err != nil {
		return StateMissing, err
	}
	_, _, snap, exists, err := readShell(c.Paths)
	if err != nil {
		return StateMissing, err
	}
	if !s.Settings.Enabled && !exists {
		return StateOK, nil
	}
	if s.Settings.Enabled && snap.Applied == s.Settings.block() {
		return StateOK, nil
	}
	return StateMissing, nil
}
func (s shellStep) Apply(c *Context) error {
	if err := s.Settings.validate(); err != nil {
		return err
	}
	if !s.Settings.Enabled {
		return s.Remove(c)
	}
	path, data, snap, _, err := readShell(c.Paths)
	if err != nil {
		return err
	}
	if c.DryRun {
		return nil
	}
	block := s.Settings.block()
	next := string(data)
	if snap.Applied != "" {
		next = strings.Replace(next, snap.Applied, block, 1)
	} else {
		next += block
	}
	snap.Pending = block
	raw, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	if err = writeFileAtomic(shellBackup(c.Paths), raw, 0600); err != nil {
		return err
	}
	if err = writeFileAtomic(path, []byte(next), snap.Mode); err != nil {
		return err
	}
	snap.Applied, snap.Pending = block, ""
	raw, _ = json.MarshalIndent(snap, "", "  ")
	return writeFileAtomic(shellBackup(c.Paths), raw, 0600)
}
func (s shellStep) Remove(c *Context) error {
	path, data, snap, exists, err := readShell(c.Paths)
	if err != nil {
		return err
	}
	if !exists || c.DryRun {
		return nil
	}
	next := string(data)
	if snap.Applied != "" {
		next = strings.Replace(next, snap.Applied, "", 1)
	}
	if !snap.Existed && next == "" {
		err = os.Remove(path)
		if os.IsNotExist(err) {
			err = nil
		}
	} else {
		err = writeFileAtomic(path, []byte(next), snap.Mode)
	}
	if err != nil {
		return err
	}
	return os.Remove(shellBackup(c.Paths))
}
func (m *macModel) shellRows() []macRow {
	settings := m.shellSettings
	state := "Off"
	if settings.Enabled {
		state = "On"
	}
	rows := []macRow{{ID: "shell-enable", Name: "[" + state + "] Use modern commands in my terminal", Summary: "Enable selected integrations in new Zsh terminals. Review before applying."}}
	for _, o := range shellOptions {
		check := "[ ] "
		if oneOf(o.ID, settings.Commands) {
			check = "[x] "
		}
		rows = append(rows, macRow{ID: "shell-option:" + o.ID, Name: check + o.Name, Summary: o.Summary})
	}
	return append(rows, macRow{ID: "shell-review", Name: "Review & install", Summary: "Install required tools and apply your choices."}, macRow{ID: "shell-undo", Name: "Undo modern terminal commands", Summary: "Remove Magus’s shell block after review. Tools stay installed. Open a new terminal afterwards."})
}

// Track only dependencies added by this screen; preset/manual picks stay selected.
func (m *macModel) syncShellPackages() {
	if m.shellAutoPackages == nil {
		m.shellAutoPackages = map[string]bool{}
	}
	var required []string
	if m.selected["shell:configure"] {
		required = m.shellSettings.packages()
	}
	for id := range m.shellAutoPackages {
		if !oneOf(id, required) {
			delete(m.selected, id)
			delete(m.shellAutoPackages, id)
		}
	}
	for _, id := range required {
		if !m.selected[id] && m.needsSelection(id) {
			m.selected[id] = true
			m.shellAutoPackages[id] = true
		}
	}
}
