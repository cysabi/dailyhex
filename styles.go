package main

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// styles
var maxNameLen = 14
var mainWidth = 23
var formWidth = mainWidth - 2

type Styles struct {
	isDark       bool
	Base         lipgloss.Style
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	FormBox      lipgloss.Style
	NormalText   lipgloss.Style
	MovesCount   lipgloss.Style
	BoardGuesses lipgloss.Style
	BoardGrade   lipgloss.Style
	BoardArrows  lipgloss.Style
	GameBox      lipgloss.Style
	ColorBox     lipgloss.Style
	InputBox     lipgloss.Style
	MoveBox      lipgloss.Style
	Disabled     lipgloss.Style
	FormError    lipgloss.Style
	FormTheme    *huh.Theme
}

func (s Styles) New(r *lipgloss.Renderer, secret string) Styles {
	return Styles{
		isDark:       r.HasDarkBackground(),
		Base:         r.NewStyle(),
		Title:        r.NewStyle().Width(mainWidth).AlignHorizontal(lipgloss.Center).Bold(true),
		Subtitle:     r.NewStyle().Width(mainWidth).AlignHorizontal(lipgloss.Center).Foreground(lipgloss.Color("8")),
		NormalText:   r.NewStyle().Foreground(nil),
		MovesCount:   r.NewStyle().Width(3).AlignHorizontal(lipgloss.Right).Foreground(lipgloss.AdaptiveColor{Light: "7", Dark: "0"}),
		BoardGuesses: r.NewStyle().AlignHorizontal(lipgloss.Right),
		BoardGrade:   r.NewStyle().Width(2).Height(1),
		BoardArrows:  r.NewStyle().Foreground(nil).Bold(true),
		GameBox:      r.NewStyle().Width(mainWidth),
		ColorBox:     r.NewStyle().Width(2).Height(1).Margin(0, 1),
		InputBox:     r.NewStyle().Border(lipgloss.RoundedBorder()),
		MoveBox:      r.NewStyle().PaddingTop(1),
		Disabled:     r.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("8")),
		FormBox:      r.NewStyle().Width(mainWidth).PaddingLeft(2),
		FormError:    r.NewStyle().Width(formWidth).PaddingLeft(1).Foreground(lipgloss.Color("1")),
		FormTheme:    makeFormTheme(r, secret),
	}
}

// 0 8 7 secret
// bold

func makeFormTheme(r *lipgloss.Renderer, secret string) *huh.Theme {
	var t huh.Theme
	t.Form.Renderer(r)

	t.FieldSeparator = r.NewStyle().SetString("\n\n\n")

	// group
	t.Blurred.Base = r.NewStyle().BorderStyle(lipgloss.HiddenBorder()).BorderLeft(true)

	// prompts
	t.Blurred.SelectSelector = r.NewStyle().Foreground(lipgloss.Color("8")).Bold(true).SetString("> ")
	t.Blurred.TextInput.Prompt = r.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)

	// text
	t.Blurred.UnselectedOption = r.NewStyle().Foreground(lipgloss.Color("8"))
	t.Blurred.SelectedOption = r.NewStyle().Foreground(lipgloss.Color("8"))
	t.Blurred.TextInput.Text = r.NewStyle().Foreground(lipgloss.Color("8"))
	t.Blurred.TextInput.Placeholder = t.Blurred.TextInput.Text.Foreground(lipgloss.Color("8"))

	// ~ FOCUSED ~
	t.Focused = t.Blurred

	// prompts
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(nil)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(nil)

	// text
	t.Focused.UnselectedOption = r.NewStyle().Foreground(nil)
	t.Focused.SelectedOption = r.NewStyle().Foreground(lipgloss.Color("#" + secret))
	t.Focused.TextInput.Text = r.NewStyle().Foreground(lipgloss.Color("#" + secret))

	return &t
}
