package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPPiImportSourceSelection(t *testing.T) {
	s, sourceDir := newTestServerWithExtras(t, nil, "")
	home, agent, root, account := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PI_CODING_AGENT_DIR", agent)
	data := []byte("source: " + sourceDir + "\ntargets:\n  pi-work: {agent: pi, config_dir: " + account + "}\nmcp:\n  projects:\n    " + root + ": {}\n")
	if err := os.WriteFile(s.configPath(), data, 0600); err != nil {
		t.Fatal(err)
	}
	list := httptest.NewRecorder()
	s.handleMCPList(list, httptest.NewRequest(http.MethodGet, "/api/mcp", nil))
	var view struct {
		ImportSources map[string][]struct{ Target, Path, PiExtension string }
	}
	if list.Code != 200 || json.Unmarshal(list.Body.Bytes(), &view) != nil {
		t.Fatal(list.Body.String())
	}
	for _, scope := range []string{"", root} {
		for _, target := range []string{"pi", "pi-work"} {
			base := agent
			if target == "pi-work" {
				base = account
			}
			if scope != "" {
				base = filepath.Join(root, ".pi")
			}
			found := map[string]string{}
			for _, source := range view.ImportSources[scope] {
				if source.Target == target {
					found[source.PiExtension] = source.Path
				}
			}
			if scope != "" && target == "pi-work" {
				if len(found) != 0 {
					t.Fatal("project offered an account")
				}
				continue
			}
			for mode, file := range map[string]string{"builtin": "mcp.json", "pi-mcp-adapter": "mcp-adapter.json", "pi-mcp-extension": "mcp.json"} {
				if mode == "pi-mcp-extension" && target == "pi-work" {
					if _, exists := found[mode]; exists {
						t.Fatal("extension offered an account")
					}
					continue
				}
				modeBase := base
				if mode == "pi-mcp-extension" && scope == "" {
					modeBase = filepath.Join(home, ".pi", "agent")
				}
				if found[mode] != filepath.Join(modeBase, file) {
					t.Fatalf("%s %s %s: %+v", scope, target, mode, found)
				}
			}
			if err := os.MkdirAll(base, 0755); err != nil {
				t.Fatal(err)
			}
			for file, command := range map[string]string{"mcp.json": "native", "mcp-adapter.json": "adapter"} {
				if err := os.WriteFile(filepath.Join(base, file), []byte(`{"mcpServers":{"docs":{"command":"`+command+`"}}}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if scope == "" && target == "pi" {
				extensionDir := filepath.Join(home, ".pi", "agent")
				if err := os.MkdirAll(extensionDir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(extensionDir, "mcp.json"), []byte(`{"mcpServers":{"docs":{"command":"native","transport":"stdio","lifecycle":"eager"}}}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			for mode, command := range map[string]string{"builtin": "native", "pi-mcp-adapter": "adapter", "pi-mcp-extension": "native"} {
				if mode == "pi-mcp-extension" && target == "pi-work" {
					continue
				}
				body, _ := json.Marshal(map[string]string{"from": target, "root": scope, "piExtension": mode})
				w := httptest.NewRecorder()
				s.handleMCPImport(w, httptest.NewRequest(http.MethodPost, "/api/mcp/import", strings.NewReader(string(body))))
				var result struct {
					Candidates []struct {
						From     string
						Warnings []string
						Server   struct{ Command, PiExtension string }
					}
				}
				if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || len(result.Candidates) != 1 || result.Candidates[0].From != target || result.Candidates[0].Server.Command != command || result.Candidates[0].Server.PiExtension != mode || strings.Contains(strings.Join(result.Candidates[0].Warnings, ";"), "Select the Pi MCP mode") {
					t.Fatalf("%s %s %s: %s", scope, target, mode, w.Body)
				}
			}
		}
	}
	// An exact source must neither fall back to the other file nor accept non-Pi modes.
	if err := os.Remove(filepath.Join(agent, "mcp.json")); err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{`{"from":"pi","piExtension":"builtin"}`, `{"from":"pi","piExtension":"unknown"}`, `{"from":"claude","piExtension":"builtin"}`} {
		w := httptest.NewRecorder()
		s.handleMCPImport(w, httptest.NewRequest(http.MethodPost, "/api/mcp/import", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatalf("accepted %s: %s", body, w.Body)
		}
	}
}

func TestMCPAPIPreviewAndApply(t *testing.T) {
	s, _ := newTestServerWithExtras(t, nil, "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	body := `{"mutation":{"name":"docs","server":{"url":"https://example.com/mcp","targets":["claude"]}}}`
	preview := httptest.NewRecorder()
	s.handleMCPPreview(preview, httptest.NewRequest(http.MethodPost, "/api/mcp/preview", strings.NewReader(body)))
	if preview.Code != 200 {
		t.Fatalf("preview: %s", preview.Body.String())
	}
	var p struct {
		Revision string `json:"revision"`
	}
	if err := json.Unmarshal(preview.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude.json")); !os.IsNotExist(err) {
		t.Fatal("preview wrote native file")
	}
	apply := httptest.NewRecorder()
	request := strings.TrimSuffix(body, "}") + `,"sync":true,"revision":"` + p.Revision + `"}`
	s.handleMCPConfigure(apply, httptest.NewRequest(http.MethodPost, "/api/mcp", strings.NewReader(request)))
	if apply.Code != 200 {
		t.Fatalf("apply: %s", apply.Body.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".claude.json")); err != nil {
		t.Fatal(err)
	}
}

func TestMCPListIgnoresUnresolvableUnusedTarget(t *testing.T) {
	s, _ := newTestServerWithExtras(t, nil, "")
	xdg := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", xdg)
	for _, name := range []string{"opencode.json", "opencode.jsonc"} {
		path := filepath.Join(xdg, "opencode", name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	w := httptest.NewRecorder()
	s.handleMCPList(w, httptest.NewRequest(http.MethodGet, "/api/mcp", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
}

func TestMCPRoutesRejectRebindingHost(t *testing.T) {
	s, _ := newTestServerWithExtras(t, nil, "")
	s.addr = "127.0.0.1:19420"
	routes := []string{"GET /api/mcp", "GET /api/mcp/check", "POST /api/mcp", "POST /api/mcp/preview", "POST /api/mcp/import", "POST /api/mcp/restore"}
	for _, route := range routes {
		method, path, _ := strings.Cut(route, " ")
		req := httptest.NewRequest(method, path, strings.NewReader(`{}`))
		req.RemoteAddr = "127.0.0.1:5555"
		req.Host = "attacker.example:19420"
		w := httptest.NewRecorder()
		s.mux.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("%s: status = %d, want 403", route, w.Code)
		}
	}
}

func TestMCPPiBuiltinProjectRenderAndPaths(t *testing.T) {
	s, _ := newTestServerWithExtras(t, nil, "")
	home, root := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PI_CODING_AGENT_DIR", "")
	configPath := s.configPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, []byte("\nmcp:\n  targets: [pi]\n  servers:\n    docs:\n      command: docs\n      piExtension: builtin\n  projects:\n    "+root+":\n      servers: {}\n")...)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"mutation": map[string]any{"project": root, "name": "docs", "server": map[string]any{"command": "docs", "piExtension": "builtin", "targets": []string{"pi"}, "piOptions": map[string]any{"exposure": "deferred"}}}})
	result := httptest.NewRecorder()
	s.handleMCPRender(result, httptest.NewRequest(http.MethodPost, "/api/mcp/render", strings.NewReader(string(body))))
	var rendered struct {
		Rendered []struct {
			Path    string
			Content string
			Error   string
		}
	}
	if result.Code != 200 || json.Unmarshal(result.Body.Bytes(), &rendered) != nil || len(rendered.Rendered) != 1 || rendered.Rendered[0].Path != filepath.Join(root, ".pi", "mcp.json") || rendered.Rendered[0].Error != "" {
		t.Fatalf("wrong project render: %s", result.Body)
	}
	if _, err := os.Stat(rendered.Rendered[0].Path); !os.IsNotExist(err) {
		t.Fatal("render wrote native config")
	}
	list := httptest.NewRecorder()
	s.handleMCPList(list, httptest.NewRequest(http.MethodGet, "/api/mcp", nil))
	var view struct{ Paths map[string]string }
	if list.Code != 200 || json.Unmarshal(list.Body.Bytes(), &view) != nil || view.Paths["pi"] != filepath.Join(home, ".pi", "agent", "mcp.json") {
		t.Fatalf("wrong mode path label: %s", list.Body)
	}
}

func TestMCPRenderRejectsProjectOverrideInProjectMode(t *testing.T) {
	s, _ := newTestServerWithExtras(t, nil, "")
	s.projectRoot = t.TempDir()
	body := `{"mutation":{"project":"/other/project","name":"docs","server":{"command":"docs","targets":["pi"],"piExtension":"builtin"}}}`
	w := httptest.NewRecorder()
	s.handleMCPRender(w, httptest.NewRequest(http.MethodPost, "/api/mcp/render", strings.NewReader(body)))
	if w.Code != 400 || !strings.Contains(w.Body.String(), "project mode") {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
}
