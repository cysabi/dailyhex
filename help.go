package main

import (
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type Help struct {
	state    *state
	helpText string
}

func (m Help) New() Help {
	file, err := os.ReadFile("README.md")
	if err != nil {
		log.Error(err)
	}
	if err != nil {
		log.Error(err)
	}
	log.Warn(m.state.styles.isDark)
	stylePath := "glamour-light.json"
	if m.state.styles.isDark {
		stylePath = "glamour-dark.json"
	}
	helpText, err := glamour.Render(string(file), stylePath)
	if err != nil {
		log.Error(err)
	}
	m.helpText = helpText

	return m
}

func (m Help) Init() tea.Cmd {
	return textinput.Blink
}

func (m Help) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Help) View() string {
	return m.state.styles.Base.Width(m.state.width - 8).Height(m.state.height - 8).AlignHorizontal(lipgloss.Center).Render(m.helpText)
}
