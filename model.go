package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	state *state

	// screens
	Title Title
	Play  Play
	Board Board
	Help  Help
}

func (m Model) New() Model {
	m.Title = Title{state: m.state}.New()
	m.Play = Play{state: m.state}
	m.Board = Board{state: m.state}
	m.Help = Help{state: m.state}
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.Title.Init(), textinput.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.state.height = msg.Height
		m.state.width = msg.Width
		m.Play.Viewport.Height = m.state.height - 10

	case tea.KeyMsg:
		m.state.showCountdown = false
		if m.state.screen == PlayScreen && m.state.gameState == Win {
			m.state.screen = TitleScreen
			m.Title = m.Title.New()
			return m, nil
		}
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.state.screen == TitleScreen {
				return m, tea.Quit
			}
			m.state.screen = TitleScreen
			m.Title = m.Title.New()
			return m, nil
		}
	}

	switch m.state.screen {
	case TitleScreen:
		title, cmd := m.Title.Update(msg)
		if t, ok := title.(Title); ok {
			m.Title = t
		}
		cmds = append(cmds, cmd)

		if m.Title.Form.State == huh.StateCompleted {

			m.state.SetName(m.Title.Form.Get("name").(string))
			newScreen := m.Title.Form.Get("screen").(Screen)

			if newScreen == PlayScreen {
				if m.state.GetDone() {
					m.state.showCountdown = true
					m.Title = m.Title.New()
				} else {
					m.state.screen = newScreen
					m.Play = m.Play.New()
				}
			}
			if newScreen == BoardScreen {
				m.state.screen = newScreen
				m.Board = m.Board.New()
			}
			if newScreen == HelpScreen {
				m.state.screen = newScreen
				m.Help = m.Help.New()
			}
		}

	case PlayScreen:
		game, cmd := m.Play.Update(msg)
		if g, ok := game.(Play); ok {
			m.Play = g
		}
		cmds = append(cmds, cmd)

	case BoardScreen:
		board, cmd := m.Board.Update(msg)
		if b, ok := board.(Board); ok {
			m.Board = b
		}
		cmds = append(cmds, cmd)

	case HelpScreen:
		help, cmd := m.Help.Update(msg)
		if b, ok := help.(Help); ok {
			m.Help = b
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	subtitle := m.state.styles.Subtitle.Render("day " + fmt.Sprint(m.state.day))
	rightArrow := " >"
	if m.state.dayPage == 0 {
		rightArrow = " ·"
	}
	if m.state.screen == BoardScreen {
		styl := m.state.styles.BoardArrows
		subtitle = m.state.styles.Subtitle.Render(styl.Render("< ") + m.state.styles.Base.Foreground(lipgloss.Color("8")).Render("day "+fmt.Sprint(m.state.day+m.state.dayPage)) + styl.Render(rightArrow))
	} else if m.state.gameState != Idle {
		subtitle = m.state.styles.Subtitle.Foreground(lipgloss.Color(m.state.gameState)).Render(m.Play.StateMsg())
	}

	color := m.state.secret
	if m.state.screen == BoardScreen {
		color = secret(m.state.day + m.state.dayPage)
	}

	banner := m.state.styles.Base.Margin(2).Render(
		lipgloss.JoinVertical(0,
			m.state.styles.Title.Foreground(lipgloss.Color("#"+color)).Render("dailyhex!"),
			subtitle,
		),
	)
	switch m.state.screen {
	case TitleScreen:
		view := m.Title.View()
		if m.state.showCountdown {
			view = lipgloss.JoinVertical(0,
				view,
				m.state.styles.FormError.Render("* opens in "+dist()))
		}
		return lipgloss.Place(m.state.width, m.state.height, lipgloss.Center, lipgloss.Top,
			lipgloss.JoinVertical(lipgloss.Center, banner, view))
	case PlayScreen:
		return lipgloss.Place(m.state.width, m.state.height, lipgloss.Center, lipgloss.Top,
			lipgloss.JoinVertical(lipgloss.Center, banner, m.Play.View()))
	case BoardScreen:
		return lipgloss.Place(m.state.width, m.state.height, lipgloss.Center, lipgloss.Top,
			lipgloss.JoinVertical(lipgloss.Center, banner, m.Board.View()))
	case HelpScreen:
		return lipgloss.Place(m.state.width, m.state.height, lipgloss.Center, lipgloss.Top,
			lipgloss.JoinVertical(lipgloss.Center, banner, m.Help.View()))
	}
	return "uh oh"
}
