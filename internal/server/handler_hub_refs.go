package server

import (
	"errors"
	"net/http"

	"skillshare/internal/install"
)

type hubRefsResponse struct {
	Pinnable      bool     `json:"pinnable"`
	Current       string   `json:"current"`
	DefaultBranch string   `json:"defaultBranch"`
	Branches      []string `json:"branches"`
	Tags          []string `json:"tags"`
	Source        *string  `json:"source,omitempty"`
}

// handleHubRefs lists the branches and tags a hub entry's source can pin, and
// with "ref" in the body also returns the source rewritten to pin that ref.
func (s *Server) handleHubRefs(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	parseOpts := s.parseOpts()
	s.mu.RUnlock()

	var body struct {
		Source string  `json:"source"`
		Ref    *string `json:"ref"`
	}
	if err := decodeJSON(w, r, &body, defaultJSONBodyLimit); err != nil {
		if !errors.Is(err, errBodyTooLarge) {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
		}
		return
	}
	if body.Source == "" {
		writeError(w, http.StatusBadRequest, "source is required")
		return
	}
	source, err := install.ParseSourceWithOptions(body.Source, parseOpts)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid source: "+err.Error())
		return
	}

	resp := hubRefsResponse{Branches: []string{}, Tags: []string{}}
	if _, ok := source.PinnedRef(); !ok {
		writeJSON(w, resp)
		return
	}
	var refs install.RemoteRefs
	list, err := refs.List(source)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := refs.Settle(source); err != nil {
		writeError(w, http.StatusBadRequest, "invalid source: "+err.Error())
		return
	}
	resp.Pinnable = true
	resp.Current, _ = source.PinnedRef()
	resp.DefaultBranch = list.DefaultBranch
	resp.Branches = list.Branches
	resp.Tags = list.Tags
	if body.Ref != nil {
		pinned, err := source.AtRef(*body.Ref)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		resp.Source = &pinned
	}
	writeJSON(w, resp)
}
