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
		{f.DatePrevious, f.DateNext, f.SetToday, f.Done, f.Delete, f.EditEvent, f.AddEvent},
		{f.EditBalance, k.Confirm, f.FocusTable},
		{f.Reload, f.Save, k.Quit},
	}
}

type State int

const (
	stateForecastView State = iota
	stateEventView
)

type Tui struct {
	keymap       MainKeyMap
	help         help.Model
	forecastView ForecastView
	eventView    EventView

	state   State
	account *Account
}

func (t Tui) Init() tea.Cmd {
	return nil
}

func (t Tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	f := NewForecastViewKeyMap()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.help.Width = msg.Width
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.keymap.Help):
			t.help.ShowAll = !t.help.ShowAll
			return t, nil
		case key.Matches(msg, t.keymap.Quit):
			return t, tea.Quit

		// Although these keypresses are enabled only for specific views, we check here because there is
		// no way for a subview to inform the parent view to switch to another subview (e.g. going from
		// ForecastView to EventView).
		//
		// We could pass the Tui instance to the subviews as well but it (a) creates a bidirectional
		// pointer relationship and (b) doesn't work due to the Model interface passing by value and
		// creating a new copy of Tui each time Update or View is called.
		//
		// ForecastView keypresses
		case t.state == stateForecastView && key.Matches(msg, f.AddEvent):
			t.eventView.unsetEvent()
			t.state = stateEventView
		case t.state == stateForecastView && key.Matches(msg, f.EditEvent):
			t.eventView.setEvent(t.forecastView.getSelectedTransaction().event)
			t.state = stateEventView

		// EventView keypresses
		case t.state == stateEventView && key.Matches(msg, f.FocusTable):
			t.eventView.unsetEvent()
			t.state = stateForecastView
		case t.state == stateEventView && key.Matches(msg, t.keymap.Confirm):
			// we must call getEvent in both add or edit mode: it pulls data from textinputs
			event := t.eventView.getEvent()
			if !t.eventView.hasEvent() {
				t.account.addEvent(event)
			}

			t.eventView.unsetEvent()
			t.state = stateForecastView
		}
	}

	var cmd tea.Cmd
	switch t.state {
	case stateForecastView:
		cmd = t.forecastView.Update(msg)
	case stateEventView:
		cmd = t.eventView.Update(msg)
	}

	return t, cmd
}

func (t Tui) View() string {
	var b strings.Builder

	switch t.state {
	case stateForecastView:
		b.WriteString(t.forecastView.View())
	case stateEventView:
		b.WriteString(t.eventView.View())
	}

	b.WriteString(t.help.View(t.keymap))
	b.WriteString("\n")
	return b.String()
}

func newTui(account *Account) Tui {
	t := Tui{
		keymap:       NewMainKeyMap(),
		help:         help.New(),
		forecastView: NewForecastView(account),
		eventView:    NewEventView(),

		state:   stateForecastView,
		account: account,
	}

	// bidirectional relationship between main view and subviews, necessary so that subviews can
	// instruct main view to switch to another subview
	return t
}

func (t *Tui) run() {
	if _, err := tea.NewProgram(*t).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
