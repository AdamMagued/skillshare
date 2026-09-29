package main

import (
	"fmt"
	"strings"
	"time"

	"skillshare/internal/config"
	"skillshare/internal/theme"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type extrasInitPhase int

const (
	extrasPhaseNameInput     extrasInitPhase = iota
	extrasPhaseKindSelect                    // folder or single file?
	extrasPhaseFileInput                     // single file: source file name
	extrasPhaseSourceInput                   // ask for custom source directory
	extrasPhaseTargetInput                   // ask for target path
	extrasPhaseAsInput                       // single file: target file name
	extrasPhaseModeSelect                    // choose sync mode
	extrasPhaseFlattenToggle                 // flatten files into target root?
	extrasPhaseAddMore                       // add another target?
	extrasPhaseConfirm                       // show summary and confirm
)

type extrasInitTarget struct {
	path    string
	mode    string
	flatten bool
	as      string // single file: target filename ("" = the source file name)
}

type extrasInitTUIModel struct {
	phase       extrasInitPhase
	name        string
	sourceValue string
	singleFile  bool
	kindCursor  int    // 0 folder, 1 single file
	file        string // single file: source file name
	targets     []extrasInitTarget
	currMode    int // cursor index into m.modes()

	textInput   textinput.Model
	sourceInput textinput.Model
	done        bool
	cancelled   bool
	err         error
}

var syncModes = config.ValidSyncModes // folder extras; import needs a single file

// singleFileSyncModes are offered for a single-file extra. symlink is left out:
// for one file it does the same as merge.
var singleFileSyncModes = []string{"merge", "copy", "import"}

// modes returns the sync modes offered for the extra being created.
func (m extrasInitTUIModel) modes() []string {
	if m.singleFile {
		return singleFileSyncModes
	}
	return syncModes
}

// targetLabel returns a target as listed in the wizard: its path (for a
// single file, the file it writes) and settings.
func (m extrasInitTUIModel) targetLabel(t extrasInitTarget) string {
	path, modeLabel := t.path, t.mode
	if m.singleFile {
		path = singleFileTargetPath(config.ExtraConfig{File: m.file}, config.ExtraTargetConfig{Path: t.path, As: t.as})
	}
	if t.flatten {
		modeLabel += ", flatten"
	}
	return fmt.Sprintf("%s (%s)", path, modeLabel)
}

func newExtrasInitTUIModel() extrasInitTUIModel {
	ti := textinput.New()
	ti.Placeholder = "rules"
	ti.Focus()
	ti.PromptStyle = theme.Accent()
	ti.Cursor.Style = theme.Accent()

	si := textinput.New()
	si.Placeholder = "Leave empty to use default"
	si.CharLimit = 256
	si.PromptStyle = theme.Accent()
	si.Cursor.Style = theme.Accent()

	return extrasInitTUIModel{
		phase:       extrasPhaseNameInput,
		textInput:   ti,
		sourceInput: si,
	}
}

func (m extrasInitTUIModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m extrasInitTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "esc":
			// esc on first phase cancels; on later phases go back one step
			switch m.phase {
			case extrasPhaseNameInput:
				m.cancelled = true
				return m, tea.Quit
			case extrasPhaseKindSelect:
				m.phase = extrasPhaseNameInput
				m.textInput.SetValue(m.name)
				m.textInput.Placeholder = "rules"
				return m, nil
			case extrasPhaseFileInput:
				m.err = nil
				m.phase = extrasPhaseKindSelect
				return m, nil
			case extrasPhaseSourceInput:
				if m.singleFile {
					m.phase = extrasPhaseFileInput
					m.textInput.SetValue(m.file)
					m.textInput.Placeholder = "system.md"
					return m, nil
				}
				m.phase = extrasPhaseKindSelect
				return m, nil
			case extrasPhaseTargetInput:
				if len(m.targets) == 0 {
					m.phase = extrasPhaseSourceInput
					m.sourceInput.SetValue(m.sourceValue)
					m.sourceInput.Focus()
					return m, nil
				}
				m.phase = extrasPhaseAddMore
				return m, nil
			case extrasPhaseAsInput:
				m.err = nil
				m.targets = m.targets[:len(m.targets)-1] // remove the pending target
				m.phase = extrasPhaseTargetInput
				m.textInput.SetValue("")
				m.textInput.Placeholder = targetPlaceholder(len(m.targets))
				return m, nil
			case extrasPhaseModeSelect:
				if m.singleFile {
					m.phase = extrasPhaseAsInput
					m.textInput.SetValue(m.targets[len(m.targets)-1].as)
					m.textInput.Placeholder = m.file
					return m, nil
				}
				m.targets = m.targets[:len(m.targets)-1] // remove the pending target
				m.phase = extrasPhaseTargetInput
				m.textInput.SetValue("")
				m.textInput.Placeholder = targetPlaceholder(len(m.targets))
				return m, nil
			case extrasPhaseFlattenToggle:
				m.phase = extrasPhaseModeSelect
				return m, nil
			case extrasPhaseAddMore:
				if m.singleFile || m.targets[len(m.targets)-1].mode == "symlink" {
					m.phase = extrasPhaseModeSelect
				} else {
					m.phase = extrasPhaseFlattenToggle
				}
				return m, nil
			case extrasPhaseConfirm:
				m.phase = extrasPhaseAddMore
				return m, nil
			}
		}

		switch m.phase {
		case extrasPhaseNameInput:
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(m.textInput.Value())
				if name == "" {
					return m, nil
				}
				if err := config.ValidateExtraName(name); err != nil {
					m.err = err
					return m, nil
				}
				m.name = name
				m.err = nil
				m.phase = extrasPhaseKindSelect
				return m, nil
			}

		case extrasPhaseKindSelect:
			switch msg.String() {
			case "up", "k":
				m.kindCursor = 0
			case "down", "j":
				m.kindCursor = 1
			case "enter", " ":
				m.singleFile = m.kindCursor == 1
				if m.singleFile {
					m.phase = extrasPhaseFileInput
					m.textInput.SetValue(m.file)
					m.textInput.Placeholder = "system.md"
					return m, nil
				}
				m.phase = extrasPhaseSourceInput
				m.sourceInput.Focus()
			}
			return m, nil

		case extrasPhaseFileInput:
			switch msg.String() {
			case "enter":
				file := strings.TrimSpace(m.textInput.Value())
				if file == "" {
					return m, nil
				}
				if err := config.ValidateExtraConfig(config.ExtraConfig{Name: m.name, File: file}); err != nil {
					m.err = err
					return m, nil
				}
				m.file = file
				m.err = nil
				m.phase = extrasPhaseSourceInput
				m.sourceInput.Focus()
				return m, nil
			}

		case extrasPhaseSourceInput:
			switch msg.String() {
			case "enter":
				m.sourceValue = strings.TrimSpace(m.sourceInput.Value())
				m.phase = extrasPhaseTargetInput
				m.textInput.SetValue("")
				m.textInput.Placeholder = targetPlaceholder(0)
				return m, nil
			}

		case extrasPhaseTargetInput:
			switch msg.String() {
			case "enter":
				path := strings.TrimSpace(m.textInput.Value())
				if path == "" {
					return m, nil
				}
				m.targets = append(m.targets, extrasInitTarget{path: path})
				m.currMode = 0
				if m.singleFile {
					m.phase = extrasPhaseAsInput
					m.textInput.SetValue("")
					m.textInput.Placeholder = m.file
					return m, nil
				}
				m.phase = extrasPhaseModeSelect
				return m, nil
			}

		case extrasPhaseAsInput:
			switch msg.String() {
			case "enter":
				as := strings.TrimSpace(m.textInput.Value())
				if as == m.file {
					as = ""
				}
				tc := config.ExtraTargetConfig{Path: m.targets[len(m.targets)-1].path, As: as}
				if err := config.ValidateExtraConfig(config.ExtraConfig{Name: m.name, File: m.file, Targets: []config.ExtraTargetConfig{tc}}); err != nil {
					m.err = err
					return m, nil
				}
				m.err = nil
				m.targets[len(m.targets)-1].as = as
				m.currMode = 0
				m.phase = extrasPhaseModeSelect
				return m, nil
			}

		case extrasPhaseModeSelect:
			switch msg.String() {
			case "up", "k":
				if m.currMode > 0 {
					m.currMode--
				}
				return m, nil
			case "down", "j":
				if m.currMode < len(m.modes())-1 {
					m.currMode++
				}
				return m, nil
			case "enter", " ":
				mode := m.modes()[m.currMode]
				m.targets[len(m.targets)-1].mode = mode
				if m.singleFile || mode == "symlink" {
					m.phase = extrasPhaseAddMore // no flatten for one file or a directory symlink
				} else {
					m.phase = extrasPhaseFlattenToggle
				}
				return m, nil
			}
			return m, nil

		case extrasPhaseFlattenToggle:
			switch msg.String() {
			case "y", "Y":
				m.targets[len(m.targets)-1].flatten = true
				m.phase = extrasPhaseAddMore
				return m, nil
			case "n", "N", "enter":
				m.phase = extrasPhaseAddMore
				return m, nil
			}
			return m, nil

		case extrasPhaseAddMore:
			switch msg.String() {
			case "y", "Y":
				m.phase = extrasPhaseTargetInput
				m.textInput.SetValue("")
				m.textInput.Placeholder = targetPlaceholder(len(m.targets))
				return m, nil
			case "n", "N", "enter":
				m.phase = extrasPhaseConfirm
				return m, nil
			}
			return m, nil

		case extrasPhaseConfirm:
			switch msg.String() {
			case "y", "Y", "enter":
				m.done = true
				return m, tea.Quit
			case "n", "N":
				m.cancelled = true
				return m, tea.Quit
			}
			return m, nil
		}
	}

	// Delegate to textinput for typing phases
	switch m.phase {
	case extrasPhaseNameInput, extrasPhaseFileInput, extrasPhaseTargetInput, extrasPhaseAsInput:
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	case extrasPhaseSourceInput:
		var cmd tea.Cmd
		m.sourceInput, cmd = m.sourceInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m extrasInitTUIModel) View() string {
	var b strings.Builder

	b.WriteString(theme.Title().Render("Extras Init"))
	b.WriteString("\n\n")

	switch m.phase {
	case extrasPhaseNameInput:
		b.WriteString(theme.Accent().Render("Extra name: "))
		b.WriteString(m.textInput.View())
		if m.err != nil {
			b.WriteString("\n" + theme.Danger().Render(m.err.Error()))
		}
		b.WriteString("\n\n")
		b.WriteString(theme.Dim().MarginLeft(2).Render("enter confirm  esc cancel"))

	case extrasPhaseKindSelect:
		b.WriteString(theme.Dim().Render(fmt.Sprintf("Name: %s", m.name)))
		b.WriteString("\n\n")
		b.WriteString(theme.Accent().Render("What do you want to sync?"))
		b.WriteString("\n")
		kinds := [][2]string{{"Folder", " (every file in a folder)"}, {"Single file", " (one file, renamed per target if needed)"}}
		for i, k := range kinds {
			if i == m.kindCursor {
				b.WriteString(theme.Accent().Render("▸ "+k[0]) + theme.Dim().Render(k[1]))
			} else {
				b.WriteString(theme.Dim().Render("  " + k[0] + k[1]))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(theme.Dim().MarginLeft(2).Render("↑↓/jk navigate  enter/space select  esc back"))

	case extrasPhaseFileInput:
		b.WriteString(theme.Dim().Render(fmt.Sprintf("Name: %s", m.name)))
		b.WriteString("\n\n")
		b.WriteString(theme.Accent().Render("Source file name: "))
		b.WriteString(m.textInput.View())
		if m.err != nil {
			b.WriteString("\n" + theme.Danger().Render(m.err.Error()))
		}
		b.WriteString("\n\n")
		b.WriteString(theme.Dim().MarginLeft(2).Render("enter confirm  esc back"))

	case extrasPhaseSourceInput:
		b.WriteString(theme.Dim().Render(fmt.Sprintf("Name: %s", m.name)))
		if m.singleFile {
			b.WriteString("\n")
			b.WriteString(theme.Dim().Render(fmt.Sprintf("File: %s", m.file)))
		}
		b.WriteString("\n\n")
		b.WriteString(theme.Accent().Render("Source directory (optional): "))
		b.WriteString(m.sourceInput.View())
		b.WriteString("\n\n")
		b.WriteString(theme.Dim().MarginLeft(2).Render("enter to skip (use default)  esc back"))

	case extrasPhaseTargetInput:
		b.WriteString(theme.Dim().Render(fmt.Sprintf("Name: %s", m.name)))
		if m.singleFile {
			b.WriteString("\n")
			b.WriteString(theme.Dim().Render(fmt.Sprintf("File: %s", m.file)))
		}
		if m.sourceValue != "" {
			b.WriteString("\n")
			b.WriteString(theme.Dim().Render(fmt.Sprintf("Source: %s", m.sourceValue)))
		}
		if len(m.targets) > 0 {
			b.WriteString("\n")
			for _, t := range m.targets {
				b.WriteString(theme.Dim().Render("  → " + m.targetLabel(t)))
				b.WriteString("\n")
			}
		}
		b.WriteString("\n")
		b.WriteString(theme.Accent().Render(fmt.Sprintf("Target #%d path: ", len(m.targets)+1)))
		b.WriteString(m.textInput.View())
		b.WriteString("\n\n")
		b.WriteString(theme.Dim().MarginLeft(2).Render("enter confirm  esc back"))

	case extrasPhaseAsInput:
		b.WriteString(theme.Dim().Render(fmt.Sprintf("Name: %s", m.name)))
		b.WriteString("\n")
		b.WriteString(theme.Dim().Render(fmt.Sprintf("Target: %s", m.targets[len(m.targets)-1].path)))
		b.WriteString("\n\n")
		b.WriteString(theme.Accent().Render("Target file name: "))
		b.WriteString(m.textInput.View())
		if m.err != nil {
			b.WriteString("\n" + theme.Danger().Render(m.err.Error()))
		}
		b.WriteString("\n\n")
		b.WriteString(theme.Dim().MarginLeft(2).Render(fmt.Sprintf("enter to keep %s  esc back", m.file)))

	case extrasPhaseModeSelect:
		b.WriteString(theme.Dim().Render(fmt.Sprintf("Name: %s", m.name)))
		b.WriteString("\n")
		b.WriteString(theme.Dim().Render(fmt.Sprintf("Target: %s", m.targetLabel(m.targets[len(m.targets)-1]))))
		b.WriteString("\n\n")
		b.WriteString(theme.Accent().Render("Sync mode:"))
		b.WriteString("\n")
		for i, mode := range m.modes() {
			cursor := "  "
			if i == m.currMode {
				cursor = "▸ "
			}
			var desc string
			switch {
			case m.singleFile && mode == "merge":
				desc = " (file symlink, default)"
			case m.singleFile && mode == "copy":
				desc = " (file copy)"
			case mode == "import":
				desc = " (@path line in the target file)"
			case mode == "merge":
				desc = " (per-file symlinks, default)"
			case mode == "copy":
				desc = " (file copies)"
			case mode == "symlink":
				desc = " (directory symlink)"
			}
			if i == m.currMode {
				b.WriteString(theme.Accent().Render(cursor+mode) + theme.Dim().Render(desc))
			} else {
				b.WriteString(theme.Dim().Render(cursor + mode + desc))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(theme.Dim().MarginLeft(2).Render("↑↓/jk navigate  enter/space select  esc back"))

	case extrasPhaseFlattenToggle:
		b.WriteString(theme.Dim().Render(fmt.Sprintf("Name: %s", m.name)))
		b.WriteString("\n")
		lastTarget := m.targets[len(m.targets)-1]
		b.WriteString(theme.Dim().Render(fmt.Sprintf("Target: %s (%s)", lastTarget.path, lastTarget.mode)))
		b.WriteString("\n\n")
		b.WriteString(theme.Accent().Render("Flatten files into target root? (y/N) "))
		b.WriteString("\n\n")
		b.WriteString(theme.Dim().MarginLeft(2).Render("y yes  n/enter no  esc back"))

	case extrasPhaseAddMore:
		b.WriteString(theme.Dim().Render(fmt.Sprintf("Name: %s", m.name)))
		b.WriteString("\n")
		for _, t := range m.targets {
			b.WriteString(theme.Dim().Render("  → " + m.targetLabel(t)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(theme.Accent().Render("Add another target? (y/N) "))
		b.WriteString("\n\n")
		b.WriteString(theme.Dim().MarginLeft(2).Render("y yes  n/enter no  esc back"))

	case extrasPhaseConfirm:
		b.WriteString(theme.Accent().Render("Summary:"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Name: %s\n", m.name))
		if m.singleFile {
			b.WriteString(fmt.Sprintf("  File: %s\n", m.file))
		}
		if m.sourceValue != "" {
			b.WriteString(fmt.Sprintf("  Source: %s\n", m.sourceValue))
		}
		for _, t := range m.targets {
			b.WriteString(fmt.Sprintf("  → %s\n", m.targetLabel(t)))
		}
		b.WriteString("\n")
		b.WriteString(theme.Accent().Render("Create this extra? (Y/n) "))
		b.WriteString("\n\n")
		b.WriteString(theme.Dim().MarginLeft(2).Render("y/enter confirm  n cancel"))
	}

	return b.String()
}

// targetPlaceholder returns a contextual placeholder for the target input.
func targetPlaceholder(n int) string {
	placeholders := []string{
		"~/.claude/rules",
		"~/.cursor/rules",
		"~/.codex/rules",
	}
	if n < len(placeholders) {
		return placeholders[n]
	}
	return "~/.<tool>/rules"
}

// cmdExtrasInitTUI launches the interactive wizard when no arguments are provided.
func cmdExtrasInitTUI(mode runMode, cwd string) error {
	m := newExtrasInitTUIModel()
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	result, ok := finalModel.(extrasInitTUIModel)
	if !ok || result.cancelled || !result.done {
		return nil
	}

	start := time.Now()
	opts := extrasInitOptions{name: result.name, source: result.sourceValue, file: result.file}
	if result.singleFile {
		// Each target keeps its own file name and mode; merge is the default.
		for _, t := range result.targets {
			if t.mode == "merge" {
				t.mode = ""
			}
			opts.targets = append(opts.targets, t)
		}
	} else {
		// Collect targets and the first non-empty mode (mode applies globally)
		syncMode := ""
		flatten := false
		for _, t := range result.targets {
			if syncMode == "" && t.mode != "" && t.mode != "merge" {
				syncMode = t.mode
			}
			if t.flatten {
				flatten = true
			}
		}
		for _, t := range result.targets {
			opts.targets = append(opts.targets, extrasInitTarget{path: t.path, mode: syncMode, flatten: flatten})
		}
	}
	if err := validateExtrasInit(opts); err != nil {
		return err
	}

	if mode == modeProject {
		if err := config.ValidateProjectExtraSource(opts.source); err != nil {
			return err
		}
		return extrasInitProject(cwd, opts, start)
	}
	return extrasInitGlobal(opts, start)
}
