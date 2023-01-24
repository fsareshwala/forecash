package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leekchan/accounting"
	"golang.org/x/term"
)

type Tui struct {
	table table.Model

	account      *Account
	transactions []Transaction
}

func (t Tui) Init() tea.Cmd {
	return nil
}

func (t Tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	tx := t.transactions[t.table.Cursor()]
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h":
			t.account.txDatePrevious(&tx)
		case "l":
			t.account.txDateNext(&tx)
		case "d":
			t.account.txComplete(&tx, false)
		case "x":
			t.account.txComplete(&tx, true)
		case "t":
			t.account.txSetToToday(&tx)
		case "r":
			t.account.reload()
		case "w":
			t.account.save()
		case "q":
			return t, tea.Quit
		}
	}

	t.regenerateRows()
	t.table, cmd = t.table.Update(msg)
	return t, cmd
}

func (t Tui) View() string {
	ac := accounting.Accounting{Symbol: "$", Precision: 2}
	formatted_balance := ac.FormatMoney(t.account.Balance)

	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
	table := style.Render(t.table.View())

	prefix := strings.Repeat(" ", 83)
	return fmt.Sprintf("%sCurrent balance: %s\n%s\n", prefix, formatted_balance, table)
}

func newTui(account *Account) Tui {
	columns := []table.Column{
		{Title: "Date", Width: 20},
		{Title: "Description", Width: 40},
		{Title: "Income", Width: 15},
		{Title: "Expense", Width: 15},
		{Title: "Balance", Width: 15},
	}

	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	t := table.New(
		table.WithFocused(true),
		table.WithHeight(height*3/4),
		table.WithWidth(width),
		table.WithColumns(columns),
	)

	t.KeyMap.HalfPageUp = key.NewBinding(key.WithDisabled())
	t.KeyMap.HalfPageDown = key.NewBinding(key.WithDisabled())

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(s)

	tui := Tui{
		table:   t,
		account: account,
	}

	tui.regenerateRows()
	return tui
}

func (t *Tui) regenerateRows() {
	until := time.Now().AddDate(0, 4, 0)
	t.transactions = t.account.predict(until)

	ac := accounting.Accounting{Symbol: "$", Precision: 2}
	balance := t.account.Balance

	rows := make([]table.Row, 0, len(t.transactions))
	for _, t := range t.transactions {
		var income string
		var expense string

		if t.event.Amount > 0 {
			income = ac.FormatMoney(t.event.Amount)
		} else {
			expense = ac.FormatMoney(t.event.Amount * -1)
		}

		balance += t.event.Amount
		rows = append(rows, table.Row{
			t.date.Format("January 2, 2006"),
			t.event.Description,
			income,
			expense,
			ac.FormatMoney(balance),
		})
	}

	t.table.SetRows(rows)
}

func (t *Tui) run() {
	if _, err := tea.NewProgram(*t).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
