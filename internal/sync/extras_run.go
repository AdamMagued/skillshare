package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"skillshare/internal/config"
)

// MissingSourcePolicy says what RunExtraTargets does when an extra's source
// directory does not exist.
type MissingSourcePolicy int

const (
	// MissingSourceCreate creates the directory. If that fails, no target is synced.
	MissingSourceCreate MissingSourcePolicy = iota
	// MissingSourceCreateIgnoreError creates the directory and syncs the
	// targets even when creating it fails.
	MissingSourceCreateIgnoreError
	// MissingSourceSkip leaves the directory missing and syncs no target.
	MissingSourceSkip
)

// ExtraRunOptions holds what differs between the callers of RunExtraTargets.
type ExtraRunOptions struct {
	DryRun bool
	Force  bool
	// ProjectRoot is passed to the sync; empty in global mode.
	ProjectRoot   string
	MissingSource MissingSourcePolicy
	// ResolvePath turns a configured target path into the path to sync.
	ResolvePath func(path string) string
	// ResolveExtension loads a target's transform extension by its configured value.
	ResolveExtension func(ext string) (*ExtensionSpec, error)
	// AgentTargetPaths holds cleaned agents sync target paths. Targets of the
	// extra named "agents" at these paths are skipped. Nil skips nothing.
	AgentTargetPaths map[string]bool
}

// ExtraRun is the outcome of RunExtraTargets for one extra.
type ExtraRun struct {
	SourceCreated bool  // the missing source directory was created
	SourceMissing bool  // the source directory is missing, so no target was synced
	SourceErr     error // creating the missing source directory failed
	Targets       []ExtraTargetRun
}

// ExtraTargetRun is the outcome of one target. At most one of SkippedBy,
// ModeErr/ExtensionErr (both may be set), Err and Result describes it.
type ExtraTargetRun struct {
	Target config.ExtraTargetConfig // the target as configured
	Path   string                   // the resolved target path
	// Mode is the effective configured mode when the target is skipped or
	// ModeErr is set, "copy" when only ExtensionErr is set, and the mode the
	// target synced with once the sync ran.
	Mode         string
	SkippedBy    string // "agents" when the agents sync already manages the path
	ModeErr      error  // the target has an extension but a non-copy mode
	ExtensionErr error  // the target's extension could not be resolved
	Err          error  // the sync failed
	Result       *ExtraResult
}

// RunExtraTargets syncs every target of extra from sourceDir. It only syncs and
// reports; callers print, log and count the results.
func RunExtraTargets(extra config.ExtraConfig, sourceDir string, opts ExtraRunOptions) ExtraRun {
	var run ExtraRun
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		switch opts.MissingSource {
		case MissingSourceSkip:
			run.SourceMissing = true
			return run
		case MissingSourceCreate:
			if err := os.MkdirAll(sourceDir, 0755); err != nil {
				run.SourceMissing = true
				run.SourceErr = err
				return run
			}
			run.SourceCreated = true
		case MissingSourceCreateIgnoreError:
			run.SourceCreated = os.MkdirAll(sourceDir, 0755) == nil
		}
	}

	run.Targets = make([]ExtraTargetRun, 0, len(extra.Targets))
	for _, t := range extra.Targets {
		run.Targets = append(run.Targets, runExtraTarget(extra, t, sourceDir, opts))
	}
	return run
}

func runExtraTarget(extra config.ExtraConfig, t config.ExtraTargetConfig, sourceDir string, opts ExtraRunOptions) ExtraTargetRun {
	tr := ExtraTargetRun{Target: t, Path: opts.ResolvePath(t.Path), Mode: EffectiveMode(t.Mode)}

	if extra.Name == "agents" && opts.AgentTargetPaths[filepath.Clean(tr.Path)] {
		tr.SkippedBy = "agents"
		return tr
	}

	var spec *ExtensionSpec
	if t.Extension != "" {
		mode, modeErr := ResolveExtensionMode(t.Mode)
		if modeErr == nil {
			tr.Mode = mode
		}
		tr.ModeErr = modeErr
		spec, tr.ExtensionErr = opts.ResolveExtension(t.Extension)
		if tr.ModeErr != nil || tr.ExtensionErr != nil {
			return tr
		}
	}

	tr.Result, tr.Err = SyncExtraTarget(extra, t, sourceDir, tr.Path, tr.Mode, opts.DryRun, opts.Force, opts.ProjectRoot, spec)
	tr.Mode = ExtraTargetMode(tr.Mode, extra.File != "")
	return tr
}

// SyncExtraTarget syncs one target of an extra. Single-file extras (file:)
// sync just that file to <target>/<as or file>; others sync the directory.
func SyncExtraTarget(extra config.ExtraConfig, target config.ExtraTargetConfig, sourceDir, targetPath, mode string, dryRun, force bool, projectRoot string, spec *ExtensionSpec) (*ExtraResult, error) {
	if extra.File == "" {
		return SyncExtra(sourceDir, targetPath, mode, dryRun, force, target.Flatten, projectRoot, spec)
	}
	if spec != nil {
		return nil, fmt.Errorf("extensions are not supported for single-file extras")
	}
	return SyncExtraFile(NewExtraFile(sourceDir, extra.File, targetPath, target.As, mode), dryRun, projectRoot)
}
