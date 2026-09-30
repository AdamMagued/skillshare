package main

import (
	"os"
	"path/filepath"
	"skillshare/internal/mcp"
	"strings"
	"testing"
)

func TestMCPPiImportAutodetectedFileUsesSelectedMode(t *testing.T) {
	for _, mode := range []string{"builtin", "pi-mcp-adapter", "pi-mcp-extension"} {
		t.Run(mode, func(t *testing.T) {
			s := mcpTUIService(t)
			path := filepath.Join(s.Home, "export.json")
			if err := os.WriteFile(path, []byte(`{"mcpServers":{"a":{"command":"c","exposure":"direct"}}}`), 0600); err != nil {
				t.Fatal(err)
			}
			o, err := parseMCPOptions([]string{"a", "--file", path, "--pi-extension", mode, "--target", "claude", "--no-tui"})
			if err != nil {
				t.Fatal(err)
			}
			output := captureStdout(t, func() { err = runMCPImport(s, o) })
			if err != nil {
				t.Fatal(err)
			}
			source, err := mcp.LoadSource(s.ConfigPath)
			if err != nil || source.Servers["a"].PiExtension != mode {
				t.Fatalf("%+v %v", source, err)
			}
			if mode == "pi-mcp-extension" && (len(source.Servers["a"].PiOptions) != 0 || !strings.Contains(output, "Agent-specific field not imported: exposure")) {
				t.Fatalf("extension options or warning: %+v %s", source.Servers["a"], output)
			}
			if mode == "pi-mcp-adapter" && !strings.Contains(output, "Pi built-in field") {
				t.Fatalf("adapter mode silently accepted exposure: %s", output)
			}
		})
	}
}

func TestMCPPiImportAutodetectedFileRejectsOtherFormats(t *testing.T) {
	s := mcpTUIService(t)
	path := filepath.Join(s.Home, "gemini.json")
	if err := os.WriteFile(path, []byte(`{"mcpServers":{"a":{"httpUrl":"https://example.com/mcp"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	o, err := parseMCPOptions([]string{"a", "--file", path, "--pi-extension", "builtin", "--no-tui"})
	if err != nil {
		t.Fatal(err)
	}
	if err = runMCPImport(s, o); err == nil || !strings.Contains(err.Error(), "detected gemini") {
		t.Fatalf("non-Pi input accepted: %v", err)
	}
	source, err := mcp.LoadSource(s.ConfigPath)
	if err != nil || source.Servers["a"].Command != "" || source.Servers["a"].URL != "" {
		t.Fatalf("source changed: %+v %v", source, err)
	}
}

func TestMCPPiImportSelectsModeFile(t *testing.T) {
	for _, mode := range []string{"builtin", "pi-mcp-adapter", "pi-mcp-extension"} {
		t.Run(mode, func(t *testing.T) {
			s := mcpTUIService(t)
			dir := filepath.Join(s.Home, ".pi", "agent")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			for file, entry := range map[string]string{"mcp.json": `{"command":"native"}`, "mcp-adapter.json": `{"command":"adapter","directTools":true}`} {
				if err := os.WriteFile(filepath.Join(dir, file), []byte(`{"mcpServers":{"docs":`+entry+`}}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			o, err := parseMCPOptions([]string{"docs", "--from", "pi", "--pi-extension", mode, "--target", "claude", "--no-tui"})
			if err != nil {
				t.Fatal(err)
			}
			if err := runMCPImport(s, o); err != nil {
				t.Fatal(err)
			}
			source, err := mcp.LoadSource(s.ConfigPath)
			want := "native"
			if mode == "pi-mcp-adapter" {
				want = "adapter"
			}
			if err != nil || source.Servers["docs"].Command != want || source.Servers["docs"].PiExtension != mode {
				t.Fatalf("%+v %v", source, err)
			}
		})
	}
}

func TestMCPPiPruneCanBeDisabled(t *testing.T) {
	s := mcpTUIService(t)
	o, err := parseMCPOptions([]string{"docs", "--target", "pi", "--pi-extension", "builtin", "--pi-options-prune", "--url", "https://example.com/mcp", "--no-tui"})
	if err != nil {
		t.Fatal(err)
	}
	if err = runMCPAdd(s, o); err != nil {
		t.Fatal(err)
	}
	o, err = parseMCPOptions([]string{"docs", "--pi-options-prune=false", "--no-tui"})
	if err != nil {
		t.Fatal(err)
	}
	if err = runMCPEdit(s, o); err != nil {
		t.Fatal(err)
	}
	source, err := mcp.LoadSource(s.ConfigPath)
	if err != nil || source.Servers["docs"].PiOptionsPrune {
		t.Fatalf("%+v %v", source, err)
	}
}

func TestMCPPiImportExportedExtensionFile(t *testing.T) {
	s := mcpTUIService(t)
	path := filepath.Join(s.Home, "export.json")
	if err := os.WriteFile(path, []byte(`{"mcpServers":{"docs":{"command":"docs","lifecycle":"eager"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	o, err := parseMCPOptions([]string{"docs", "--from", "pi", "--file", path, "--pi-extension", "pi-mcp-extension", "--target", "claude", "--no-tui"})
	if err != nil {
		t.Fatal(err)
	}
	if err = runMCPImport(s, o); err != nil {
		t.Fatal(err)
	}
	source, err := mcp.LoadSource(s.ConfigPath)
	if err != nil || source.Servers["docs"].PiExtension != "pi-mcp-extension" || source.Servers["docs"].PiOptions != nil {
		t.Fatalf("%+v %v", source, err)
	}
}

func TestMCPImportFromOtherAgentCanChoosePiMode(t *testing.T) {
	s := mcpTUIService(t)
	if err := os.WriteFile(filepath.Join(s.Home, ".claude.json"), []byte(`{"mcpServers":{"docs":{"command":"docs"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	o, err := parseMCPOptions([]string{"docs", "--from", "claude", "--pi-extension", "builtin", "--target", "pi", "--no-tui"})
	if err != nil {
		t.Fatal(err)
	}
	if err = runMCPImport(s, o); err != nil {
		t.Fatal(err)
	}
}

func TestMCPAddDefaultsPiToBuiltin(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		args       []string
	}{
		{"explicit target", "builtin", []string{"--target", "pi"}},
		{"inherited target", "builtin", nil},
		{"chosen mode kept", "pi-mcp-adapter", []string{"--target", "pi", "--pi-extension", "pi-mcp-adapter"}},
		{"no Pi", "", []string{"--target", "claude"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := mcpTUIService(t)
			if err := os.WriteFile(s.ConfigPath, []byte("mcp: {targets: [pi], servers: {}}\n"), 0600); err != nil {
				t.Fatal(err)
			}
			o, err := parseMCPOptions(append([]string{"docs", "--url", "https://example.com/mcp", "--no-tui"}, tc.args...))
			if err != nil {
				t.Fatal(err)
			}
			if err = runMCPAdd(s, o); err != nil {
				t.Fatal(err)
			}
			source, err := mcp.LoadSource(s.ConfigPath)
			if err != nil || source.Servers["docs"].PiExtension != tc.want {
				t.Fatalf("%+v %v", source.Servers["docs"], err)
			}
		})
	}
}

func TestMCPImportIntoPiDefaultsToBuiltin(t *testing.T) {
	s := mcpTUIService(t)
	if err := os.WriteFile(filepath.Join(s.Home, ".claude.json"), []byte(`{"mcpServers":{"docs":{"command":"docs"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	o, err := parseMCPOptions([]string{"docs", "--from", "claude", "--target", "pi", "--no-tui"})
	if err != nil {
		t.Fatal(err)
	}
	if err = runMCPImport(s, o); err != nil {
		t.Fatal(err)
	}
	source, err := mcp.LoadSource(s.ConfigPath)
	if err != nil || source.Servers["docs"].PiExtension != "builtin" {
		t.Fatalf("%+v %v", source.Servers["docs"], err)
	}
}

func TestMCPPiCLI(t *testing.T) {
	s := mcpTUIService(t)
	o, err := parseMCPOptions([]string{"docs", "--target", "pi", "--pi-extension", "pi-mcp-extension", "--url", "https://example.com/mcp", "--no-tui"})
	if err != nil {
		t.Fatal(err)
	}
	if err = runMCPAdd(s, o); err != nil {
		t.Fatal(err)
	}
	source, err := mcp.LoadSource(s.ConfigPath)
	if err != nil || source.Servers["docs"].PiExtension != "pi-mcp-extension" {
		t.Fatalf("%+v %v", source, err)
	}
	o, err = parseMCPOptions([]string{"docs", "--pi-extension", "pi-mcp-adapter", "--no-tui"})
	if err != nil {
		t.Fatal(err)
	}
	if err = runMCPEdit(s, o); err != nil {
		t.Fatal(err)
	}
	source, err = mcp.LoadSource(s.ConfigPath)
	if err != nil || source.Servers["docs"].PiExtension != "pi-mcp-adapter" {
		t.Fatalf("%+v %v", source, err)
	}
}

type piPrompts struct{ scriptedMCPPrompts }

func (p *piPrompts) choose(c checklistConfig) ([]int, error) {
	for i, item := range c.items {
		if item.label == "pi" || item.label == "pi-mcp-extension" {
			return []int{i}, nil
		}
	}
	return nil, errMCPCancelled
}
func TestMCPPiTUI(t *testing.T) {
	s := mcpTUIService(t)
	servers := []mcp.Server{{Command: "echo"}, {URL: "https://example.com/mcp"}}
	targets, err := chooseMCPTargets(s, servers, nil, &piPrompts{})
	if err != nil || len(targets) != 1 || targets[0] != "pi" {
		t.Fatalf("%v %v", targets, err)
	}
	for _, server := range servers {
		if server.PiExtension != "pi-mcp-extension" {
			t.Fatal("extension selection lost")
		}
	}
}

func TestMCPDirectToolsFlag(t *testing.T) {
	s := mcpTUIService(t)
	run := func(handler func(*mcp.Service, mcpOptions) error, args ...string) any {
		t.Helper()
		o, err := parseMCPOptions(append(args, "--no-tui"))
		if err != nil {
			t.Fatal(err)
		}
		if err = handler(s, o); err != nil {
			t.Fatal(err)
		}
		source, err := mcp.LoadSource(s.ConfigPath)
		if err != nil {
			t.Fatal(err)
		}
		return source.Servers["tools"].DirectTools
	}
	if got := run(runMCPAdd, "tools", "--target", "pi", "--pi-extension", "pi-mcp-adapter", "--direct-tools", "true", "--url", "https://example.com/mcp"); got != true {
		t.Fatalf("add: %v", got)
	}
	if got, ok := run(runMCPEdit, "tools", "--direct-tools", "search_docs,fetch").([]any); !ok || len(got) != 2 || got[1] != "fetch" {
		t.Fatalf("edit to a list: %v", got)
	}
	if got := run(runMCPEdit, "tools", "--direct-tools", "search"); got != "search" {
		t.Fatalf("edit to search: %v", got)
	}
}

func TestMCPPiOptionsFlag(t *testing.T) {
	s := mcpTUIService(t)
	run := func(handler func(*mcp.Service, mcpOptions) error, args ...string) map[string]any {
		t.Helper()
		o, err := parseMCPOptions(append(args, "--no-tui"))
		if err != nil {
			t.Fatal(err)
		}
		if err = handler(s, o); err != nil {
			t.Fatal(err)
		}
		source, err := mcp.LoadSource(s.ConfigPath)
		if err != nil {
			t.Fatal(err)
		}
		return source.Servers["tools"].PiOptions
	}
	if got := run(runMCPAdd, "tools", "--target", "pi", "--pi-extension", "pi-mcp-adapter", "--pi-options", `{"excludeTools":["*emulator*"]}`, "--url", "https://example.com/mcp"); len(got) != 1 {
		t.Fatalf("add: %v", got)
	}
	if got := run(runMCPEdit, "tools", "--pi-options", `{"approveTools":["delete_*"]}`); len(got) != 1 || got["approveTools"] == nil {
		t.Fatalf("edit replaces the options: %v", got)
	}
	if got := run(runMCPEdit, "tools", "--pi-options", `{}`); len(got) != 0 {
		t.Fatalf("an empty object clears them: %v", got)
	}
	if _, err := parseMCPOptions([]string{"tools", "--pi-options", `["a"]`}); err == nil {
		t.Fatal("a JSON array was accepted")
	}
}
