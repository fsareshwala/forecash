package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type ForecastViewKeyMap struct {
	DatePrevious key.Binding
	DateNext     key.Binding
	Delete       key.Binding
	Done         key.Binding
	SetToday     key.Binding
	Reload       key.Binding

	FocusTable  key.Binding
	EditBalance key.Binding

	LineUp     key.Binding
	LineDown   key.Binding
	GotoTop    key.Binding
	GotoBottom key.Binding
}

func NewForecastViewKeyMap() ForecastViewKeyMap {
	return ForecastViewKeyMap{
		DatePrevious: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "date to previous day"),
		),
		DateNext: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "date to next day"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Done: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "done"),
		),
		SetToday: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "set to today"),
		),
		Reload: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reload"),
		),

		FocusTable: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "focus table"),
		),
		EditBalance: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "edit balance"),
		),

		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
	}
}

type ForecastView struct {
	keymap  ForecastViewKeyMap
	table   table.Model
	balance textinput.Model

	account      *Account
	transactions []Transaction
}

func NewForecastView(account *Account) ForecastView {
	columns := []table.Column{
		{Title: "Date", Width: 20},
		{Title: "Description", Width: 40},
		{Title: "Income", Width: 15},
		{Title: "Expense", Width: 15},
		{Title: "Balance", Width: 15},
	}

	style := table.DefaultStyles()
	style.Header = style.Header.
		Bold(false).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true)

	style.Selected = style.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("27")).
		Bold(false)

	t := table.New(
		table.WithFocused(true),
		table.WithColumns(columns),
		table.WithStyles(style),
	)

	t.KeyMap.PageUp = key.NewBinding(key.WithDisabled())
	t.KeyMap.PageDown = key.NewBinding(key.WithDisabled())
	t.KeyMap.HalfPageUp = key.NewBinding(key.WithDisabled())
	t.KeyMap.HalfPageDown = key.NewBinding(key.WithDisabled())

	b := textinput.New()
	b.Prompt = "Current balance: "

	f := ForecastView{
		keymap:  NewForecastViewKeyMap(),
		table:   t,
		balance: b,

		account:      account,
		transactions: nil,
	}

	f.regenerateRows()
	return f
}

func (f *ForecastView) regenerateRows() {
	until := time.Now().AddDate(0, 4, 0)
	f.transactions = f.account.predict(until)

	balance := f.account.Balance

	rows := make([]table.Row, 0, len(f.transactions))
	for _, transaction := range f.transactions {
		var income string
		var expense string

		if transaction.event.Amount > 0 {
			income = f.account.currency.FormatMoney(transaction.event.Amount)
		} else {
			expense = f.account.currency.FormatMoney(transaction.event.Amount * -1)
		}

		balance += transaction.event.Amount
		balance_str := f.account.currency.FormatMoney(balance)
		if balance < 0 {
			balance_str = fmt.Sprintf(("\x1b[31m%s\x1b[0m"), balance_str)
		}

		rows = append(rows, table.Row{
			transaction.date.Format("January 2, 2006"),
			transaction.event.Description,
			income,
			expense,
			balance_str,
		})
	}

	_, term_height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	height := len(f.transactions)
	if height > term_height {
		height = term_height
	}

	f.table.SetHeight(height)
	f.table.SetRows(rows)
}

func (f *ForecastView) View() string {
	if !f.balance.Focused() {
		f.balance.SetValue(f.account.currency.FormatMoney(f.account.Balance))
		f.balance.Blur()
	}

	var b strings.Builder
	b.WriteString(strings.Repeat(" ", 82))
	b.WriteString(f.balance.View())
	b.WriteString("\n\n")
	b.WriteString(f.table.View())
	b.WriteString("\n\n")
	return b.String()
}

func (f *ForecastView) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if f.table.Focused() {
			cmd = f.handleTableInput(msg)
		} else if f.balance.Focused() {
			cmd = f.handleBalanceInput(msg)
		}
	}

	f.regenerateRows()
	return cmd
}

func (f *ForecastView) handleTableInput(msg tea.Msg) tea.Cmd {
	tx := f.transactions[f.table.Cursor()]

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keymap.DatePrevious):
			f.account.txDatePrevious(&tx)
		case key.Matches(msg, f.keymap.DateNext):
			f.account.txDateNext(&tx)
		case key.Matches(msg, f.keymap.Delete):
			f.account.txComplete(&tx, false)
		case key.Matches(msg, f.keymap.Done):
			f.account.txComplete(&tx, true)
		case key.Matches(msg, f.keymap.SetToday):
			f.account.txSetToToday(&tx)
		case key.Matches(msg, f.keymap.Reload):
			f.account.reload()

		case key.Matches(msg, f.keymap.EditBalance):
			f.table.Blur()
			f.balance.SetValue(fmt.Sprintf("%.02f", f.account.Balance))
			f.balance.CursorEnd()
			f.balance.Focus()
		}
	}

	var cmd tea.Cmd
	f.table, cmd = f.table.Update(msg)
	return cmd
}

func (f *ForecastView) handleBalanceInput(msg tea.Msg) tea.Cmd {
	mainKeyMap := NewMainKeyMap()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, f.keymap.FocusTable):
			f.balance.Blur()
			f.table.Focus()
		case key.Matches(msg, mainKeyMap.Confirm):
			if result, err := strconv.ParseFloat(f.balance.Value(), 32); err == nil {
				f.account.Balance = float32(result)
			}

			f.balance.Blur()
			f.table.Focus()
		}
	}

	var cmd tea.Cmd
	f.balance, cmd = f.balance.Update(msg)
	return cmd
}