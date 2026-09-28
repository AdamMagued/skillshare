package sync

import (
	"os"
	"slices"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/utils"
)

// TargetSkills returns the skills a target actually loads, for analyze and
// context-cost reports.
//
// A symlink-mode target exposes the whole source folder, so every discovered
// skill counts, .skillignore'd ones included (pass a discovery that keeps them,
// such as DiscoverSourceSkillsForAnalyze). Any other mode gets the enabled
// skills its include/exclude and `targets:` filters let through, plus the
// skills that already live in the target folder without coming from the
// source (Local), which the tool loads just the same.
func TargetSkills(name string, target config.TargetConfig, defaultMode, sourcePath string, discovered []DiscoveredSkill) ([]DiscoveredSkill, error) {
	sc := target.SkillsConfig()
	if !sc.IsEnabled() {
		// Nothing is synced; the tool loads only what already sits in its folder.
		return localTargetSkills(sc.Path, sourcePath), nil
	}
	mode := sc.Mode
	if mode == "" {
		mode = defaultMode
	}
	if mode == "symlink" {
		return discovered, nil
	}

	enabled := slices.DeleteFunc(slices.Clone(discovered), func(s DiscoveredSkill) bool { return s.Disabled })
	skills, err := FilterSkills(enabled, sc.Include, sc.Exclude)
	if err != nil {
		return nil, err
	}
	skills = FilterSkillsByTarget(skills, name)
	return append(skills, localTargetSkills(sc.Path, sourcePath)...), nil
}

// localTargetSkills finds skills that sit in targetPath as real folders but are
// not managed by skillshare: merge mode links source skills, so a real folder
// is local; copy mode records its copies in the manifest, so anything missing
// from it is local. Errors mean "nothing to report" — the target folder may
// not exist yet.
func localTargetSkills(targetPath, sourcePath string) []DiscoveredSkill {
	if targetPath == "" || sourcePath == "" {
		return nil
	}
	info, err := os.Stat(targetPath)
	if err != nil || !info.IsDir() || utils.ResolveSymlink(targetPath) == utils.ResolveSymlink(sourcePath) {
		return nil
	}
	manifest, err := ReadManifest(targetPath)
	if err != nil {
		return nil
	}
	// filepath.Walk does not follow symlinks, so merge-mode links never yield a SKILL.md.
	found, _, _, err := discoverSourceSkillsInternal(targetPath, discoverOptions{collectContext: true})
	if err != nil {
		return nil
	}
	var local []DiscoveredSkill
	for _, s := range found {
		top, _, _ := strings.Cut(s.RelPath, "/")
		if _, managed := manifest.Managed[s.RelPath]; managed {
			continue
		}
		if _, managed := manifest.Managed[top]; managed {
			continue
		}
		s.Local = true
		s.Targets = nil // a folder the tool already reads is loaded whatever its path implies
		local = append(local, s)
	}
	return local
}
