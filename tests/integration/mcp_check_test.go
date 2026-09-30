//go:build !online

package integration

import (
	"encoding/json"
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
