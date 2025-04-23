package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type Title struct {
	state *state
	Form  *huh.Form
}

func (m Title) New() Title {
	username := m.state.GetName()

	playOption := huh.NewOption(string(PlayScreen), PlayScreen)
	boardOption := huh.NewOption(string(BoardScreen), BoardScreen)

	if m.state.GetDone() {
		playOption.Key = m.state.styles.Disabled.Render(string(PlayScreen))
		if !m.state.showCountdown {
			boardOption = boardOption.Selected(true)
		}
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("name").Value(&username).CharLimit(maxNameLen).Placeholder("what's ur name?").Prompt("? ").Validate(
				func(str string) error {
					if len(str) == 0 {
						return errors.New("what's ur name!?")
					}
					return nil
				},
			),
			huh.NewSelect[Screen]().
				Key("screen").
				Options(playOption, boardOption),
		),
	).WithWidth(formWidth).WithShowHelp(false).WithShowErrors(false).WithTheme(m.state.styles.FormTheme)

	if len(username) > 0 {
		form.NextField()
	}

	m.Form = form
	return m
}

func (m Title) Init() tea.Cmd {
	return tea.Batch(m.Form.Init(), textinput.Blink)
}

func (m Title) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.Form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.Form = f
	}

	return m, cmd
}

func (m Title) View() string {
	errs := m.Form.Errors()
	if len(errs) > 0 {
		return m.state.styles.FormBox.Render(lipgloss.JoinVertical(0,
			m.Form.View(),
			m.state.styles.FormError.Render("* "+errs[0].Error()),
		))
	} else {
		return m.state.styles.FormBox.Render(m.Form.View())
	}

}

func dist() string {
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now().In(loc)

	next11 := time.Date(now.Year(), now.Month(), now.Day(), 11, 0, 0, 0, loc)
	if now.After(next11) {
		next11 = next11.Add(24 * time.Hour)
	}

	diff := next11.Sub(now)

	hours := int(diff.Hours())
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60

	if hours > 0 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	} else if minutes > 0 {
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	} else {
		if seconds == 1 {
			return "1 second"
		}
		return fmt.Sprintf("%d seconds", seconds)
	}
}
