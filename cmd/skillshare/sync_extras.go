package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"skillshare/internal/config"
	"skillshare/internal/oplog"
	"skillshare/internal/sync"
	"skillshare/internal/ui"
)

// extrasAgentsName is the extras entry name that may overlap with the agents sync system.
const extrasAgentsName = "agents"

type syncExtrasJSONOutput struct {
	Extras   []syncExtrasJSONEntry `json:"extras"`
	Duration string                `json:"duration"`
}

type syncExtrasJSONEntry struct {
	Name    string                 `json:"name"`
	Targets []syncExtrasJSONTarget `json:"targets"`
}

type syncExtrasJSONTarget struct {
	Path      string   `json:"path"`
	Mode      string   `json:"mode"`
	Synced    int      `json:"synced"`
	Skipped   int      `json:"skipped"`
	Pruned    int      `json:"pruned"`
	Error     string   `json:"error,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
	SkippedBy string   `json:"skipped_by,omitempty"`
}

func cmdSyncExtras(args []string) error {
	start := time.Now()

	mode, rest, err := parseModeArgs(args)
	if err != nil {
		return err
	}

	dryRun, force, jsonOutput, _ := parseSyncFlags(rest)

	cwd, _ := os.Getwd()
	mode = resolveAutoMode(mode, cwd)

	applyModeLabel(mode)

	if mode == modeProject {
		return cmdSyncExtrasProject(cwd, dryRun, force, jsonOutput, start)
	}
	return cmdSyncExtrasGlobal(dryRun, force, jsonOutput, start)
}

func cmdSyncExtrasGlobal(dryRun, force, jsonOutput bool, start time.Time) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, err := config.ValidateConfig(cfg); err != nil {
		return err
	}
	if len(cfg.Extras) == 0 {
		// Clean up empty extras directory
		removeEmptyDir(config.ExtrasParentDir(cfg.EffectiveSkillsSource()))

		if jsonOutput {
			return writeJSON(&syncExtrasJSONOutput{Extras: []syncExtrasJSONEntry{}, Duration: formatDuration(start)})
		}
		ui.Info("No extras configured.")
		fmt.Println()
		ui.Info("Add extras to your config.yaml:")
		fmt.Println()
		fmt.Println("  extras:")
		fmt.Println("    - name: rules")
		fmt.Println("      targets:")
		fmt.Println("        - path: ~/.claude/rules")
		fmt.Println("        - path: ~/.cursor/rules")
		fmt.Println("          mode: copy")
		return nil
	}

	configDir := filepath.Dir(cfg.EffectiveSkillsSource())

	// Auto-migrate legacy extras directories (flat → extras/<name>/)
	if warnings := config.MigrateExtrasDir(configDir, cfg.Extras); len(warnings) > 0 {
		for _, w := range warnings {
			ui.Warning(w)
		}
	}

	if dryRun && !jsonOutput {
		ui.Warning("Dry run mode - no changes will be made")
	}

	// Detect overlap between extras "agents" and the agents sync system
	var agentTargetPaths map[string]bool
	for _, extra := range cfg.Extras {
		if extra.Name == extrasAgentsName {
			agentTargetPaths = collectAgentTargetPathsGlobal(cfg)
			break
		}
	}

	var totals extrasSyncTotals
	var jsonEntries []syncExtrasJSONEntry

	if !jsonOutput {
		ui.Header(ui.WithModeLabel("Syncing extras"))
	}

	opts := sync.ExtraRunOptions{
		DryRun:        dryRun,
		Force:         force,
		MissingSource: sync.MissingSourceCreate,
		ResolvePath:   config.ExpandPath,
		ResolveExtension: func(ext string) (*sync.ExtensionSpec, error) {
			return resolveExtension(ext, globalExtensionsDir())
		},
		AgentTargetPaths: agentTargetPaths,
	}

	for _, extra := range cfg.Extras {
		extraSource := config.ResolveExtrasSourceDir(extra, cfg.EffectiveExtrasSource(), cfg.EffectiveSkillsSource())

		run := sync.RunExtraTargets(extra, extraSource, opts)
		if run.SourceErr != nil {
			if !jsonOutput {
				ui.Warning("Failed to create source directory: %s", shortenPath(extraSource))
			}
			if jsonOutput {
				jsonEntries = append(jsonEntries, syncExtrasJSONEntry{Name: extra.Name, Targets: []syncExtrasJSONTarget{}})
			}
			continue
		}
		if run.SourceCreated && !jsonOutput {
			ui.Info("Created source directory: %s", shortenPath(extraSource))
		}

		jsonEntry := syncExtrasJSONEntry{Name: extra.Name}
		for _, tr := range run.Targets {
			reportExtraTarget(extra.Name, tr, jsonOutput, &totals)
			jsonEntry.Targets = append(jsonEntry.Targets, extraTargetJSON(tr, tr.Target.Path))
		}
		jsonEntries = append(jsonEntries, jsonEntry)
	}

	// Oplog
	status := "ok"
	if totals.errors > 0 {
		status = "partial"
	}
	e := oplog.NewEntry("sync-extras", status, time.Since(start))
	e.Args = map[string]any{
		"extras_count": len(cfg.Extras),
		"synced":       totals.synced,
		"skipped":      totals.skipped,
		"pruned":       totals.pruned,
		"errors":       totals.errors,
		"dry_run":      dryRun,
		"force":        force,
	}
	oplog.WriteWithLimit(config.ConfigPath(), oplog.OpsFile, e, logMaxEntries()) //nolint:errcheck

	if jsonOutput {
		output := syncExtrasJSONOutput{
			Extras:   jsonEntries,
			Duration: formatDuration(start),
		}
		if err := writeJSON(&output); err != nil {
			return err
		}
		if totals.errors > 0 {
			return &jsonSilentError{cause: fmt.Errorf("%d extras sync error(s)", totals.errors)}
		}
		return nil
	}

	ui.ExtrasSyncSummary(ui.ExtrasSyncStats{
		Targets:  totals.targets,
		Synced:   totals.synced,
		Skipped:  totals.skipped,
		Pruned:   totals.pruned,
		Duration: time.Since(start),
	})

	if totals.errors > 0 {
		return fmt.Errorf("%d extras sync error(s)", totals.errors)
	}
	return nil
}

func cmdSyncExtrasProject(cwd string, dryRun, force, jsonOutput bool, start time.Time) error {
	projCfg, err := config.LoadProject(cwd)
	if err != nil {
		return err
	}

	if _, err := config.ValidateProjectConfig(projCfg, cwd); err != nil {
		return err
	}
	if len(projCfg.Extras) == 0 {
		// Clean up empty extras directory
		removeEmptyDir(config.ExtrasParentDirProject(projCfg.EffectiveExtrasSource(cwd)))

		if jsonOutput {
			return writeJSON(&syncExtrasJSONOutput{Extras: []syncExtrasJSONEntry{}, Duration: formatDuration(start)})
		}
		ui.Info("No extras configured in project.")
		ui.Info("Run 'skillshare extras init <name> --target <path> -p' to add one.")
		return nil
	}

	if dryRun && !jsonOutput {
		ui.Warning("Dry run mode - no changes will be made")
	}

	// Detect overlap between extras "agents" and the agents sync system
	var agentTargetPaths map[string]bool
	for _, extra := range projCfg.Extras {
		if extra.Name == extrasAgentsName {
			agentTargetPaths = collectAgentTargetPathsProject(cwd)
			break
		}
	}

	var totals extrasSyncTotals
	var jsonEntries []syncExtrasJSONEntry

	if !jsonOutput {
		ui.Header(ui.WithModeLabel("Syncing extras"))
	}

	opts := sync.ExtraRunOptions{
		DryRun:        dryRun,
		Force:         force,
		ProjectRoot:   cwd,
		MissingSource: sync.MissingSourceSkip,
		// Expand ~ and resolve relative paths against project root
		ResolvePath: func(path string) string { return resolveProjectPath(cwd, path) },
		ResolveExtension: func(ext string) (*sync.ExtensionSpec, error) {
			return resolveExtension(ext, projectExtensionsDir(cwd))
		},
		AgentTargetPaths: agentTargetPaths,
	}

	for _, extra := range projCfg.Extras {
		extraSource := config.ResolveExtrasSourceDirProject(extra, projCfg.EffectiveExtrasSource(cwd), cwd)

		run := sync.RunExtraTargets(extra, extraSource, opts)
		if run.SourceMissing {
			if !jsonOutput {
				ui.Info("Source directory does not exist: %s", extraSource)
				ui.Info("Create it to start syncing %s", extra.Name)
			}
			if jsonOutput {
				jsonEntries = append(jsonEntries, syncExtrasJSONEntry{Name: extra.Name, Targets: []syncExtrasJSONTarget{}})
			}
			continue
		}

		jsonEntry := syncExtrasJSONEntry{Name: extra.Name}
		for _, tr := range run.Targets {
			reportExtraTarget(extra.Name, tr, jsonOutput, &totals)
			// Targets that reached the sync report the resolved path.
			path := tr.Target.Path
			if tr.SkippedBy == "" && tr.ModeErr == nil && tr.ExtensionErr == nil {
				path = tr.Path
			}
			jsonEntry.Targets = append(jsonEntry.Targets, extraTargetJSON(tr, path))
		}
		jsonEntries = append(jsonEntries, jsonEntry)
	}

	status := "ok"
	if totals.errors > 0 {
		status = "partial"
	}
	e := oplog.NewEntry("sync-extras", status, time.Since(start))
	e.Args = map[string]any{
		"extras_count": len(projCfg.Extras),
		"synced":       totals.synced,
		"skipped":      totals.skipped,
		"pruned":       totals.pruned,
		"errors":       totals.errors,
		"dry_run":      dryRun,
		"force":        force,
		"scope":        "project",
	}
	oplog.WriteWithLimit(config.ProjectConfigPath(cwd), oplog.OpsFile, e, logMaxEntries()) //nolint:errcheck

	if jsonOutput {
		output := syncExtrasJSONOutput{
			Extras:   jsonEntries,
			Duration: formatDuration(start),
		}
		if err := writeJSON(&output); err != nil {
			return err
		}
		if totals.errors > 0 {
			return &jsonSilentError{cause: fmt.Errorf("%d extras sync error(s)", totals.errors)}
		}
		return nil
	}

	ui.ExtrasSyncSummary(ui.ExtrasSyncStats{
		Targets:  totals.targets,
		Synced:   totals.synced,
		Skipped:  totals.skipped,
		Pruned:   totals.pruned,
		Duration: time.Since(start),
	})

	if totals.errors > 0 {
		return fmt.Errorf("%d extras sync error(s)", totals.errors)
	}
	return nil
}

// syncVerb returns a user-facing verb for the given sync mode.
func syncVerb(mode string) string {
	switch mode {
	case "copy":
		return "copied"
	case "symlink":
		return "linked"
	case "import":
		return "imported"
	default:
		return "synced"
	}
}

// runExtrasSync runs extras sync and returns JSON entries without printing.
// Used by sync --all --json to merge extras into the skills JSON output.
// agentTargetPaths is used to skip extras "agents" targets that overlap with the agents sync system.
func runExtrasSyncEntries(extras []config.ExtraConfig, sourceFunc func(config.ExtraConfig) string, dryRun, force bool, projectRoot string, agentTargetPaths map[string]bool) []syncExtrasJSONEntry {
	// Resolve the per-target transform extension so --all --json applies
	// it like a normal sync instead of copying files verbatim.
	extDir := globalExtensionsDir()
	resolvePath := config.ExpandPath
	if projectRoot != "" {
		extDir = projectExtensionsDir(projectRoot)
		resolvePath = func(path string) string { return resolveProjectPath(projectRoot, path) }
	}
	opts := sync.ExtraRunOptions{
		DryRun:        dryRun,
		Force:         force,
		ProjectRoot:   projectRoot,
		MissingSource: sync.MissingSourceSkip,
		ResolvePath:   resolvePath,
		ResolveExtension: func(ext string) (*sync.ExtensionSpec, error) {
			return resolveExtension(ext, extDir)
		},
		AgentTargetPaths: agentTargetPaths,
	}

	entries := make([]syncExtrasJSONEntry, 0, len(extras))
	for _, extra := range extras {
		entry := syncExtrasJSONEntry{Name: extra.Name}
		run := sync.RunExtraTargets(extra, sourceFunc(extra), opts)
		if run.SourceMissing {
			entry.Targets = []syncExtrasJSONTarget{}
		}
		for _, tr := range run.Targets {
			entry.Targets = append(entry.Targets, extraTargetJSON(tr, tr.Path))
		}
		entries = append(entries, entry)
	}
	return entries
}

// extrasSyncTotals accumulates target outcomes for the summary and oplog.
type extrasSyncTotals struct {
	synced, skipped, pruned, errors, targets int
}

// reportExtraTarget adds one target's outcome to totals and, unless
// jsonOutput, prints it.
func reportExtraTarget(extraName string, tr sync.ExtraTargetRun, jsonOutput bool, totals *extrasSyncTotals) {
	totals.targets++
	shortTarget := shortenPath(tr.Path)

	if tr.SkippedBy != "" {
		if !jsonOutput {
			ui.Warning("Skipping extras %q target %s — already managed by agents sync", extraName, shortTarget)
		}
		return
	}
	if err := extraTargetFailure(tr); err != nil {
		if !jsonOutput {
			ui.Warning("%s: %v", shortTarget, err)
		}
		totals.errors++
		return
	}

	result := tr.Result
	totals.synced += result.Synced
	totals.skipped += result.Skipped
	totals.pruned += result.Pruned
	totals.errors += len(result.Errors)
	if jsonOutput {
		return
	}

	shownMode := tr.Mode
	verb := syncVerb(shownMode)
	if result.Synced > 0 {
		parts := []string{fmt.Sprintf("%d files %s", result.Synced, verb)}
		if result.Pruned > 0 {
			parts = append(parts, fmt.Sprintf("%d pruned", result.Pruned))
		}
		ui.Success("%s  %s (%s)", shortTarget, strings.Join(parts, ", "), shownMode)
	} else if result.Skipped > result.Preserved {
		ui.Warning("%s  %d files skipped (use --force to override)", shortTarget, result.Skipped-result.Preserved)
	} else if result.Preserved == 0 {
		ui.Success("%s  up to date (%s)", shortTarget, shownMode)
	}
	if result.Preserved > 0 {
		ui.Success("%s  %d local preserved", shortTarget, result.Preserved)
	}

	for _, e := range result.Errors {
		ui.Warning("    %s", e)
	}
	for _, w := range result.Warnings {
		ui.Info("    %s", w)
	}
}

// extraTargetFailure returns the error that stopped a target, checking the
// mode before the extension.
func extraTargetFailure(tr sync.ExtraTargetRun) error {
	switch {
	case tr.ModeErr != nil:
		return tr.ModeErr
	case tr.ExtensionErr != nil:
		return tr.ExtensionErr
	default:
		return tr.Err
	}
}

// extraTargetJSON returns the JSON form of one target, reported at path.
func extraTargetJSON(tr sync.ExtraTargetRun, path string) syncExtrasJSONTarget {
	jt := syncExtrasJSONTarget{Path: path, Mode: tr.Mode, SkippedBy: tr.SkippedBy}
	if err := extraTargetFailure(tr); err != nil {
		jt.Error = err.Error()
		return jt
	}
	if tr.Result != nil {
		jt.Synced = tr.Result.Synced
		jt.Skipped = tr.Result.Skipped
		jt.Pruned = tr.Result.Pruned
		jt.Warnings = tr.Result.Warnings
		if len(tr.Result.Errors) > 0 {
			jt.Error = strings.Join(tr.Result.Errors, "; ")
		}
	}
	return jt
}

// cachedHome caches the home directory for shortenPath.
var cachedHome = func() string {
	h, _ := os.UserHomeDir()
	return h
}()

// shortenPath replaces the home directory prefix with ~.
func shortenPath(p string) string {
	if cachedHome != "" && strings.HasPrefix(p, cachedHome) {
		return "~" + p[len(cachedHome):]
	}
	return p
}
