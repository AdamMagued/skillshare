package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skillshare/internal/backup"
	"skillshare/internal/config"
	"skillshare/internal/install"
)

func TestHandleSync_MergeMode(t *testing.T) {
	tgtPath := filepath.Join(t.TempDir(), "claude-skills")
	s, src := newTestServerWithTargets(t, map[string]string{"claude": tgtPath})
	addSkill(t, src, "alpha")

	body := `{"dryRun":false}`
	req := httptest.NewRequest(http.MethodPost, "/api/sync", strings.NewReader(body))
	rr := httptest.NewRecorder()
	s.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Results []map[string]any `json:"results"`
	}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 sync result, got %d", len(resp.Results))
	}
	if resp.Results[0]["target"] != "claude" {
		t.Errorf("expected target 'claude', got %v", resp.Results[0]["target"])
	}
}

func TestHandleSync_BacksUpTargetsFirst(t *testing.T) {
	tgtPath := filepath.Join(t.TempDir(), "claude-skills")
	s, src := newTestServerWithTargets(t, map[string]string{"claude": tgtPath})
	addSkill(t, src, "alpha")
	addSkill(t, tgtPath, "local-only")

	req := httptest.NewRequest(http.MethodPost, "/api/sync", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()
	s.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if backups, err := backup.List(); err != nil || len(backups) != 1 {
		t.Fatalf("expected one pre-sync backup, got %d (err: %v)", len(backups), err)
	}
}

func TestHandleSync_IgnoredSkillNotPrunedFromRegistry(t *testing.T) {
	tgtPath := filepath.Join(t.TempDir(), "claude-skills")
	s, src := newTestServerWithTargets(t, map[string]string{"claude": tgtPath})

	// Create a skill with install metadata (so it appears in registry)
	addSkill(t, src, "kept-skill")
	addSkillMeta(t, src, "kept-skill", "github.com/user/kept")

	// Create another skill that will be ignored
	addSkill(t, src, "ignored-skill")
	addSkillMeta(t, src, "ignored-skill", "github.com/user/ignored")

	// Add .skillignore to exclude the second skill
	os.WriteFile(filepath.Join(src, ".skillignore"), []byte("ignored-skill\n"), 0644)

	// Pre-populate store with both entries and persist to disk
	// (server auto-reloads metadata from disk on each request)
	s.skillsStore = install.NewMetadataStore()
	s.skillsStore.Set("kept-skill", &install.MetadataEntry{Source: "github.com/user/kept"})
	s.skillsStore.Set("ignored-skill", &install.MetadataEntry{Source: "github.com/user/ignored"})
	if err := s.skillsStore.Save(src); err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}

	// Run sync (non-dry-run)
	body := `{"dryRun":false}`
	req := httptest.NewRequest(http.MethodPost, "/api/sync", strings.NewReader(body))
	rr := httptest.NewRecorder()
	s.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Both entries should survive — ignored skill still exists on disk
	names := s.skillsStore.List()
	if len(names) != 2 {
		t.Fatalf("expected 2 metadata entries after sync, got %d: %v", len(names), names)
	}
}

func TestHandleSync_NoTargets(t *testing.T) {
	s, _ := newTestServer(t) // no targets configured

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/sync", strings.NewReader(body))
	rr := httptest.NewRecorder()
	s.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Results []any `json:"results"`
	}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp.Results) != 0 {
		t.Errorf("expected 0 results for no targets, got %d", len(resp.Results))
	}
}

func TestHandleSync_AgentPrunesOrphanWhenSourceEmpty(t *testing.T) {
	s, _ := newTestServer(t)

	agentSource := filepath.Join(t.TempDir(), "agents")
	agentTarget := filepath.Join(t.TempDir(), "claude-agents")
	if err := os.MkdirAll(agentSource, 0o755); err != nil {
		t.Fatalf("mkdir agent source: %v", err)
	}
	if err := os.MkdirAll(agentTarget, 0o755); err != nil {
		t.Fatalf("mkdir agent target: %v", err)
	}
	orphanPath := filepath.Join(agentTarget, "tutor.md")
	if err := os.Symlink(filepath.Join(agentSource, "tutor.md"), orphanPath); err != nil {
		t.Fatalf("seed orphan agent symlink: %v", err)
	}

	s.cfg.AgentsSource = agentSource
	s.cfg.Targets["claude"] = config.TargetConfig{
		Skills: &config.ResourceTargetConfig{Path: filepath.Join(t.TempDir(), "claude-skills")},
		Agents: &config.ResourceTargetConfig{Path: agentTarget},
	}
	if err := s.cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/sync", strings.NewReader(`{"kind":"agent"}`))
	rr := httptest.NewRecorder()
	s.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if _, err := os.Lstat(orphanPath); !os.IsNotExist(err) {
		t.Fatalf("expected orphan agent symlink to be pruned, got err=%v", err)
	}

	var resp struct {
		Results []struct {
			Target string   `json:"target"`
			Pruned []string `json:"pruned"`
		} `json:"results"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal sync response: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 sync result, got %d", len(resp.Results))
	}
	if resp.Results[0].Target != "claude" {
		t.Fatalf("expected claude target, got %q", resp.Results[0].Target)
	}
	if len(resp.Results[0].Pruned) != 1 || resp.Results[0].Pruned[0] != "tutor.md" {
		t.Fatalf("expected pruned tutor.md, got %+v", resp.Results[0].Pruned)
	}
}

func TestHandleSync_AgentsRespectTargetExcludeFilter(t *testing.T) {
	s, _ := newTestServer(t)

	agentSource := filepath.Join(t.TempDir(), "agents")
	agentTarget := filepath.Join(t.TempDir(), "claude-agents")
	if err := os.MkdirAll(agentSource, 0o755); err != nil {
		t.Fatalf("mkdir agent source: %v", err)
	}
	if err := os.MkdirAll(agentTarget, 0o755); err != nil {
		t.Fatalf("mkdir agent target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentSource, "tutor.md"), []byte("# Tutor"), 0o644); err != nil {
		t.Fatalf("write agent: %v", err)
	}

	s.cfg.AgentsSource = agentSource
	s.cfg.Targets["claude"] = config.TargetConfig{
		Skills: &config.ResourceTargetConfig{Path: filepath.Join(t.TempDir(), "claude-skills")},
		Agents: &config.ResourceTargetConfig{
			Path:    agentTarget,
			Exclude: []string{"*"},
		},
	}
	if err := s.cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/sync", strings.NewReader(`{"kind":"agent"}`))
	rr := httptest.NewRecorder()
	s.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if _, err := os.Lstat(filepath.Join(agentTarget, "tutor.md")); !os.IsNotExist(err) {
		t.Fatalf("excluded agent should not be synced, got err=%v", err)
	}
}

func TestHandleSync_AgentsPruneExcludedTargetAgent(t *testing.T) {
	s, _ := newTestServer(t)

	agentSource := filepath.Join(t.TempDir(), "agents")
	agentTarget := filepath.Join(t.TempDir(), "claude-agents")
	if err := os.MkdirAll(agentSource, 0o755); err != nil {
		t.Fatalf("mkdir agent source: %v", err)
	}
	if err := os.MkdirAll(agentTarget, 0o755); err != nil {
		t.Fatalf("mkdir agent target: %v", err)
	}
	sourceAgent := filepath.Join(agentSource, "tutor.md")
	if err := os.WriteFile(sourceAgent, []byte("# Tutor"), 0o644); err != nil {
		t.Fatalf("write agent: %v", err)
	}
	targetAgent := filepath.Join(agentTarget, "tutor.md")
	if err := os.Symlink(sourceAgent, targetAgent); err != nil {
		t.Fatalf("seed synced agent: %v", err)
	}

	s.cfg.AgentsSource = agentSource
	s.cfg.Targets["claude"] = config.TargetConfig{
		Skills: &config.ResourceTargetConfig{Path: filepath.Join(t.TempDir(), "claude-skills")},
		Agents: &config.ResourceTargetConfig{
			Path:    agentTarget,
			Exclude: []string{"*"},
		},
	}
	if err := s.cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/sync", strings.NewReader(`{"kind":"agent"}`))
	rr := httptest.NewRecorder()
	s.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if _, err := os.Lstat(targetAgent); !os.IsNotExist(err) {
		t.Fatalf("excluded synced agent should be pruned, got err=%v", err)
	}
}

func TestHandleSync_ProjectSyncsOnlyThatProjectsSkills(t *testing.T) {
	global := filepath.Join(t.TempDir(), "claude-skills")
	s, src := newTestServerWithTargets(t, map[string]string{"claude": global})
	addSkill(t, src, "team-a")
	root := filepath.Join(filepath.Dir(src), "app")
	addProject(t, s, root)

	rr := projectRequest(t, s, http.MethodPost, "/api/sync", `{"project":"`+root+`"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, ".claude", "skills", "team-a")); err != nil {
		t.Fatalf("project target not synced: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(global, "team-a")); !os.IsNotExist(err) {
		t.Fatalf("global target was synced, err=%v", err)
	}
}

func TestHandleSync_ProjectSyncsOnlyThatProjectsAgents(t *testing.T) {
	s, src := newTestServer(t)
	agentSource := filepath.Join(t.TempDir(), "agents")
	globalAgents := filepath.Join(t.TempDir(), "claude-agents")
	for _, dir := range []string{agentSource, globalAgents} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(agentSource, "tutor.md"), []byte("# Tutor"), 0o644); err != nil {
		t.Fatal(err)
	}
	s.cfg.AgentsSource = agentSource
	s.cfg.Targets["claude"] = config.TargetConfig{
		Skills: &config.ResourceTargetConfig{Path: filepath.Join(t.TempDir(), "claude-skills")},
		Agents: &config.ResourceTargetConfig{Path: globalAgents},
	}
	if err := s.cfg.Save(); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(filepath.Dir(src), "app")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"create":true,"root":"` + root + `","targets":["claude"],"agents":{"mode":"merge","include":[],"exclude":[]}}`
	if rr := projectRequest(t, s, http.MethodPut, "/api/projects", body); rr.Code != http.StatusOK {
		t.Fatalf("add project: %d %s", rr.Code, rr.Body)
	}

	rr := projectRequest(t, s, http.MethodPost, "/api/sync", `{"kind":"agent","project":"`+root+`"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if _, err := os.Lstat(filepath.Join(root, ".claude", "agents", "tutor.md")); err != nil {
		t.Fatalf("project agent not synced: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(globalAgents, "tutor.md")); !os.IsNotExist(err) {
		t.Fatalf("global agent was synced, err=%v", err)
	}
}

func TestHandleSync_UnknownProjectIsRejected(t *testing.T) {
	s, src := newTestServer(t)
	rr := projectRequest(t, s, http.MethodPost, "/api/sync", `{"project":"`+filepath.Join(filepath.Dir(src), "nope")+`"}`)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleSync_ReportsFolderConflicts(t *testing.T) {
	shared := filepath.Join(t.TempDir(), "agents-skills")
	s, src := newTestServerWithTargets(t, map[string]string{"universal": shared, "codex": shared})
	addSkill(t, src, "alpha")
	s.cfg.Targets["universal"] = config.TargetConfig{Skills: &config.ResourceTargetConfig{Path: shared, Exclude: []string{"feature-radar*"}}}
	if err := s.cfg.Save(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	for _, tc := range []struct{ method, url, body string }{
		{http.MethodPost, "/api/sync", `{"dryRun":true}`},
		{http.MethodGet, "/api/diff", ""},
	} {
		req := httptest.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
		rr := httptest.NewRecorder()
		s.handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d: %s", tc.url, rr.Code, rr.Body.String())
		}
		var resp struct {
			Warnings        []string                      `json:"warnings"`
			FolderConflicts []config.SkillsFolderConflict `json:"folder_conflicts"`
		}
		json.Unmarshal(rr.Body.Bytes(), &resp)
		if len(resp.FolderConflicts) != 1 {
			t.Fatalf("%s: expected 1 folder conflict, got %+v", tc.url, resp.FolderConflicts)
		}
		c := resp.FolderConflicts[0]
		if c.Keep != "universal" || len(c.Stop) != 1 || c.Stop[0] != "codex" {
			t.Errorf("%s: conflict = %+v, want keep universal, stop codex", tc.url, c)
		}
		for _, w := range resp.Warnings {
			if strings.Contains(w, "Skill path overlap") {
				t.Errorf("%s: overlap fully explained by the conflict, got warning %q", tc.url, w)
			}
		}
	}
}
