package server

import (
	"encoding/json"
	"net/http"
	"slices"
	"time"

	"skillshare/internal/config"
	"skillshare/internal/instructions"
	syncpkg "skillshare/internal/sync"
)

// sharedInstructionsModes are the modes a shared instruction file can reach a
// target with. import needs a target that follows @path lines.
var sharedInstructionsModes = []string{"import", "symlink", "copy"}

// handlePutSharedInstructionsMode — PUT /api/instructions/{name}/targets/{target}/mode
// Changes how one attached target gets the shared file and syncs it again. The
// restore point recorded when the target was first attached stays.
func (s *Server) handlePutSharedInstructionsMode(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	name, target := r.PathValue("name"), r.PathValue("target")
	var body struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !slices.Contains(sharedInstructionsModes, body.Mode) {
		writeError(w, http.StatusBadRequest, "mode must be import, symlink, or copy")
		return
	}
	if body.Mode == "symlink" && !syncpkg.CanCreateFileLink() {
		writeError(w, http.StatusBadRequest, "file links need Windows Developer Mode; use copy")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	it, found, ok := s.targetInstructions(target)
	if !found {
		writeError(w, http.StatusNotFound, "target not found: "+target)
		return
	}
	if !ok {
		writeError(w, http.StatusBadRequest, target+" has no instruction file")
		return
	}
	if body.Mode == "import" && !it.Import {
		writeError(w, http.StatusBadRequest, target+" does not read @import lines")
		return
	}
	res := s.instructionsResolver()
	i, j := instructions.Find(s.cfg.Extras, name, it.Path, res)
	if i == -1 {
		writeError(w, http.StatusNotFound, "shared instruction file not found: "+name)
		return
	}
	if j == -1 {
		writeError(w, http.StatusNotFound, name+" is not attached to "+target)
		return
	}
	// A link or copy replaces the whole file, so it holds one shared file only.
	if body.Mode != "import" && len(instructions.Assignments(s.cfg.Extras, it.Path, res)) > 1 {
		writeError(w, http.StatusConflict, target+" uses several shared files; only import can hold more than one")
		return
	}

	tc := &s.cfg.Extras[i].Targets[j]
	if tc.Mode == body.Mode {
		writeJSON(w, map[string]any{"success": true})
		return
	}
	args := map[string]any{"name": name, "target": target, "mode": body.Mode, "from": tc.Mode, "scope": "ui"}
	fail := func(err error) {
		s.writeOpsLog("instructions-mode", "error", start, args, err.Error())
		writeError(w, http.StatusInternalServerError, err.Error())
	}
	prev := tc.Mode
	tc.Mode = body.Mode
	if err := config.ValidateExtraConfig(s.cfg.Extras[i]); err != nil {
		tc.Mode = prev
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, err := syncpkg.SyncExtraFile(instructions.ExtraFile(s.cfg.Extras[i], j, res), false, ""); err != nil {
		tc.Mode = prev
		fail(err)
		return
	}
	if err := s.saveAndReloadConfig(); err != nil {
		fail(err)
		return
	}
	s.writeOpsLog("instructions-mode", "ok", start, args, "")
	writeJSON(w, map[string]any{"success": true})
}

type sharedRestorePreview struct {
	Kind       string     `json:"kind"`
	Path       string     `json:"path"`
	Content    string     `json:"content"`           // file after restore (kind content)
	Current    string     `json:"current"`           // file now
	LinkTo     string     `json:"link_to,omitempty"` // kind link
	RecordedAt *time.Time `json:"recorded_at,omitempty"`
	// Drift: edits made after attaching are backed up, not restored.
	Drift bool `json:"drift"`
}

// handleSharedInstructionsRestorePreview — GET /api/instructions/{name}/restore-preview?target=X
// Shows what restoring one target puts back, read from the record made when
// the target was attached (see sync.RestoreExtraTarget).
func (s *Server) handleSharedInstructionsRestorePreview(w http.ResponseWriter, r *http.Request) {
	name, target := r.PathValue("name"), r.URL.Query().Get("target")
	if target == "" {
		writeError(w, http.StatusBadRequest, "target is required")
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	it, found, ok := s.targetInstructions(target)
	if !found {
		writeError(w, http.StatusNotFound, "target not found: "+target)
		return
	}
	if !ok {
		writeError(w, http.StatusBadRequest, target+" has no instruction file")
		return
	}
	res := s.instructionsResolver()
	i, j := instructions.Find(s.cfg.Extras, name, it.Path, res)
	if j == -1 {
		writeError(w, http.StatusNotFound, name+" is not attached to "+target)
		return
	}
	f := instructions.ExtraFile(s.cfg.Extras[i], j, res)
	writeJSON(w, restorePreview(f))
}

func restorePreview(f syncpkg.ExtraFile) sharedRestorePreview {
	current, _ := readLimited(f.Target)
	sp := syncpkg.PreviewRestoreExtraTarget(f)
	p := sharedRestorePreview{Kind: sp.Kind, Path: f.Target, Content: sp.Content, Current: string(current), LinkTo: sp.LinkTo, Drift: sp.Drift}
	if !sp.RecordedAt.IsZero() {
		p.RecordedAt = &sp.RecordedAt
	}
	return p
}
