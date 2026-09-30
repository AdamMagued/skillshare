package mcp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func checkService(t *testing.T, config string) *Service {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	return &Service{ConfigPath: path, Home: dir, StateDir: filepath.Join(dir, "state"), Platform: "linux"}
}

// offlineCheck resolves every command and host and finds no variables, unless a test says otherwise.
func offlineCheck(names ...string) CheckOptions {
	return CheckOptions{
		Names:      names,
		LookupEnv:  func(string) (string, bool) { return "", false },
		LookPath:   func(name string) (string, error) { return "/usr/bin/" + name, nil },
		LookupHost: func(context.Context, string) ([]string, error) { return []string{"192.0.2.1"}, nil },
	}
}

func findings(t *testing.T, report *CheckReport, name string) []CheckFinding {
	t.Helper()
	for _, server := range report.Servers {
		if server.Name == name {
			return server.Findings
		}
	}
	t.Fatalf("no result for %s in %+v", name, report.Servers)
	return nil
}

// subjectOf is the subject of the first finding of that check, or "" when there is none.
func subjectOf(list []CheckFinding, check string) string {
	for _, f := range list {
		if f.Check == check {
			return f.Subject
		}
	}
	return ""
}

func hasFinding(list []CheckFinding, level, check, target string) bool {
	for _, f := range list {
		if f.Level == level && f.Check == check && f.Target == target {
			return true
		}
	}
	return false
}

func TestCheckMissingEnvIsErrorWithoutValue(t *testing.T) {
	s := checkService(t, "mcp:\n  targets: [claude]\n  servers:\n    gh:\n      url: https://api.example.com/mcp\n      bearerToken: {fromEnv: GH_TOKEN}\n      headers:\n        X-Team: {fromEnv: TEAM_ID}\n")
	opts := offlineCheck()
	opts.LookupEnv = func(name string) (string, bool) {
		if name == "TEAM_ID" {
			return "secret-team", true
		}
		return "", false
	}
	report, err := s.Check(opts)
	if err != nil {
		t.Fatal(err)
	}
	list := findings(t, report, "gh")
	if !hasFinding(list, "error", "env", "") || report.Servers[0].OK {
		t.Fatalf("missing GH_TOKEN must be an error: %+v", list)
	}
	if got := subjectOf(list, "env"); got != "GH_TOKEN" {
		t.Fatalf("subject = %q, want the variable name", got)
	}
	for _, f := range list {
		if strings.Contains(f.Message, "TEAM_ID") || strings.Contains(f.Message, "secret-team") {
			t.Fatalf("a set variable was reported: %+v", f)
		}
	}
}

func TestCheckCommandNotOnPath(t *testing.T) {
	s := checkService(t, "mcp:\n  targets: [claude]\n  servers:\n    local:\n      command: no-such-mcp-binary\n")
	opts := offlineCheck()
	opts.LookPath = func(string) (string, error) { return "", errors.New("not found") }
	report, err := s.Check(opts)
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(findings(t, report, "local"), "error", "command", "") {
		t.Fatalf("unresolved command must be an error: %+v", report.Servers)
	}
	if got := subjectOf(findings(t, report, "local"), "command"); got != "no-such-mcp-binary" {
		t.Fatalf("subject = %q, want the command", got)
	}
}

func TestCheckUnresolvedHostIsWarningAndSkippable(t *testing.T) {
	s := checkService(t, "mcp:\n  targets: [claude]\n  servers:\n    docs:\n      url: https://mcp.skillshare.invalid/mcp\n")
	opts := offlineCheck()
	opts.LookupHost = nil // the real resolver: .invalid never resolves
	report, err := s.Check(opts)
	if err != nil {
		t.Fatal(err)
	}
	if !hasFinding(findings(t, report, "docs"), "warning", "dns", "") || !report.Servers[0].OK {
		t.Fatalf("unresolved host must be a warning only: %+v", report.Servers)
	}
	if got := subjectOf(findings(t, report, "docs"), "dns"); got != "mcp.skillshare.invalid" {
		t.Fatalf("subject = %q, want the host", got)
	}
	opts.SkipDNS = true
	opts.LookupHost = func(context.Context, string) ([]string, error) {
		t.Fatal("--no-dns must not look up hosts")
		return nil, nil
	}
	if report, err = s.Check(opts); err != nil {
		t.Fatal(err)
	}
	if hasFinding(findings(t, report, "docs"), "warning", "dns", "") {
		t.Fatal("DNS finding reported with SkipDNS")
	}
}

func TestCheckSyncState(t *testing.T) {
	s := checkService(t, "mcp:\n  targets: [claude, cursor]\n  servers:\n    docs:\n      url: https://example.com/mcp\n")
	report, err := s.Check(offlineCheck())
	if err != nil {
		t.Fatal(err)
	}
	list := findings(t, report, "docs")
	if !hasFinding(list, "warning", "sync", "claude") || report.Summary.Warnings != 2 {
		t.Fatalf("unsynced entries must warn: %+v", list)
	}
	p, err := s.Preview()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Apply(p.Revision); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.Home, ".claude.json"), []byte(`{"mcpServers":{"docs":{"type":"http","url":"https://changed.example/mcp"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if report, err = s.Check(offlineCheck()); err != nil {
		t.Fatal(err)
	}
	list = findings(t, report, "docs")
	if !hasFinding(list, "error", "sync", "claude") || !hasFinding(list, "info", "sync", "cursor") || report.Summary.Errors != 1 {
		t.Fatalf("an edited entry must be a conflict and the other in sync: %+v", list)
	}
}

func TestCheckClientRuleNamesTarget(t *testing.T) {
	s := checkService(t, "mcp:\n  targets: [claude, cursor]\n  servers:\n    workspace:\n      url: https://example.com/mcp\n    docs:\n      url: https://example.com/docs\n")
	report, err := s.Check(offlineCheck())
	if err != nil {
		t.Fatal(err)
	}
	list := findings(t, report, "workspace")
	if !hasFinding(list, "error", "client-rule", "claude") || hasFinding(list, "error", "client-rule", "cursor") {
		t.Fatalf("Claude's reserved name must be refused for claude only: %+v", list)
	}
	if !hasFinding(list, "warning", "sync", "cursor") || hasFinding(list, "warning", "sync", "claude") {
		t.Fatalf("the accepted target must still report its sync state: %+v", list)
	}
	if !hasFinding(findings(t, report, "docs"), "warning", "sync", "claude") {
		t.Fatal("a refused server must not hide the others' sync state")
	}
	if report, err = s.Check(offlineCheck("docs")); err != nil || report.Summary.Errors != 0 {
		t.Fatalf("an unselected refused server must not fail the check: %v %+v", err, report)
	}
}

func TestCheckServerWithoutTargets(t *testing.T) {
	s := checkService(t, "mcp:\n  targets: [claude]\n  servers:\n    parked:\n      url: https://example.com/mcp\n      targets: []\n")
	report, err := s.Check(offlineCheck())
	if err != nil {
		t.Fatal(err)
	}
	list := findings(t, report, "parked")
	if len(list) != 1 || !hasFinding(list, "info", "targets", "") || !report.Servers[0].OK {
		t.Fatalf("a server without targets is kept in Skillshare only: %+v", list)
	}
}

func TestCheckUnknownNameListsKnown(t *testing.T) {
	s := checkService(t, "mcp:\n  targets: [claude]\n  servers:\n    docs:\n      url: https://example.com/mcp\n")
	_, err := s.Check(offlineCheck("missing"))
	if err == nil || !strings.Contains(err.Error(), "docs") {
		t.Fatalf("unknown name must list known servers, got %v", err)
	}
}
