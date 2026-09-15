package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var testSkillFiles = map[string]agentSkillFile{
	"SKILL.md":               {Data: []byte("---\nname: frontend-design\ndescription: Test skill\n---\n"), Mode: 0644},
	"references/patterns.md": {Data: []byte("# Patterns\n"), Mode: 0644},
}

func stubSkillFetch(t *testing.T) {
	t.Helper()
	old := fetchAgentSkill
	fetchAgentSkill = func(context.Context, agentSkill) (map[string]agentSkillFile, error) { return testSkillFiles, nil }
	t.Cleanup(func() { fetchAgentSkill = old })
}

func skillPath(p Paths, id, rel string, claude bool) string {
	root := agentSkillRoots(p, id)[0]
	if claude {
		root = agentSkillRoots(p, id)[1]
	}
	return filepath.Join(root, rel)
}

func TestAgentSkillInstallsForCodexAndClaudeAndRestores(t *testing.T) {
	c := terminalTestContext(t)
	stubSkillFetch(t)
	step := agentSkillStep{IDValue: "frontend-design"}
	if state, err := step.Check(c); err != nil || state != StateMissing {
		t.Fatal(state, err)
	}
	if err := step.Apply(c); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{skillPath(c.Paths, "frontend-design", "SKILL.md", false), skillPath(c.Paths, "frontend-design", "SKILL.md", true)} {
		body, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(body, testSkillFiles["SKILL.md"].Data) {
			t.Fatalf("skill missing at %s: %v", path, err)
		}
	}
	if state, err := step.Check(c); err != nil || state != StateOK {
		t.Fatal(state, err)
	}
	if err := step.Remove(c); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{skillPath(c.Paths, "frontend-design", "SKILL.md", false), skillPath(c.Paths, "frontend-design", "SKILL.md", true)} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("owned skill was not removed from %s", path)
		}
	}
}

func TestAgentSkillPreservesExternalFilesAndDryRun(t *testing.T) {
	c := terminalTestContext(t)
	stubSkillFetch(t)
	step := agentSkillStep{IDValue: "frontend-design"}
	paths := []string{skillPath(c.Paths, "frontend-design", "SKILL.md", false), skillPath(c.Paths, "frontend-design", "SKILL.md", true)}
	if err := writeFileAtomic(paths[0], testSkillFiles["SKILL.md"].Data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := step.Apply(c); err != nil {
		t.Fatal(err)
	}
	if err := step.Remove(c); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(paths[0]); err != nil {
		t.Fatal("removed a matching skill Magus did not create")
	}
	if _, err := os.Stat(paths[1]); !os.IsNotExist(err) {
		t.Fatal("did not remove the skill Magus created")
	}

	if err := writeFileAtomic(paths[1], []byte("user version"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := step.Apply(c); err == nil {
		t.Fatal("overwrote a differing external skill")
	}
	if body, _ := os.ReadFile(paths[1]); string(body) != "user version" {
		t.Fatal("changed a differing external skill")
	}

	os.Remove(paths[0])
	os.Remove(paths[1])
	c.DryRun = true
	if err := step.Apply(c); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("dry run wrote a skill")
		}
	}
}

func TestAgentSkillMenuAndManifest(t *testing.T) {
	c := terminalTestContext(t)
	m := newMacModel(c.Paths, "", newMacManifest(), true, time.Second)
	m.screen = macScreenSkills
	for i, row := range m.rows() {
		if row.ID == "skill:frontend-design" {
			m.cursor = i
			break
		}
	}
	press(m, "enter")
	if !m.selected["skill:frontend-design"] {
		t.Fatal("skill was not selected")
	}
	manifest := m.selectionManifest()
	if len(manifest.Mac.Skills) != 1 || manifest.Mac.Skills[0] != "frontend-design" {
		t.Fatalf("skill missing from manifest: %#v", manifest.Mac.Skills)
	}
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	steps := macSteps(manifest)
	if len(steps) != 1 || steps[0].ID() != "skill:frontend-design" {
		t.Fatalf("unexpected skill plan: %#v", steps)
	}
	manifest.Mac.Skills = append(manifest.Mac.Skills, "frontend-design")
	if err := manifest.Validate(); err == nil {
		t.Fatal("duplicate skill passed validation")
	}
}

func TestAgentSkillRejectsReceiptTraversal(t *testing.T) {
	c := terminalTestContext(t)
	step := agentSkillStep{IDValue: "frontend-design"}
	outside := filepath.Join(c.Paths.Home, "outside")
	if err := writeFileAtomic(outside, []byte("outside"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(step.receipt(c.Paths), []byte(`{"Revision":"x","Files":[{"Path":"`+outside+`","SHA256":"x","Owned":true}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := step.Remove(c); err == nil {
		t.Fatal("accepted an out-of-scope receipt path")
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatal("removed out-of-scope file")
	}
}

func TestDownloadAgentSkillExtractsOnlyDeclaredDirectory(t *testing.T) {
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	for name, body := range map[string]string{
		"repo-revision/skills/frontend-design/SKILL.md":        "skill",
		"repo-revision/skills/frontend-design/references/a.md": "reference",
		"repo-revision/README.md":                              "ignore",
	} {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive.Bytes()) }))
	defer server.Close()
	files, err := downloadAgentSkill(context.Background(), agentSkill{ArchiveURL: server.URL, Repository: "test", Subdirectory: "skills/frontend-design"})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || string(files["SKILL.md"].Data) != "skill" || string(files["references/a.md"].Data) != "reference" {
		t.Fatalf("unexpected extracted files: %#v", files)
	}
}

func TestDownloadAgentSkillExtractsRepositoryRoot(t *testing.T) {
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	font := bytes.Repeat([]byte("f"), (2<<20)+1)
	for name, body := range map[string][]byte{
		"repo-revision/SKILL.md":                       []byte("skill"),
		"repo-revision/template/public/fonts/font.ttf": font,
	} {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive.Bytes()) }))
	defer server.Close()
	files, err := downloadAgentSkill(context.Background(), agentSkill{ArchiveURL: server.URL, Repository: "test", Subdirectory: "."})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || string(files["SKILL.md"].Data) != "skill" || len(files["template/public/fonts/font.ttf"].Data) != len(font) {
		t.Fatalf("unexpected extracted files: %#v", files)
	}
}
