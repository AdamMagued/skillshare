package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
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
// The success response always includes warnings: []string from SyncExtraFile.
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
	if err := config.ValidateImportMode(body.Mode, target, it.Import); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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

	tc := &s.cfg.Extras[i].Targets[j]
	prev := tc.Mode
	tc.Mode = body.Mode
	validationErr := config.ValidateExtraConnections(s.cfg.Extras, res.SourceDir, res.TargetDir, name)
	tc.Mode = prev
	if validationErr != nil {
		if !writeExtraTargetConflict(w, validationErr, target) {
			writeError(w, http.StatusBadRequest, validationErr.Error())
		}
		return
	}
	if tc.Mode == body.Mode {
		writeJSON(w, map[string]any{"success": true, "warnings": []string{}})
		return
	}
	args := map[string]any{"name": name, "target": target, "mode": body.Mode, "from": tc.Mode, "scope": "ui"}
	fail := func(err error) {
		s.writeOpsLog("instructions-mode", "error", start, args, err.Error())
		writeError(w, http.StatusInternalServerError, err.Error())
	}
	tc.Mode = body.Mode
	if err := config.ValidateExtraConfig(s.cfg.Extras[i]); err != nil {
		tc.Mode = prev
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := syncpkg.SyncExtraFile(instructions.ExtraFile(s.cfg.Extras[i], j, res), false, "")
	if err != nil {
		tc.Mode = prev
		fail(err)
		return
	}
	if result.Skipped > 0 {
		tc.Mode = prev
		diagnostic := strings.Join(result.Warnings, "; ")
		s.writeOpsLog("instructions-mode", "error", start, args, diagnostic)
		writeError(w, http.StatusConflict, diagnostic)
		return
	}
	if err := s.saveAndReloadConfig(); err != nil {
		fail(err)
		return
	}
	s.writeOpsLog("instructions-mode", "ok", start, args, "")
	warnings := append([]string{}, result.Warnings...)
	writeJSON(w, map[string]any{"success": true, "warnings": warnings})
}

// writeExtraTargetConflict preserves a structured conflict for UI callers.
func writeExtraTargetConflict(w http.ResponseWriter, err error, target string) bool {
	var conflict *config.ExtraTargetConflict
	if !errors.As(err, &conflict) {
		return false
	}
	writeCodedError(w, http.StatusConflict, "instructions_target_held", err.Error(), map[string]string{"name": conflict.Name, "target": target})
	return true
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
