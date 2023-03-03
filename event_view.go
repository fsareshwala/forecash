package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EventViewKeyMap struct {
	PreviousField key.Binding
	NextField     key.Binding
	Cancel        key.Binding
}

func NewEventViewKeyMap() EventViewKeyMap {
	return EventViewKeyMap{
		PreviousField: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "previous field"),
		),
		NextField: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "next field"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

type FocusedField int

const (
	month FocusedField = iota
	day
	year
	description
	amount
	// repeat
	sentinel
)

type EventView struct {
	keymap  EventViewKeyMap
	inputs  []textinput.Model
	focused int

	event *Event
}

func NewEventView() EventView {
	now := time.Now()
	inputs := make([]textinput.Model, sentinel)

	inputs[month] = textinput.New()
	inputs[month].Placeholder = fmt.Sprintf("%02d ", now.Month())
	inputs[month].CharLimit = 2
	inputs[month].Width = 2
	inputs[month].Prompt = ""
	inputs[month].Validate = validateMonth

	inputs[day] = textinput.New()
	inputs[day].Placeholder = fmt.Sprintf("%02d ", now.Day())
	inputs[day].CharLimit = 2
	inputs[day].Width = 2
	inputs[day].Prompt = ""
	inputs[day].Validate = validateDay

	inputs[year] = textinput.New()
	inputs[year].Placeholder = fmt.Sprintf("%d", now.Year())
	inputs[year].CharLimit = 4
	inputs[year].Width = 4
	inputs[year].Prompt = ""
	inputs[year].Validate = validateInteger

	inputs[description] = textinput.New()
	inputs[description].Placeholder = "Enter a description..."
	inputs[description].Prompt = ""

	inputs[amount] = textinput.New()
	inputs[amount].Placeholder = "123.45"
	inputs[amount].Prompt = "$"
	inputs[amount].Validate = validateFloat

	inputs[month].Focus()

	return EventView{
		keymap:  NewEventViewKeyMap(),
		inputs:  inputs,
		focused: 0,
	}
}

func (e *EventView) hasEvent() bool {
	return e.event != nil
}

func (e *EventView) getEvent() *Event {
	var event *Event
	if e.event == nil {
		event = &Event{
			Frequency: Once,
		}
	} else {
		event = e.event
	}

	// the textinput validator already ensures that the number is valid so no need to check for errors
	input_month, _ := strconv.ParseInt(e.inputs[month].Value(), 10, 8)
	input_day, _ := strconv.ParseInt(e.inputs[day].Value(), 10, 8)
	input_year, _ := strconv.ParseInt(e.inputs[year].Value(), 10, 16)
	input_amount, _ := strconv.ParseFloat(e.inputs[amount].Value(), 32)

	new_month := time.Month(input_month)
	new_day := int(input_day)
	new_year := int(input_year)
	new_amount := float32(input_amount)

	event.Date = time.Date(new_year, new_month, new_day, 0, 0, 0, 0, time.Local)
	event.Description = e.inputs[description].Value()
	event.Amount = new_amount

	// TODO: have to add ability to select a repeat to the form, then can do this logic correctly.
	// Until then, just use the frequency from the already existing event, or once for a new event.
	return event
}

func (e *EventView) setEvent(event *Event) {
	e.event = event
	e.inputs[month].SetValue(fmt.Sprintf("%02d", event.Date.Month()))
	e.inputs[day].SetValue(fmt.Sprintf("%02d", event.Date.Day()))
	e.inputs[year].SetValue(fmt.Sprintf("%d", event.Date.Year()))
	e.inputs[description].SetValue(event.Description)
	e.inputs[amount].SetValue(fmt.Sprintf("%.02f", event.Amount))

	// setting value apparently focuses the textinput
	for i := range e.inputs {
		e.inputs[i].Blur()
	}

	e.inputs[month].Focus()
}

func (e *EventView) unsetEvent() {
	e.event = nil

	for i := range e.inputs {
		e.inputs[i].Reset()
	}

	e.inputs[month].Focus()
}

func validateMonth(str string) error {
	// The textinput will already ensure that the number is of the correct length
	result, err := strconv.ParseInt(str, 10, 8)
	if err != nil {
		return err
	}

	if result < 1 || result > 12 {
		return fmt.Errorf("month is invalid")
	}

	return nil
}

func validateDay(str string) error {
	// The textinput will already ensure that the number is of the correct length
	result, err := strconv.ParseInt(str, 10, 8)
	if err != nil {
		return err
	}

	if result < 1 || result > 31 {
		return fmt.Errorf("day is invalid")
	}

	return nil
}

func validateInteger(str string) error {
	// The textinput will already ensure that the number is of the correct length
	_, err := strconv.ParseInt(str, 10, 16)
	return err
}

func validateFloat(str string) error {
	// The textinput will already ensure that the number is of the correct length
	_, err := strconv.ParseFloat(str, 32)
	return err
}

func (e *EventView) View() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("27"))

	var b strings.Builder
	b.WriteString(style.Render("Date"))
	b.WriteString("\n")
	b.WriteString(e.inputs[month].View())
	b.WriteString("/ ")
	b.WriteString(e.inputs[day].View())
	b.WriteString("/ ")
	b.WriteString(e.inputs[year].View())
	b.WriteString("\n\n")

	b.WriteString(style.Render("Description"))
	b.WriteString("\n")
	b.WriteString(e.inputs[description].View())
	b.WriteString("\n\n")

	b.WriteString(style.Render("Amount"))
	b.WriteString("\n")
	b.WriteString(e.inputs[amount].View())
	b.WriteString("\n\n")

	return b.String()
}

func (e *EventView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			e.nextInput()
		case tea.KeyShiftTab:
			e.prevInput()
		}

		for i := range e.inputs {
			e.inputs[i].Blur()
		}

		e.inputs[e.focused].Focus()
	}

	for i := range e.inputs {
		e.inputs[i], _ = e.inputs[i].Update(msg)
	}

	return nil
}

func (e *EventView) nextInput() {
	e.focused = (e.focused + 1) % len(e.inputs)
}

func (e *EventView) prevInput() {
	e.focused--

	if e.focused < 0 {
		e.focused = len(e.inputs) - 1
	}
}
