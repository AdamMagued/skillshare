package server

import (
	"net/http"

	"skillshare/internal/mcp"
)

// handleMCPCheck answers `skillshare mcp check --json` for the dashboard. It is read-only
// and writes no operation log entry. Findings that are errors still return 200; only a
// check that cannot run fails the request. ?dns=0 skips host lookups.
func (s *Server) handleMCPCheck(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	report, err := s.mcpService().Check(mcp.CheckOptions{SkipDNS: r.URL.Query().Get("dns") == "0"})
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	writeJSON(w, report)
}
