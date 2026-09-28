package main

import (
	"fmt"
	"strings"
	"time"

	"skillshare/internal/config"
	"skillshare/internal/oplog"
	"skillshare/internal/sync"
	"skillshare/internal/ui"
)

// stripNoSkillsFlag removes --no-skills from target add arguments.
func stripNoSkillsFlag(args []string) ([]string, bool) {
	rest := make([]string, 0, len(args))
	noSkills := false
	for _, arg := range args {
		if arg == "--no-skills" {
			noSkills = true
			continue
		}
		rest = append(rest, arg)
	}
	return rest, noSkills
}

// setTargetSkillsGlobal switches skills sync for a global target. Turning it
// off saves the config, then removes the folder's links into the source.
func setTargetSkillsGlobal(cfg *config.Config, name string, target config.TargetConfig, enabled, dryRun bool) error {
	start := time.Now()
	res, err := switchTargetSkills(name, enabled, dryRun, func() error {
		target.EnsureSkills().SetEnabled(enabled)
		cfg.Targets[name] = target
		return cfg.Save()
	}, func() (map[string]config.TargetConfig, error) {
		return cfg.Targets, nil
	}, cfg.EffectiveSkillsSource())
	logTargetSkillsOp(config.ConfigPath(), name, enabled, dryRun, res, start, err)
	return err
}

// setTargetSkillsProject is setTargetSkillsGlobal for a project target.
func setTargetSkillsProject(cfg *config.ProjectConfig, idx int, enabled, dryRun bool, root string) error {
	start := time.Now()
	name := cfg.Targets[idx].Name
	res, err := switchTargetSkills(name, enabled, dryRun, func() error {
		cfg.Targets[idx].EnsureSkills().SetEnabled(enabled)
		return cfg.Save(root)
	}, func() (map[string]config.TargetConfig, error) {
		return config.ResolveProjectTargets(root, cfg)
	}, cfg.EffectiveSkillsSource(root))
	logTargetSkillsOp(config.ProjectConfigPath(root), name, enabled, dryRun, res, start, err)
	return err
}

func switchTargetSkills(name string, enabled, dryRun bool, save func() error, targets func() (map[string]config.TargetConfig, error), sourcePath string) (*sync.SkillsOffResult, error) {
	if dryRun {
		ui.Warning("Dry run mode - no changes will be made")
	} else if err := save(); err != nil {
		return nil, err
	}
	if enabled {
		if dryRun {
			ui.Info("%s: would turn skills on", name)
		} else {
			ui.Success("%s: skills on", name)
		}
		ui.Info("Run 'skillshare sync' to sync skills to this target")
		return nil, nil
	}

	all, err := targets()
	if err != nil {
		return nil, err
	}
	res, err := sync.DetachSkills(all, name, sourcePath, dryRun)
	if err != nil {
		return nil, fmt.Errorf("%s: skills off saved, but removing links failed: %w", name, err)
	}
	renderSkillsOff(name, dryRun, res)
	return res, nil
}

func renderSkillsOff(name string, dryRun bool, res *sync.SkillsOffResult) {
	removeVerb, keepVerb := "removed", "kept"
	if dryRun {
		ui.Info("%s: would turn skills off", name)
		removeVerb, keepVerb = "would remove", "would keep"
	} else {
		ui.Success("%s: skills off", name)
	}
	if res.SharedWith != "" {
		ui.Info("  skills folder is shared with %s, left as is", res.SharedWith)
		return
	}
	if len(res.Removed) > 0 {
		ui.Info("  %s %d link(s): %s", removeVerb, len(res.Removed), strings.Join(res.Removed, ", "))
	}
	if len(res.Kept) > 0 {
		ui.Info("  %s %d: %s", keepVerb, len(res.Kept), strings.Join(res.Kept, ", "))
	}
	ui.Info("  Agents, MCP servers and instructions are still managed")
}

func logTargetSkillsOp(cfgPath, name string, enabled, dryRun bool, res *sync.SkillsOffResult, start time.Time, err error) {
	e := oplog.NewEntry("target", statusFromErr(err), time.Since(start))
	e.Args = map[string]any{
		"action":  "skills",
		"name":    name,
		"enabled": enabled,
		"dry_run": dryRun,
	}
	if res != nil {
		e.Args["removed"] = len(res.Removed)
		e.Args["kept"] = len(res.Kept)
	}
	if err != nil {
		e.Message = err.Error()
	}
	oplog.WriteWithLimit(cfgPath, oplog.OpsFile, e, logMaxEntries()) //nolint:errcheck
}
