package selection

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type KeyMap struct {
	Previous key.Binding
	Next     key.Binding
}

func NewKeyMap() KeyMap {
	return KeyMap{
		Previous: key.NewBinding(key.WithKeys("k", "up", "left")),
		Next:     key.NewBinding(key.WithKeys("j", "down", "right")),
	}
}

type Model struct {
	Cursor string
	KeyMap KeyMap

	choices  []string
	selected int
	focus    bool

	SelectedStyle lipgloss.Style
	TextStyle     lipgloss.Style
}

func New(choices []string) Model {
	return Model{
		Cursor: ">",
		KeyMap: NewKeyMap(),

		choices:  choices,
		selected: 0,
		focus:    false,

		SelectedStyle: lipgloss.NewStyle(),
		TextStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}
}

func (m *Model) Focus() {
	m.focus = true
}

func (m *Model) Blur() {
	m.focus = false
}

func (m *Model) Reset() {
	m.selected = 0
}

func (m Model) Selected() int {
	return m.selected
}

func (m *Model) SetSelected(i int) error {
	if i >= len(m.choices) {
		return fmt.Errorf("cannot set selected to %d, it's greater than the number of choices (%d)", i,
			len(m.choices))
	}

	m.selected = i
	return nil
}

func (m *Model) previous() {
	m.selected--

	if m.selected < 0 {
		m.selected = len(m.choices) - 1
	}
}

func (m *Model) next() {
	m.selected = (m.selected + 1) % len(m.choices)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Previous):
			m.previous()
		case key.Matches(msg, m.KeyMap.Next):
			m.next()
		}
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	for i := 0; i < len(m.choices); i++ {
		if i == m.selected {
			style := m.TextStyle
			if m.focus {
				style = m.SelectedStyle
			}

			b.WriteString(style.Render(m.Cursor))
			b.WriteString(" ")
			b.WriteString(style.Render(m.choices[i]))
		} else {
			b.WriteString(strings.Repeat(" ", len(m.Cursor)+1))
			b.WriteString(m.TextStyle.Render(m.choices[i]))
		}

		if i != len(m.choices)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
