package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func writeMCPSource(t *testing.T, s *Server, mcp string) {
	t.Helper()
	f, err := os.OpenFile(s.configPath(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(mcp); err != nil {
		t.Fatal(err)
	}
}

func TestMCPCheckReportsFindingsWithOK(t *testing.T) {
	s, _ := newTestServerWithExtras(t, nil, "")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	t.Setenv("SKILLSHARE_TEST_UNSET_TOKEN", "")
	writeMCPSource(t, s, "mcp:\n  targets: [claude]\n  servers:\n    docs:\n      url: https://mcp.skillshare.invalid/mcp\n      bearerToken: {fromEnv: SKILLSHARE_TEST_UNSET_TOKEN}\n")
	w := httptest.NewRecorder()
	s.handleMCPCheck(w, httptest.NewRequest(http.MethodGet, "/api/mcp/check?dns=0", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
	var report struct {
		Servers []struct {
			Name     string `json:"name"`
			OK       bool   `json:"ok"`
			Findings []struct {
				Level   string `json:"level"`
				Check   string `json:"check"`
				Subject string `json:"subject"`
			} `json:"findings"`
		} `json:"servers"`
		Summary struct {
			Errors int `json:"errors"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Servers) != 1 || report.Servers[0].OK || report.Summary.Errors != 1 {
		t.Fatalf("the unset token must be one error: %s", w.Body.String())
	}
	for _, f := range report.Servers[0].Findings {
		if f.Check == "dns" {
			t.Fatalf("dns=0 must skip host lookups: %s", w.Body.String())
		}
		if f.Check == "env" && f.Subject != "SKILLSHARE_TEST_UNSET_TOKEN" {
			t.Fatalf("env subject = %q", f.Subject)
		}
	}
}

func TestMCPCheckFailsWhenCheckCannotRun(t *testing.T) {
	s, _ := newTestServerWithExtras(t, nil, "")
	writeMCPSource(t, s, "mcp:\n  servers: [not, a, map]\n")
	w := httptest.NewRecorder()
	s.handleMCPCheck(w, httptest.NewRequest(http.MethodGet, "/api/mcp/check", nil))
	if w.Code < 400 || !strings.Contains(w.Body.String(), "error") {
		t.Fatalf("status = %d body = %s", w.Code, w.Body.String())
	}
}
