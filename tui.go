package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type MainKeyMap struct {
	Confirm key.Binding
	Help    key.Binding
	Save    key.Binding
	Quit    key.Binding
}

func NewMainKeyMap() MainKeyMap {
	return MainKeyMap{
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Save: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "save"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k MainKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k MainKeyMap) FullHelp() [][]key.Binding {
	f := NewForecastViewKeyMap()
	return [][]key.Binding{
		{f.LineUp, f.LineDown, f.GotoTop, f.GotoBottom},
		{f.DatePrevious, f.DateNext, f.SetToday, f.Done, f.Delete},
		{f.EditBalance, k.Confirm, f.FocusTable},
		{f.Reload, k.Save, k.Quit},
	}
}

type Tui struct {
	keymap MainKeyMap
	help   help.Model

	account      *Account
	forecastView ForecastView
}

func (t Tui) Init() tea.Cmd {
	return nil
}

func (t Tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.keymap.Help):
			t.help.ShowAll = !t.help.ShowAll
			return t, nil
		case key.Matches(msg, t.keymap.Save):
			t.account.save()
			return t, nil
		case key.Matches(msg, t.keymap.Quit):
			return t, tea.Quit
		}
	}

	cmd := t.forecastView.Update(msg)
	return t, cmd
}

func (t Tui) View() string {
	var b strings.Builder
	b.WriteString(t.forecastView.View())
	b.WriteString(t.help.View(t.keymap))
	b.WriteString("\n")
	return b.String()
}

func newTui(account *Account) Tui {
	return Tui{
		keymap: NewMainKeyMap(),
		help:   help.New(),

		account:      account,
		forecastView: NewForecastView(account),
	}
}

func (t *Tui) run() {
	if _, err := tea.NewProgram(*t).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
