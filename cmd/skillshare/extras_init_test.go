package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestValidateExtrasInit(t *testing.T) {
	target := func(mode string, flatten bool, as string) []extrasInitTarget {
		return []extrasInitTarget{{path: "/tmp/t", mode: mode, flatten: flatten, as: as}}
	}
	cases := []struct {
		name string
		opts extrasInitOptions
		want string // "" = valid
	}{
		{"single file with as", extrasInitOptions{name: "p", file: "system.md", targets: target("", false, "APPEND.md")}, ""},
		{"single file import", extrasInitOptions{name: "p", file: "system.md", targets: target("import", false, "")}, ""},
		{"as without file", extrasInitOptions{name: "p", targets: target("", false, "APPEND.md")}, "--as requires --file"},
		{"import without file", extrasInitOptions{name: "p", targets: target("import", false, "")}, "import mode requires a single-file extra"},
		{"flatten with file", extrasInitOptions{name: "p", file: "system.md", targets: target("", true, "")}, "flatten cannot be used with a single-file extra"},
		{"file is a path", extrasInitOptions{name: "p", file: "a/system.md", targets: target("", false, "")}, "must be a plain filename"},
		{"as is a path", extrasInitOptions{name: "p", file: "system.md", targets: target("", false, "../x.md")}, "must be a plain filename"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateExtrasInit(tc.opts)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestExtrasInitTUI_SingleFileFlow(t *testing.T) {
	var m tea.Model = newExtrasInitTUIModel()
	typeText := func(s string) {
		for _, r := range s {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}
	}
	key := func(k string) {
		switch k {
		case "enter":
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		case "down":
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		}
	}

	typeText("pi-prompt")
	key("enter")
	key("down") // Single file
	key("enter")
	typeText("system.md")
	key("enter")
	key("enter") // default source
	typeText("/tmp/pi")
	key("enter")
	key("enter") // keep the source file name
	key("enter") // merge

	got := m.(extrasInitTUIModel)
	if got.phase != extrasPhaseAddMore {
		t.Fatalf("phase = %v, want add-more (no flatten step for a single file)", got.phase)
	}
	if !got.singleFile || got.file != "system.md" || len(got.targets) != 1 || got.targets[0].as != "" || got.targets[0].mode != "merge" {
		t.Errorf("model = singleFile %v file %q targets %+v", got.singleFile, got.file, got.targets)
	}
}
