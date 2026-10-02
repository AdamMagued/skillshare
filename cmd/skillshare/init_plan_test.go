package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDetectCLIDirectories_SkillsFirstThenByName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, dir := range []string{".cursor", ".claude/skills/a", ".claude/skills/b", ".gemini/skills/c"} {
		os.MkdirAll(filepath.Join(home, dir), 0o755)
	}

	var names []string
	for _, d := range detectCLIDirectories() {
		names = append(names, d.name)
	}

	want := []string{"claude", "gemini", "cursor", "universal"}
	if !slices.Equal(names, want) {
		t.Errorf("order = %v, want %v", names, want)
	}
}

func TestDetectCLIDirectories_DoesNotCreateFolders(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	os.MkdirAll(filepath.Join(home, ".cursor"), 0o755)

	detectCLIDirectories()

	if _, err := os.Stat(filepath.Join(home, ".cursor", "skills")); err == nil {
		t.Error("detection created ~/.cursor/skills")
	}
}

func TestNewInitPlan_DefaultsMatchPressingEnter(t *testing.T) {
	detected := []detectedDir{{name: "claude", path: "/x", skillCount: 1, hasSkills: true}, {name: "universal"}}

	p := newInitPlan(&initOptions{sourcePath: "/src"}, detected, "/home")

	if !slices.Equal(p.targets, []string{"claude", "universal"}) || !p.git || !p.skill || p.mode != "merge" || p.gitScope != "skills" {
		t.Errorf("plan = %+v, want every tool, git on (skills), skill on, merge", p)
	}
}

func TestNewInitPlan_FlagsOverrideDefaults(t *testing.T) {
	detected := []detectedDir{{name: "claude"}, {name: "universal"}}

	p := newInitPlan(&initOptions{sourcePath: "/src", noTargets: true, noGit: true, noSkill: true, mode: "copy"}, detected, "/home")

	if len(p.targets) != 0 || p.git || p.skill || p.mode != "copy" {
		t.Errorf("plan = %+v, want flags to win", p)
	}
}

func TestDefaultGitScope_RemoteVersionsEverything(t *testing.T) {
	p := &initPlan{remoteURL: "git@example.com:me/skills.git"}

	if got := p.defaultGitScope(); got != "root" {
		t.Errorf("scope = %q, want root", got)
	}
}

func TestSameNameAsRemote_ListsImportsTheRepoAlsoHas(t *testing.T) {
	p := &initPlan{
		pull:    true,
		remote:  &remoteRepo{names: []string{"pdf", "tdd"}},
		imports: []importedSkill{{name: "pdf"}, {name: "mine"}},
	}

	if got := p.sameNameAsRemote(); !slices.Equal(got, []string{"pdf"}) {
		t.Errorf("same = %v, want [pdf]", got)
	}
}

func TestSameTree_ComparesContent(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	write := func(dir, name, body string) {
		os.MkdirAll(filepath.Join(dir, name), 0o755)
		os.WriteFile(filepath.Join(dir, name, "SKILL.md"), []byte(body), 0o644)
	}
	write(src, "same", "x")
	write(dst, "same", "x")
	write(src, "edited", "x")
	write(dst, "edited", "y")

	if !sameTree(filepath.Join(src, "same"), filepath.Join(dst, "same")) {
		t.Error("identical folders should match")
	}
	if sameTree(filepath.Join(src, "edited"), filepath.Join(dst, "edited")) {
		t.Error("edited folders should not match")
	}
}
