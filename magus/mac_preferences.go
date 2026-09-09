package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"howett.net/plist"
)

type preferenceValue struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}
type MacSetting struct {
	ID, Name, Summary, Domain, Key string
	Desired                        preferenceValue
}

var macSettings = []MacSetting{
	{"finder-hidden", "Show hidden files", "Reveal dotfiles and hidden folders.", "com.apple.finder", "AppleShowAllFiles", preferenceValue{"bool", "true"}},
	{"finder-extensions", "Show file extensions", "Always show the full filename extension.", "NSGlobalDomain", "AppleShowAllExtensions", preferenceValue{"bool", "true"}},
	{"finder-pathbar", "Show Finder path bar", "Show the current folder's path below each window.", "com.apple.finder", "ShowPathbar", preferenceValue{"bool", "true"}},
	{"finder-statusbar", "Show Finder status bar", "Show item counts and free space.", "com.apple.finder", "ShowStatusBar", preferenceValue{"bool", "true"}},
	{"finder-foldersfirst", "Keep folders first", "Sort folders above files.", "com.apple.finder", "_FXSortFoldersFirst", preferenceValue{"bool", "true"}},
	{"finder-searchcurrent", "Search the current folder", "Start Finder searches in the current folder.", "com.apple.finder", "FXDefaultSearchScope", preferenceValue{"string", "SCcf"}},
}

func macSetting(id string) (MacSetting, bool) {
	for _, s := range macSettings {
		if s.ID == id {
			return s, true
		}
	}
	return MacSetting{}, false
}

type preferenceSnapshot struct {
	Original preferenceValue `json:"original"`
	Applied  preferenceValue `json:"applied"`
}
type preferenceSnapshots map[string]preferenceSnapshot

func snapshotsPath(c *Context) string { return filepath.Join(c.Paths.State, "preferences.json") }
func loadSnapshots(c *Context) (preferenceSnapshots, error) {
	data, err := os.ReadFile(snapshotsPath(c))
	if os.IsNotExist(err) {
		return preferenceSnapshots{}, nil
	}
	if err != nil {
		return nil, err
	}
	var s preferenceSnapshots
	err = json.Unmarshal(data, &s)
	if s == nil && err == nil {
		err = fmt.Errorf("invalid empty preference snapshot")
	}
	return s, err
}
func saveSnapshots(c *Context, s preferenceSnapshots) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(snapshotsPath(c), data, 0600)
}
func readPreference(c *Context, s MacSetting) (preferenceValue, error) {
	out, err := c.macCommand("/usr/bin/defaults", "export", s.Domain, "-")
	if err != nil {
		if strings.Contains(out, "does not exist") || strings.Contains(out, "Domain "+s.Domain+" not found") {
			return preferenceValue{Kind: "absent"}, nil
		}
		return preferenceValue{}, err
	}
	var domain map[string]interface{}
	if _, err := plist.Unmarshal([]byte(out), &domain); err != nil {
		return preferenceValue{}, fmt.Errorf("read %s: %w", s.Domain, err)
	}
	v, ok := domain[s.Key]
	if !ok {
		return preferenceValue{Kind: "absent"}, nil
	}
	switch x := v.(type) {
	case bool:
		return preferenceValue{"bool", strconv.FormatBool(x)}, nil
	case string:
		return preferenceValue{"string", x}, nil
	case uint64:
		return preferenceValue{"int", strconv.FormatInt(int64(x), 10)}, nil
	case int64:
		return preferenceValue{"int", strconv.FormatInt(x, 10)}, nil
	case float64:
		return preferenceValue{"float", strconv.FormatFloat(x, 'g', -1, 64)}, nil
	default:
		return preferenceValue{}, fmt.Errorf("%s has an unsupported value type; leaving it unchanged", s.Key)
	}
}
func writePreference(c *Context, s MacSetting, v preferenceValue) error {
	if c.DryRun {
		return nil
	}
	var args []string
	switch v.Kind {
	case "absent":
		current, err := readPreference(c, s)
		if err != nil {
			return err
		}
		if current.Kind == "absent" {
			return nil
		}
		args = []string{"delete", s.Domain, s.Key}
	case "bool", "string", "int", "float":
		args = []string{"write", s.Domain, s.Key, "-" + v.Kind, v.Value}
	default:
		return fmt.Errorf("unsupported snapshot type %q", v.Kind)
	}
	_, err := c.macCommand("/usr/bin/defaults", args...)
	return err
}

type preferenceStep struct{ Setting MacSetting }

func (s preferenceStep) ID() string       { return "setting:" + s.Setting.ID }
func (s preferenceStep) Describe() string { return s.Setting.Name }
func (s preferenceStep) Check(c *Context) (State, error) {
	v, err := readPreference(c, s.Setting)
	if err != nil {
		return StateUnknown, err
	}
	if v == s.Setting.Desired {
		return StateOK, nil
	}
	return StateDrifted, nil
}
func (s preferenceStep) Apply(c *Context) error {
	if c.DryRun {
		return nil
	}
	snapshots, err := loadSnapshots(c)
	if err != nil {
		return err
	}
	if _, ok := snapshots[s.Setting.ID]; !ok {
		original, err := readPreference(c, s.Setting)
		if err != nil {
			return err
		}
		snapshots[s.Setting.ID] = preferenceSnapshot{original, s.Setting.Desired}
		if err := saveSnapshots(c, snapshots); err != nil {
			return fmt.Errorf("save original preference: %w", err)
		}
	}
	return writePreference(c, s.Setting, s.Setting.Desired)
}
func (s preferenceStep) Remove(c *Context) error {
	snapshots, err := loadSnapshots(c)
	if err != nil {
		return err
	}
	saved, ok := snapshots[s.Setting.ID]
	if !ok {
		return nil
	}
	current, err := readPreference(c, s.Setting)
	if err != nil {
		return err
	}
	if current != saved.Applied && current != saved.Original {
		return fmt.Errorf("%s changed outside Magus; leaving it unchanged", s.Setting.Name)
	}
	if c.DryRun {
		return nil
	}
	if current != saved.Original {
		if err := writePreference(c, s.Setting, saved.Original); err != nil {
			return err
		}
	}
	after, err := readPreference(c, s.Setting)
	if err != nil {
		return err
	}
	if after != saved.Original {
		return fmt.Errorf("restore could not be verified")
	}
	delete(snapshots, s.Setting.ID)
	return saveSnapshots(c, snapshots)
}
