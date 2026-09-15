package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type agentSkill struct {
	ID, Name, Summary     string
	SkillsURL, Repository string
	Revision, ArchiveURL  string
	Subdirectory          string
	Advanced              bool
}

var agentSkills = []agentSkill{
	{ID: "anything2explainer", Name: "Anything2Explainer", Summary: "Turns a topic into a narrated motion-graphics explainer video.", SkillsURL: "https://github.com/Vincentwei1021/anything2explainer", Repository: "https://github.com/Vincentwei1021/anything2explainer", Revision: "5b57239578284385c72ebfb2d1fce3ab61a3950a", ArchiveURL: "https://github.com/Vincentwei1021/anything2explainer/archive/5b57239578284385c72ebfb2d1fce3ab61a3950a.tar.gz", Subdirectory: "."},
	{ID: "archify", Name: "Archify", Summary: "Creates polished, explorable architecture and workflow diagrams.", SkillsURL: "https://github.com/tt-a1i/archify", Repository: "https://github.com/tt-a1i/archify", Revision: "a07fa1d5b2a10cbea110c5a2be2817397a301cdc", ArchiveURL: "https://github.com/tt-a1i/archify/archive/a07fa1d5b2a10cbea110c5a2be2817397a301cdc.tar.gz", Subdirectory: "archify"},
	{ID: "frontend-design", Name: "Frontend Design", Summary: "Distinctive, production-grade frontend design from Anthropic.", SkillsURL: "https://skills.sh/anthropics/skills/frontend-design", Repository: "https://github.com/anthropics/skills", Revision: "34040c9c568585f6929bedeaad110ad08f079624", ArchiveURL: "https://github.com/anthropics/skills/archive/34040c9c568585f6929bedeaad110ad08f079624.tar.gz", Subdirectory: "skills/frontend-design"},
	{ID: "humanizer", Name: "Humanizer", Summary: "Rewrites AI-sounding prose while preserving its meaning and voice.", SkillsURL: "https://github.com/blader/humanizer", Repository: "https://github.com/blader/humanizer", Revision: "9862685f575c65a8247f90369951df1b3416e3d6", ArchiveURL: "https://github.com/blader/humanizer/archive/9862685f575c65a8247f90369951df1b3416e3d6.tar.gz", Subdirectory: "."},
	{ID: "systematic-debugging", Name: "Systematic Debugging", Summary: "A disciplined root-cause workflow for bugs and test failures.", SkillsURL: "https://skills.sh/obra/superpowers/systematic-debugging", Repository: "https://github.com/obra/superpowers", Revision: "b36e0829c6d0140e93cfef2ca599b1b07d4a7797", ArchiveURL: "https://github.com/obra/superpowers/archive/b36e0829c6d0140e93cfef2ca599b1b07d4a7797.tar.gz", Subdirectory: "skills/systematic-debugging"},
	{ID: "test-driven-development", Name: "Test-Driven Development", Summary: "Red-green-refactor guidance with testing references.", SkillsURL: "https://skills.sh/obra/superpowers/test-driven-development", Repository: "https://github.com/obra/superpowers", Revision: "b36e0829c6d0140e93cfef2ca599b1b07d4a7797", ArchiveURL: "https://github.com/obra/superpowers/archive/b36e0829c6d0140e93cfef2ca599b1b07d4a7797.tar.gz", Subdirectory: "skills/test-driven-development"},
	{ID: "brainstorming", Name: "Brainstorming", Summary: "Turns rough ideas into reviewed designs before implementation.", SkillsURL: "https://skills.sh/obra/superpowers/brainstorming", Repository: "https://github.com/obra/superpowers", Revision: "b36e0829c6d0140e93cfef2ca599b1b07d4a7797", ArchiveURL: "https://github.com/obra/superpowers/archive/b36e0829c6d0140e93cfef2ca599b1b07d4a7797.tar.gz", Subdirectory: "skills/brainstorming"},
	{ID: "writing-plans", Name: "Writing Plans", Summary: "Creates detailed implementation plans for unfamiliar codebases.", SkillsURL: "https://skills.sh/obra/superpowers/writing-plans", Repository: "https://github.com/obra/superpowers", Revision: "b36e0829c6d0140e93cfef2ca599b1b07d4a7797", ArchiveURL: "https://github.com/obra/superpowers/archive/b36e0829c6d0140e93cfef2ca599b1b07d4a7797.tar.gz", Subdirectory: "skills/writing-plans"},
	{ID: "resolving-merge-conflicts", Name: "Resolving Merge Conflicts", Summary: "Resolves merge conflicts carefully and verifies the result.", SkillsURL: "https://skills.sh/mattpocock/skills/resolving-merge-conflicts", Repository: "https://github.com/mattpocock/skills", Revision: "3cca18b368ae95cdbdebbff572ccafa662551015", ArchiveURL: "https://github.com/mattpocock/skills/archive/3cca18b368ae95cdbdebbff572ccafa662551015.tar.gz", Subdirectory: "skills/engineering/resolving-merge-conflicts"},
	{ID: "vercel-react-best-practices", Name: "React Best Practices", Summary: "Vercel's performance guidance for React and Next.js.", SkillsURL: "https://skills.sh/vercel-labs/agent-skills/vercel-react-best-practices", Repository: "https://github.com/vercel-labs/agent-skills", Revision: "063bee94c3f4df8453406c830b0a7df0f2860278", ArchiveURL: "https://github.com/vercel-labs/agent-skills/archive/063bee94c3f4df8453406c830b0a7df0f2860278.tar.gz", Subdirectory: "skills/react-best-practices"},
	{ID: "find-skills", Name: "Find Skills", Summary: "Discovers additional community skills through the skills CLI.", SkillsURL: "https://skills.sh/vercel-labs/skills/find-skills", Repository: "https://github.com/vercel-labs/skills", Revision: "d667282815248da03a08a18272b5d2eef9caf77c", ArchiveURL: "https://github.com/vercel-labs/skills/archive/d667282815248da03a08a18272b5d2eef9caf77c.tar.gz", Subdirectory: "skills/find-skills", Advanced: true},
}

func agentSkillByID(id string) (agentSkill, bool) {
	for _, skill := range agentSkills {
		if skill.ID == id {
			return skill, true
		}
	}
	return agentSkill{}, false
}

func agentSkillRoots(p Paths, id string) []string {
	return []string{filepath.Join(p.Home, ".agents", "skills", id), filepath.Join(p.Home, ".claude", "skills", id)}
}

type agentSkillFile struct {
	Data []byte
	Mode os.FileMode
}
type agentSkillReceiptFile struct {
	Path, SHA256 string
	Owned        bool
}
type agentSkillReceipt struct {
	Revision string
	Files    []agentSkillReceiptFile
}
type agentSkillStep struct{ IDValue string }

func (s agentSkillStep) ID() string { return "skill:" + s.IDValue }
func (s agentSkillStep) Describe() string {
	skill, _ := agentSkillByID(s.IDValue)
	return "Install shared Codex + Claude skill / " + skill.Name
}
func (s agentSkillStep) receipt(p Paths) string {
	return filepath.Join(p.State, "agent-skill-"+s.IDValue+".json")
}
func (s agentSkillStep) readReceipt(p Paths) (agentSkillReceipt, error) {
	var receipt agentSkillReceipt
	raw, err := os.ReadFile(s.receipt(p))
	if err != nil {
		return receipt, err
	}
	err = json.Unmarshal(raw, &receipt)
	return receipt, err
}
func (s agentSkillStep) validInstalledPath(p Paths, candidate string) bool {
	for _, root := range agentSkillRoots(p, s.IDValue) {
		rel, err := filepath.Rel(root, candidate)
		if err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
func fileSHA256(body []byte) string { sum := sha256.Sum256(body); return hex.EncodeToString(sum[:]) }

func (s agentSkillStep) Check(c *Context) (State, error) {
	skill, ok := agentSkillByID(s.IDValue)
	if !ok {
		return StateMissing, fmt.Errorf("unknown agent skill %q", s.IDValue)
	}
	receipt, err := s.readReceipt(c.Paths)
	if os.IsNotExist(err) {
		for _, root := range agentSkillRoots(c.Paths, s.IDValue) {
			if _, statErr := os.Lstat(root); statErr == nil {
				return StateDrifted, nil
			} else if !os.IsNotExist(statErr) {
				return StateUnknown, statErr
			}
		}
		return StateMissing, nil
	}
	if err != nil {
		return StateUnknown, err
	}
	if receipt.Revision != skill.Revision || len(receipt.Files) == 0 {
		return StateDrifted, nil
	}
	for _, file := range receipt.Files {
		if !s.validInstalledPath(c.Paths, file.Path) {
			return StateUnknown, fmt.Errorf("invalid agent skill receipt path")
		}
		if err := regularTerminalPath(file.Path); err != nil {
			return StateDrifted, err
		}
		body, readErr := os.ReadFile(file.Path)
		if os.IsNotExist(readErr) {
			return StateMissing, nil
		}
		if readErr != nil {
			return StateUnknown, readErr
		}
		if fileSHA256(body) != file.SHA256 {
			return StateDrifted, fmt.Errorf("%s changed outside Magus; it will not be overwritten", file.Path)
		}
	}
	return StateOK, nil
}

var fetchAgentSkill = downloadAgentSkill

func downloadAgentSkill(ctx context.Context, skill agentSkill) (map[string]agentSkillFile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, skill.ArchiveURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: %s", skill.Repository, resp.Status)
	}
	gz, err := gzip.NewReader(io.LimitReader(resp.Body, (32<<20)+1))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	reader, files, total := tar.NewReader(gz), map[string]agentSkillFile{}, int64(0)
	subdirectory := strings.Trim(strings.TrimSpace(skill.Subdirectory), "/")
	for {
		header, nextErr := reader.Next()
		if nextErr == io.EOF {
			break
		}
		if nextErr != nil {
			return nil, nextErr
		}
		name := strings.TrimPrefix(header.Name, "./")
		rootEnd := strings.IndexByte(name, '/')
		if rootEnd < 0 {
			continue
		}
		rel := name[rootEnd+1:]
		if subdirectory != "" && subdirectory != "." {
			prefix := subdirectory + "/"
			if !strings.HasPrefix(rel, prefix) {
				continue
			}
			rel = strings.TrimPrefix(rel, prefix)
		}
		rel = path.Clean(rel)
		if rel == "." || strings.HasPrefix(rel, "../") || path.IsAbs(rel) {
			continue
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			return nil, fmt.Errorf("skill archive contains a link or special file: %s", rel)
		}
		if len(files) >= 512 || header.Size < 0 || header.Size > 20<<20 || total+header.Size > 32<<20 {
			return nil, fmt.Errorf("skill archive exceeds safety limits")
		}
		body, readErr := io.ReadAll(io.LimitReader(reader, header.Size+1))
		if readErr != nil {
			return nil, fmt.Errorf("read %s: %w", rel, readErr)
		}
		if int64(len(body)) != header.Size {
			return nil, fmt.Errorf("read %s: unexpected size", rel)
		}
		mode := os.FileMode(header.Mode) & 0o777
		if mode&0o111 == 0 {
			mode = 0o644
		} else {
			mode = 0o755
		}
		files[filepath.FromSlash(rel)] = agentSkillFile{Data: body, Mode: mode}
		total += header.Size
	}
	if _, ok := files["SKILL.md"]; !ok {
		return nil, fmt.Errorf("%s did not contain %s/SKILL.md", skill.Repository, skill.Subdirectory)
	}
	return files, nil
}

func (s agentSkillStep) Apply(c *Context) error {
	skill, ok := agentSkillByID(s.IDValue)
	if !ok {
		return fmt.Errorf("unknown agent skill %q", s.IDValue)
	}
	if c.DryRun {
		return nil
	}
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}
	parent := c.Parent
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	files, err := fetchAgentSkill(ctx, skill)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", skill.Name, err)
	}
	type pending struct {
		path string
		file agentSkillFile
	}
	var writes []pending
	receipt := agentSkillReceipt{Revision: skill.Revision}
	previousOwned := map[string]bool{}
	if previous, readErr := s.readReceipt(c.Paths); readErr == nil {
		for _, file := range previous.Files {
			if file.Owned && s.validInstalledPath(c.Paths, file.Path) {
				previousOwned[file.Path] = true
			}
		}
	} else if !os.IsNotExist(readErr) {
		return readErr
	}
	keys := make([]string, 0, len(files))
	for rel := range files {
		keys = append(keys, rel)
	}
	sort.Strings(keys)
	for _, root := range agentSkillRoots(c.Paths, s.IDValue) {
		for _, rel := range keys {
			file := files[rel]
			destination := filepath.Join(root, rel)
			if err := regularTerminalPath(destination); err != nil {
				return err
			}
			body, readErr := os.ReadFile(destination)
			owned := previousOwned[destination]
			if readErr == nil {
				if !bytes.Equal(body, file.Data) {
					return fmt.Errorf("%s already exists and differs; Magus will not overwrite it", destination)
				}
			} else if os.IsNotExist(readErr) {
				writes = append(writes, pending{destination, file})
				owned = true
			} else {
				return readErr
			}
			receipt.Files = append(receipt.Files, agentSkillReceiptFile{Path: destination, SHA256: fileSHA256(file.Data), Owned: owned})
		}
	}
	var written []string
	for _, item := range writes {
		if err := writeFileAtomic(item.path, item.file.Data, item.file.Mode); err != nil {
			for _, p := range written {
				_ = os.Remove(p)
			}
			return err
		}
		written = append(written, item.path)
	}
	raw, err := json.Marshal(receipt)
	if err == nil {
		err = writeFileAtomic(s.receipt(c.Paths), raw, 0600)
	}
	if err != nil {
		for _, p := range written {
			_ = os.Remove(p)
		}
		return err
	}
	return nil
}

func (s agentSkillStep) Remove(c *Context) error {
	receipt, err := s.readReceipt(c.Paths)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, file := range receipt.Files {
		if !s.validInstalledPath(c.Paths, file.Path) {
			return fmt.Errorf("invalid agent skill receipt path")
		}
		body, readErr := os.ReadFile(file.Path)
		if os.IsNotExist(readErr) {
			continue
		}
		if readErr != nil {
			return readErr
		}
		if file.Owned && fileSHA256(body) != file.SHA256 {
			return fmt.Errorf("%s changed outside Magus; leaving it alone", file.Path)
		}
	}
	if c.DryRun {
		return nil
	}
	for _, file := range receipt.Files {
		if file.Owned {
			if err := os.Remove(file.Path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	return os.Remove(s.receipt(c.Paths))
}

func agentSkillRows(selected map[string]bool) []macRow {
	var rows []macRow
	for _, skill := range agentSkills {
		check := "[ ] "
		if selected["skill:"+skill.ID] {
			check = "[x] "
		}
		advanced := ""
		if skill.Advanced {
			advanced = " Advanced: this skill can discover and invoke installation of other community skills."
		}
		rows = append(rows, macRow{ID: "skill:" + skill.ID, Name: check + skill.Name + " skill", Summary: skill.Summary, Source: "Agent Skills / GitHub", Note: "Source: " + skill.SkillsURL + "\nRepository: " + skill.Repository + "\nPinned revision: " + skill.Revision + "\n\nOn confirmed apply, Magus fetches the complete skill folder and installs it in the shared ~/.agents/skills directory, plus ~/.claude/skills for Claude compatibility. Existing differing files are never overwritten." + advanced})
	}
	return rows
}
