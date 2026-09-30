//go:build !online

package integration

import (
	"path/filepath"
	"strings"
	"testing"

	"skillshare/internal/testutil"
)

func TestMCPPiBuiltinLifecycle(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()
	sb.WriteConfig("targets: {}\n")
	sb.RunCLI("mcp", "add", "docs", "--target", "pi", "--pi-extension", "builtin", "--pi-options", `{"exposure":"deferred","custom":{"flag":true}}`, "--url", "https://example.com/mcp", "--no-tui", "-g").AssertSuccess(t)
	sb.RunCLI("sync", "mcp", "-g").AssertSuccess(t)
	path := filepath.Join(sb.Home, ".pi", "agent", "mcp.json")
	before := sb.ReadFile(path)
	if !strings.Contains(before, `"deferred"`) || strings.Contains(before, `"transport"`) {
		t.Fatal(before)
	}
	sb.RunCLI("sync", "mcp", "-g").AssertSuccess(t)
	if sb.ReadFile(path) != before {
		t.Fatal("re-sync changed native config")
	}
	sb.RunCLI("mcp", "import", "--from", "pi", "--dry-run", "--json", "-g").AssertSuccess(t)
	sb.RunCLI("mcp", "remove", "docs", "-g").AssertSuccess(t)
	sb.RunCLI("sync", "mcp", "-g").AssertSuccess(t)
	if strings.Contains(sb.ReadFile(path), `"docs"`) {
		t.Fatal("managed entry remained")
	}
}

// A server added for Pi without a mode is built-in, written to the file of its scope.
func TestMCPPiBuiltinDefaultFollowsScope(t *testing.T) {
	t.Run("global", func(t *testing.T) {
		sb := testutil.NewSandbox(t)
		defer sb.Cleanup()
		sb.WriteConfig("targets: {}\n")
		sb.RunCLI("mcp", "add", "docs", "--target", "pi", "--url", "https://example.com/mcp", "--no-tui", "-g").AssertSuccess(t)
		if !strings.Contains(sb.ReadFile(sb.ConfigPath), "piExtension: builtin") {
			t.Fatal(sb.ReadFile(sb.ConfigPath))
		}
		sb.RunCLI("sync", "mcp", "-g").AssertSuccess(t)
		if !strings.Contains(sb.ReadFile(filepath.Join(sb.Home, ".pi", "agent", "mcp.json")), `"docs"`) {
			t.Fatal("global Pi file lacks docs")
		}
	})
	t.Run("project mode", func(t *testing.T) {
		sb := testutil.NewSandbox(t)
		defer sb.Cleanup()
		sb.WriteConfig("targets: {}\n")
		root := sb.SetupProjectDir()
		sb.RunCLIInDir(root, "mcp", "add", "docs", "--target", "pi", "--url", "https://example.com/mcp", "--no-tui", "-p").AssertSuccess(t)
		if !strings.Contains(sb.ReadFile(filepath.Join(root, ".skillshare", "config.yaml")), "piExtension: builtin") {
			t.Fatal(sb.ReadFile(filepath.Join(root, ".skillshare", "config.yaml")))
		}
		sb.RunCLIInDir(root, "sync", "mcp", "-p").AssertSuccess(t)
		if !strings.Contains(sb.ReadFile(filepath.Join(root, ".pi", "mcp.json")), `"docs"`) {
			t.Fatal("project Pi file lacks docs")
		}
		if sb.FileExists(filepath.Join(sb.Home, ".pi", "agent", "mcp.json")) {
			t.Fatal("project server reached the global Pi file")
		}
	})
	t.Run("mcp.projects", func(t *testing.T) {
		sb := testutil.NewSandbox(t)
		defer sb.Cleanup()
		root := filepath.Join(sb.Root, "work")
		sb.WriteConfig("targets: {}\nmcp:\n  projects:\n    " + root + ":\n      servers:\n        docs:\n          url: https://example.com/mcp\n          targets: [pi]\n          piExtension: builtin\n")
		sb.RunCLI("sync", "mcp", "-g").AssertSuccess(t)
		if !strings.Contains(sb.ReadFile(filepath.Join(root, ".pi", "mcp.json")), `"docs"`) {
			t.Fatal("mcp.projects Pi file lacks docs")
		}
		if sb.FileExists(filepath.Join(sb.Home, ".pi", "agent", "mcp.json")) {
			t.Fatal("project server reached the global Pi file")
		}
	})
}
