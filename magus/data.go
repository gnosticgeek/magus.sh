package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed commands.json
var commandsJSON []byte

// Cmd is a single installable spell.
type Cmd struct {
	ID       string   `json:"id"`
	Slug     string   `json:"slug"`
	Title    string   `json:"title"`
	Summary  string   `json:"summary"`
	Order    int      `json:"order"`
	GroupID  string   `json:"groupId"`
	Danger   string   `json:"danger"`
	DeckOnly bool     `json:"deckOnly"`
	Run      []string `json:"run"`
}

// Group is a logical bundle of commands within a stage.
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Stage is one of the five alchemical phases.
type Stage struct {
	ID      string  `json:"id"`
	Num     string  `json:"num"`
	Short   string  `json:"short"`
	Tagline string  `json:"tagline"`
	Sigil   string  `json:"sigil"`
	Groups  []Group `json:"groups"`
	Items   []*Cmd  `json:"items"`
}

// Preset is a curated bundle.
type Preset struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Tagline    string   `json:"tagline"`
	CommandIDs []string `json:"commandIds"`
}

// Catalogue is the full embedded data set.
type Catalogue struct {
	Stages  []*Stage  `json:"stages"`
	Presets []*Preset `json:"presets"`

	// Index built at load time.
	cmdByID map[string]*Cmd `json:"-"`
}

func loadCatalogue() (*Catalogue, error) {
	var c Catalogue
	if err := json.Unmarshal(commandsJSON, &c); err != nil {
		return nil, fmt.Errorf("decode commands.json: %w", err)
	}
	c.cmdByID = make(map[string]*Cmd)
	for _, st := range c.Stages {
		for _, cmd := range st.Items {
			c.cmdByID[cmd.ID] = cmd
		}
	}
	return &c, nil
}

func (c *Catalogue) CmdByID(id string) *Cmd { return c.cmdByID[id] }

func (c *Catalogue) StageByID(id string) *Stage {
	for _, st := range c.Stages {
		if st.ID == id {
			return st
		}
	}
	return nil
}

// CommandsInGroup returns commands within a stage that belong to a group.
func (c *Catalogue) CommandsInGroup(stage *Stage, groupID string) []*Cmd {
	if groupID == "" {
		return stage.Items
	}
	out := make([]*Cmd, 0, len(stage.Items))
	for _, cmd := range stage.Items {
		if cmd.GroupID == groupID {
			out = append(out, cmd)
		}
	}
	return out
}

// GroupByID looks up a group within a stage.
func (c *Catalogue) GroupByID(stage *Stage, groupID string) *Group {
	for i := range stage.Groups {
		if stage.Groups[i].ID == groupID {
			return &stage.Groups[i]
		}
	}
	return nil
}

// TotalCommands returns the total count across all stages.
func (c *Catalogue) TotalCommands() int {
	n := 0
	for _, st := range c.Stages {
		n += len(st.Items)
	}
	return n
}
