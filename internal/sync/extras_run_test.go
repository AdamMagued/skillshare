package sync

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"skillshare/internal/config"
)

func runTestOptions(policy MissingSourcePolicy) ExtraRunOptions {
	return ExtraRunOptions{
		MissingSource: policy,
		ResolvePath:   func(p string) string { return p },
		ResolveExtension: func(ext string) (*ExtensionSpec, error) {
			return nil, errors.New("no extension " + ext)
		},
	}
}

func TestRunExtraTargets_MissingSourceSkip(t *testing.T) {
	src := filepath.Join(t.TempDir(), "missing")
	extra := config.ExtraConfig{Name: "rules", Targets: []config.ExtraTargetConfig{{Path: t.TempDir()}}}

	run := RunExtraTargets(extra, src, runTestOptions(MissingSourceSkip))

	if !run.SourceMissing || len(run.Targets) != 0 {
		t.Fatalf("expected skipped extra, got %+v", run)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source must not be created, stat err = %v", err)
	}
}

func TestRunExtraTargets_MissingSourceCreate(t *testing.T) {
	src := filepath.Join(t.TempDir(), "missing")
	extra := config.ExtraConfig{Name: "rules", Targets: []config.ExtraTargetConfig{{Path: t.TempDir()}}}

	run := RunExtraTargets(extra, src, runTestOptions(MissingSourceCreate))

	if !run.SourceCreated || len(run.Targets) != 1 || run.Targets[0].Err != nil {
		t.Fatalf("expected created source and one synced target, got %+v", run)
	}
}

func TestRunExtraTargets_SkipsAgentOverlap(t *testing.T) {
	src, tgt := setupExtrasTest(t, map[string]string{"a.md": "a"})
	extra := config.ExtraConfig{Name: "agents", Targets: []config.ExtraTargetConfig{{Path: tgt}}}
	opts := runTestOptions(MissingSourceSkip)
	opts.AgentTargetPaths = map[string]bool{filepath.Clean(tgt): true}

	run := RunExtraTargets(extra, src, opts)

	if got := run.Targets[0]; got.SkippedBy != "agents" || got.Result != nil {
		t.Fatalf("expected target skipped by agents, got %+v", got)
	}
}

func TestRunExtraTargets_ReportsModeAndExtensionErrors(t *testing.T) {
	src, tgt := setupExtrasTest(t, map[string]string{"a.md": "a"})
	extra := config.ExtraConfig{Name: "rules", Targets: []config.ExtraTargetConfig{{Path: tgt, Mode: "merge", Extension: "x"}}}

	got := RunExtraTargets(extra, src, runTestOptions(MissingSourceSkip)).Targets[0]

	if got.ModeErr == nil || got.ExtensionErr == nil || got.Result != nil || got.Mode != "merge" {
		t.Fatalf("expected both errors, no sync and configured mode, got %+v", got)
	}
}
