//go:build !online

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skillshare/internal/testutil"
)

func TestMCPCheckGlobal(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()
	sb.WriteConfig("targets: {}\nmcp:\n  targets: [claude]\n  servers:\n    local:\n      command: sh\n      env:\n        API_KEY: {fromEnv: SKILLSHARE_CHECK_TOKEN}\n    parked:\n      url: https://example.com/mcp\n      targets: []\n")

	r := sb.RunCLI("mcp", "check", "--json", "--no-dns", "-g")
	r.AssertExitCode(t, 1)
	var report struct {
		Servers []struct {
			Name     string `json:"name"`
			OK       bool   `json:"ok"`
			Findings []struct {
				Level   string `json:"level"`
				Check   string `json:"check"`
				Target  string `json:"target"`
				Message string `json:"message"`
			} `json:"findings"`
		} `json:"servers"`
		Summary struct {
			Errors   int `json:"errors"`
			Warnings int `json:"warnings"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(r.Stdout), &report); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, r.Stdout)
	}
	if len(report.Servers) != 2 || report.Servers[0].Name != "local" || report.Servers[0].OK || !report.Servers[1].OK {
		t.Fatalf("unexpected servers: %+v", report.Servers)
	}
	if report.Summary.Errors != 1 || report.Summary.Warnings != 1 {
		t.Fatalf("want the missing variable as the only error and the unsynced entry as the only warning: %+v", report)
	}

	sb.RunCLI("sync", "mcp", "-g").AssertSuccess(t)
	r = sb.RunCLIEnv(map[string]string{"SKILLSHARE_CHECK_TOKEN": "s3cr3t-value"}, "mcp", "check", "--no-dns", "-g")
	r.AssertSuccess(t)
	r.AssertOutputContains(t, "claude: in sync")
	r.AssertOutputContains(t, "kept in Skillshare only")
	if strings.Contains(r.Stdout, "s3cr3t-value") {
		t.Fatal("check printed a variable's value")
	}

	r = sb.RunCLI("mcp", "check", "nope", "-g")
	r.AssertExitCode(t, 1)
	r.AssertOutputContains(t, "known servers: local, parked")
}

func TestMCPCheckProjects(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()
	root := filepath.Join(sb.Home, "work", "app")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	sb.WriteConfig("targets: {}\nmcp:\n  targets: [claude]\n  servers:\n    docs:\n      url: https://example.com/mcp\n  projects:\n    ~/work/app:\n      servers:\n        docs:\n          command: no-such-mcp-binary\n")

	r := sb.RunCLI("mcp", "check", "docs", "--json", "--no-dns", "-g")
	r.AssertExitCode(t, 1)
	var report struct {
		Servers []struct {
			Name    string `json:"name"`
			Project string `json:"project"`
			OK      bool   `json:"ok"`
		} `json:"servers"`
		Summary struct {
			Errors int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(r.Stdout), &report); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, r.Stdout)
	}
	if len(report.Servers) != 2 || report.Servers[0].Project != "" || !report.Servers[0].OK || report.Servers[1].Project != root || report.Servers[1].OK || report.Summary.Errors != 1 {
		t.Fatalf("want the global docs and the project's docs with its missing command: %+v", report)
	}
	if strings.Count(r.Stdout, `"project"`) != 1 {
		t.Fatalf("a global server must omit project:\n%s", r.Stdout)
	}

	r = sb.RunCLI("mcp", "check", "--no-dns", "-g")
	r.AssertExitCode(t, 1)
	r.AssertOutputContains(t, "docs  (project ~/work/app)")
	r.AssertOutputContains(t, "2 server(s) checked: 1 error(s)")
}
