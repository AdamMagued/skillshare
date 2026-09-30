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
