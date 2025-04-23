package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Board struct {
	state *state
	page  [][]string
}

func (m Board) New() Board {
	m.setPage()
	return m
}

func (m Board) Init() tea.Cmd {
	return textinput.Blink
}

func (m Board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyLeft:
			m.state.dayPage -= 1
			m.setPage()
		case tea.KeyRight:
			m.state.dayPage += 1
			m.setPage()
		}
	}

	return m, nil
}

func (m Board) View() string {
	var rows []string

	nameWidth := maxNameLen + 2
	tableHeight := m.state.height - 10
	tableWidth := m.state.width - 8
	movesWidth := tableWidth - nameWidth

	nameStyle := m.state.styles.BoardNames.Width(nameWidth)
	movesStyle := m.state.styles.BoardGuesses.Width(movesWidth)

	rows = append(rows,
		lipgloss.JoinHorizontal(0,
			nameStyle.Foreground(lipgloss.Color("8")).Render("name\n"),
			movesStyle.Foreground(lipgloss.Color("8")).Render("moves\n"),
		),
	)

	for i, row := range m.page {
		if i > tableHeight {
			break
		}

		moves := strings.Split(row[1], ",")
		if row[2] == "true" {
			moves = append(moves, m.state.secret)
		}
		lenMoves := len(moves)
		movesSpace := movesWidth/2 - 2
		if movesSpace < len(moves) {
			moves = moves[len(moves)-movesSpace:]
		}

		for i, hex := range moves {
			moves[i] = m.state.styles.BoardGrade.Background(lipgloss.Color("#" + hex)).Render("  ")
		}

		countColor := lipgloss.Color("7")
		if row[2] == "true" {
			countColor = lipgloss.Color("2")
		}

		moves = append(moves, m.state.styles.CharGrade.Width(3).AlignHorizontal(lipgloss.Right).Foreground(countColor).Render(fmt.Sprint(lenMoves)))
		rows = append(rows,
			lipgloss.JoinHorizontal(0,
				nameStyle.Render(row[0]),
				movesStyle.Render(lipgloss.JoinHorizontal(0, moves...)),
			),
		)
	}

	return lipgloss.NewStyle().Width(m.state.width - 8).Height(m.state.height - 8).Render(lipgloss.JoinVertical(0, rows...))
}

func (m *Board) setPage() {
	m.page = m.state.GetBoard()
}
