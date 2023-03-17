package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	stateForecastView State = iota
	stateEventView
)

type Tui struct {
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
	case tea.KeyMsg:
		switch {
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
			return t, nil
		case t.state == stateForecastView && key.Matches(msg, f.EditEvent):
			t.eventView.setEvent(t.forecastView.getSelectedTransaction().event)
			t.state = stateEventView
			return t, nil
		case t.state == stateForecastView && key.Matches(msg, f.Quit):
			return t, tea.Quit

		// EventView keypresses
		case t.state == stateEventView && key.Matches(msg, f.FocusTable):
			t.eventView.unsetEvent()
			t.state = stateForecastView
			return t, nil
		case t.state == stateEventView && key.Matches(msg, f.Confirm):
			// we must call getEvent in both add or edit mode: it pulls data from textinputs
			event := t.eventView.getEvent()
			if !t.eventView.hasEvent() {
				t.account.addEvent(event)
			}

			t.eventView.unsetEvent()
			t.state = stateForecastView
			t.forecastView.regenerateRows()
			return t, nil
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

	return b.String()
}

func newTui(account *Account) Tui {
	t := Tui{
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
