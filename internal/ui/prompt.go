package ui

import (
	"errors"
	"fmt"
	"io"
	"os"

	"skillshare/internal/theme"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ErrCancelled is returned when the user leaves a prompt with esc or ctrl+c.
var ErrCancelled = errors.New("cancelled")

// promptVisibleRows is how many options a list shows before it scrolls.
const promptVisibleRows = 9

// answerLabelWidth aligns the values of collapsed answers.
const answerLabelWidth = 8

// Option is one choice in Select or MultiSelect.
type Option struct {
	Label string
	Value string
}

// Select asks for one option and returns its value. The prompt runs inline
// and erases itself when answered; call Answered to leave a summary line.
func Select(title string, options []Option, selected string) (string, error) {
	value := selected
	field := huh.NewSelect[string]().
		Title(promptTitle(title)).
		Options(huhOptions(options, nil)...).
		Value(&value)
	if len(options) > promptVisibleRows {
		field = field.Height(promptVisibleRows + 1)
	}
	return value, runPrompt(field)
}

// MultiSelect asks for any number of options and returns the chosen values
// in option order. Values in selected start checked.
func MultiSelect(title string, options []Option, selected []string) ([]string, error) {
	checked := map[string]bool{}
	for _, v := range selected {
		checked[v] = true
	}
	var values []string
	field := huh.NewMultiSelect[string]().
		Title(promptTitle(title)).
		Options(huhOptions(options, checked)...).
		Filterable(false).
		Value(&values)
	if len(options) > promptVisibleRows {
		field = field.Height(promptVisibleRows + 1)
	}
	return values, runPrompt(field)
}

// Confirm asks a yes/no question; def is the answer Enter gives.
func Confirm(title string, def bool) (bool, error) {
	value := def
	field := huh.NewConfirm().
		Title(promptTitle(title)).
		Affirmative("Yes").
		Negative("No").
		Inline(true).
		Value(&value)
	return value, runPrompt(field)
}

// Input asks for one line of text. placeholder is shown while it is empty.
func Input(title, placeholder, value string) (string, error) {
	field := huh.NewInput().
		Title(promptTitle(title)).
		Prompt("› ").
		Placeholder(placeholder).
		Value(&value)
	return value, runPrompt(field)
}

// Answered prints the one-line record an answered prompt collapses into.
func Answered(label, value string) {
	fmt.Printf("%s %-*s %s\n", theme.Success().Render("✓"), answerLabelWidth, label, value)
}

func runPrompt(field huh.Field) error {
	return runForm(field, os.Stdin, os.Stdout)
}

// runForm runs the form itself instead of huh's Run, which cancels through
// tea.Interrupt: Bubble Tea treats that as a kill and skips the final empty
// frame, leaving the question on screen after esc.
func runForm(field huh.Field, in io.Reader, out io.Writer) error {
	keys := huh.NewDefaultKeyMap()
	keys.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	keys.Select.Filter.SetEnabled(false)
	form := huh.NewForm(huh.NewGroup(field)).
		WithTheme(promptTheme()).
		WithKeyMap(keys).
		WithShowHelp(true)
	form.SubmitCmd = tea.Quit
	form.CancelCmd = tea.Quit
	if _, err := tea.NewProgram(form, tea.WithInput(in), tea.WithOutput(out)).Run(); err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			return ErrCancelled
		}
		return err
	}
	if form.State == huh.StateAborted {
		return ErrCancelled
	}
	return nil
}

// focusedButton fills the chosen Yes/No button with the accent color.
func focusedButton(button, accent lipgloss.Style) lipgloss.Style {
	t := theme.Get()
	if t.NoColor {
		return button.Reverse(true)
	}
	text := lipgloss.Color("0")
	if t.Mode == theme.ModeLight {
		text = lipgloss.Color("15")
	}
	return button.Background(accent.GetForeground()).Foreground(text).Bold(true)
}

// promptTitle marks a question with an accent "?", matching the "›" cursor.
func promptTitle(title string) string {
	return theme.Accent().Render("?") + " " + theme.Primary().Render(title)
}

func huhOptions(options []Option, checked map[string]bool) []huh.Option[string] {
	out := make([]huh.Option[string], len(options))
	for i, o := range options {
		out[i] = huh.NewOption(o.Label, o.Value).Selected(checked[o.Value])
	}
	return out
}

// promptTheme maps the skillshare palette onto huh's styles: no side bar,
// "›" cursor, ◉/○ checkboxes, accent-colored focus.
func promptTheme() *huh.Theme {
	t := huh.ThemeBase()
	primary, accent := theme.Primary(), theme.Accent()
	success, dim := theme.Success(), theme.Dim()

	f := &t.Focused
	f.Base = lipgloss.NewStyle()
	f.Card = f.Base
	f.Title = lipgloss.NewStyle()
	f.Description = dim
	f.ErrorIndicator = theme.Danger().SetString(" *")
	f.ErrorMessage = theme.Danger()
	f.SelectSelector = accent.SetString("› ")
	f.Option = primary
	f.SelectedOption = success
	f.MultiSelectSelector = accent.SetString("› ")
	f.SelectedPrefix = success.SetString("◉ ")
	f.UnselectedPrefix = dim.SetString("○ ")
	f.UnselectedOption = primary
	f.NextIndicator = dim.MarginLeft(1).SetString("›")
	f.PrevIndicator = dim.MarginRight(1).SetString("‹")
	button := lipgloss.NewStyle().Padding(0, 2).MarginLeft(1)
	f.FocusedButton = focusedButton(button, accent)
	f.BlurredButton = button.Inherit(dim)
	f.TextInput.Cursor = accent
	f.TextInput.Placeholder = dim
	f.TextInput.Prompt = accent
	f.TextInput.Text = primary

	t.Blurred = t.Focused
	t.Help.ShortKey = dim
	t.Help.ShortDesc = dim
	t.Help.ShortSeparator = dim
	return t
}
