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
// The success response always includes structured warnings from SyncExtraFile.
func (s *Server) handlePutSharedInstructionsMode(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	name, target := r.PathValue("name"), r.PathValue("target")
	var body struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeCodedError(w, http.StatusBadRequest, "instructions_invalid_json", "invalid JSON body", map[string]string{})
		return
	}
	if !slices.Contains(sharedInstructionsModes, body.Mode) {
		writeCodedError(w, http.StatusBadRequest, "instructions_invalid_mode", "mode must be import, symlink, or copy", map[string]string{"mode": body.Mode})
		return
	}
	if body.Mode == "symlink" && !syncpkg.CanCreateFileLink() {
		writeCodedError(w, http.StatusBadRequest, "instructions_file_links_unavailable", "file links need Windows Developer Mode; use copy", map[string]string{})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	it, found, ok := s.targetInstructions(target)
	if !found {
		writeCodedError(w, http.StatusNotFound, "instructions_target_not_found", "target not found: "+target, map[string]string{"target": target})
		return
	}
	if !ok {
		writeCodedError(w, http.StatusBadRequest, "instructions_no_file", target+" has no instruction file", map[string]string{"target": target})
		return
	}
	if err := config.ValidateImportMode(body.Mode, target, it.Import); err != nil {
		writeCodedError(w, http.StatusBadRequest, "instructions_import_unavailable", err.Error(), map[string]string{"target": target})
		return
	}
	res := s.instructionsResolver()
	i, j := instructions.Find(s.cfg.Extras, name, it.Path, res)
	if i == -1 {
		writeCodedError(w, http.StatusNotFound, "instructions_shared_not_found", "shared instruction file not found: "+name, map[string]string{"name": name})
		return
	}
	if j == -1 {
		writeCodedError(w, http.StatusNotFound, "instructions_not_attached", name+" is not attached to "+target, map[string]string{"name": name, "target": target})
		return
	}

	tc := &s.cfg.Extras[i].Targets[j]
	prev := tc.Mode
	tc.Mode = body.Mode
	validationErr := config.ValidateExtraConnections(s.cfg.Extras, res.SourceDir, res.TargetDir, name)
	tc.Mode = prev
	if validationErr != nil {
		if !writeExtraTargetConflict(w, validationErr, target) {
			writeCodedError(w, http.StatusBadRequest, "instructions_mode_failed", validationErr.Error(), map[string]string{"detail": validationErr.Error()})
		}
		return
	}
	if tc.Mode == body.Mode {
		writeJSON(w, map[string]any{"success": true, "warnings": []syncpkg.FileWarning{}})
		return
	}
	args := map[string]any{"name": name, "target": target, "mode": body.Mode, "from": tc.Mode, "scope": "ui"}
	fail := func(err error) {
		s.writeOpsLog("instructions-mode", "error", start, args, err.Error())
		writeCodedError(w, http.StatusInternalServerError, "instructions_mode_failed", err.Error(), map[string]string{"detail": err.Error()})
	}
	tc.Mode = body.Mode
	if err := config.ValidateExtraConfig(s.cfg.Extras[i]); err != nil {
		tc.Mode = prev
		writeCodedError(w, http.StatusBadRequest, "instructions_mode_failed", err.Error(), map[string]string{"detail": err.Error()})
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
		writeCodedError(w, http.StatusConflict, "instructions_target_directory", diagnostic, map[string]string{"path": it.Path})
		return
	}
	if err := s.saveAndReloadConfig(); err != nil {
		fail(err)
		return
	}
	s.writeOpsLog("instructions-mode", "ok", start, args, "")
	warnings := append([]syncpkg.FileWarning{}, result.FileWarnings...)
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
		writeCodedError(w, http.StatusBadRequest, "instructions_target_required", "target is required", map[string]string{})
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	it, found, ok := s.targetInstructions(target)
	if !found {
		writeCodedError(w, http.StatusNotFound, "instructions_target_not_found", "target not found: "+target, map[string]string{"target": target})
		return
	}
	if !ok {
		writeCodedError(w, http.StatusBadRequest, "instructions_no_file", target+" has no instruction file", map[string]string{"target": target})
		return
	}
	res := s.instructionsResolver()
	i, j := instructions.Find(s.cfg.Extras, name, it.Path, res)
	if j == -1 {
		writeCodedError(w, http.StatusNotFound, "instructions_not_attached", name+" is not attached to "+target, map[string]string{"name": name, "target": target})
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
