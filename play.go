package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Play struct {
	state    *state
	Input    textinput.Model
	Viewport viewport.Model
	movePos  int
}

func (m Play) New() Play {
	ti := textinput.New()
	ti.CharLimit = 6
	ti.Width = 6
	ti.Prompt = " ?= "
	ti.PromptStyle = m.state.styles.Base.Foreground(lipgloss.Color("8"))
	ti.Placeholder = "######"
	ti.PlaceholderStyle = m.state.styles.FormTheme.Blurred.TextInput.Placeholder
	ti.Focus()

	m.Input = ti
	m.Viewport = viewport.New(16, m.state.height-10)
	m.Viewport.SetContent(lipgloss.JoinVertical(0, m.displayMoves()...))
	m.state.gameState = Idle
	m.movePos = -1 // initialize move (guess) history at -1
	return m
}

func (m Play) Init() tea.Cmd {
	return tea.Batch(textinput.Blink)
}

func (m Play) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		// Up key: get previous move(s) in guess history
		case tea.KeyUp:
			moves := m.state.GetMoves()
			if len(moves) > 0 {
				if m.movePos < len(moves)-1 {
					m.movePos++
					m.Input.SetValue(moves[len(moves)-1-m.movePos])
				}
			}
			return m, nil // Return early to prevent viewport processing

			// Down key: move forward in guess history, if already viewing history
		case tea.KeyDown:
			if m.movePos > 0 {
				m.movePos--
				moves := m.state.GetMoves()
				m.Input.SetValue(moves[len(moves)-1-m.movePos])
			} else if m.movePos == 0 {
				m.movePos = -1
				m.Input.SetValue("")
			}
			return m, nil // Return early to prevent viewport processing

		case tea.KeyEnter:
			move := m.Input.Value()
			m.movePos = -1 // reset movePos

			if len(move) != 6 {
				m.state.gameState = Invalid
			} else if m.state.secret == move {
				m.state.SetDone(true)
				m.state.gameState = Win
			} else {
				m.state.AppendMove(move)
				m.Viewport.SetContent(lipgloss.JoinVertical(0, m.displayMoves()...))
				m.state.gameState = Idle
			}

			if m.state.gameState == Win {
				m.Input.Blur()
			}
			if m.state.gameState == Idle {
				m.Input.SetValue("")
			}

		default:
			if m.state.gameState == Invalid {
				m.state.gameState = Idle
			}
		}
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	m.Input, cmd = m.Input.Update(msg)
	cmds = append(cmds, cmd)

	m.Input.SetValue(strings.ToLower(toHex(m.Input.Value())))

	return m, tea.Batch(cmds...)
}

func (m Play) View() string {
	input := m.state.styles.InputBox.BorderForeground(lipgloss.Color(m.state.gameState)).Render(
		lipgloss.JoinHorizontal(0,
			m.state.styles.ColorBox.Background(lipgloss.Color("#"+m.state.secret)).Render(),
			m.Input.View(),
		),
	)

	return lipgloss.JoinVertical(0,
		input,
		m.Viewport.View(),
	)
}

func (m Play) StateMsg() string {
	if m.state.gameState == Invalid {
		return "invalid hex"
	} else if m.state.gameState == Win {
		return fmt.Sprintf("you got it!! (%d turns)", len(m.state.GetMoves())+1)
	}
	return ""
}

type CharGrade string

const (
	Green  CharGrade = "2"
	Yellow CharGrade = "3"
	Gray   CharGrade = "8"
)

func (m Play) displayMoves() []string {
	moves := m.state.GetMoves()
	slices.Reverse(moves)
	out := make([]string, len(moves))

	for i, move := range moves {
		grade := m.gradeMove(move)
		out[i] = m.displayMove(move, grade)
	}

	return out

}

func (m Play) gradeMove(move string) []CharGrade {
	grade := make([]CharGrade, len(m.state.secret))
	secret := []rune(m.state.secret)

	for i, s := range secret {
		if s == []rune(move)[i] {
			grade[i] = Green
			secret[i] = ' '
		} else {
			grade[i] = Gray
		}
	}

	for _, s := range secret {
		if s == ' ' {
			continue
		}
		for i, m := range move {
			if m == s && grade[i] == Gray {
				grade[i] = Yellow
				break
			}
		}
	}
	return grade
}

func (m Play) displayMove(move string, grade []CharGrade) string {
	var text strings.Builder
	for index, letter := range move {
		str := m.state.styles.Base.Foreground(lipgloss.Color(grade[index])).Render(string(letter))
		text.WriteString(str)
	}

	return m.state.styles.MoveBox.Render(
		lipgloss.JoinHorizontal(0,
			m.state.styles.ColorBox.Background(lipgloss.Color("#"+move)).Render(),
			m.Input.PromptStyle.Render(" == "),
			text.String(),
		))
}

func toHex(str string) string {
	var hex []rune
	for _, r := range str {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			hex = append(hex, r)
		}
	}
	return string(hex)
}
