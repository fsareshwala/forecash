package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leekchan/accounting"
	"golang.org/x/term"
)

type KeyMap struct {
	DatePrevious key.Binding
	DateNext     key.Binding
	Delete       key.Binding
	Done         key.Binding
	SetToday     key.Binding

	LineUp     key.Binding
	LineDown   key.Binding
	GotoTop    key.Binding
	GotoBottom key.Binding

	Help key.Binding

	Reload key.Binding
	Save   key.Binding
	Quit   key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.LineUp, k.LineDown, k.GotoTop, k.GotoBottom},
		{k.DatePrevious, k.DateNext, k.SetToday, k.Done, k.Delete},
		{k.Reload, k.Save, k.Quit},
	}
}

type Tui struct {
	keys    KeyMap
	help    help.Model
	table   table.Model

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
	case tea.WindowSizeMsg:
		t.help.Width = msg.Width

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.keys.DatePrevious):
			t.account.txDatePrevious(&tx)
		case key.Matches(msg, t.keys.DateNext):
			t.account.txDateNext(&tx)
		case key.Matches(msg, t.keys.Delete):
			t.account.txComplete(&tx, false)
		case key.Matches(msg, t.keys.Done):
			t.account.txComplete(&tx, true)
		case key.Matches(msg, t.keys.SetToday):
			t.account.txSetToToday(&tx)

		case key.Matches(msg, t.keys.Help):
			t.help.ShowAll = !t.help.ShowAll

		case key.Matches(msg, t.keys.Reload):
			t.account.reload()
		case key.Matches(msg, t.keys.Save):
			t.account.save()
		case key.Matches(msg, t.keys.Quit):
			return t, tea.Quit
		}
	}

	t.regenerateRows()
	t.table, cmd = t.table.Update(msg)
	return t, cmd
}

func (t Tui) View() string {
	ac := accounting.Accounting{Symbol: "$", Precision: 2}

	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.HiddenBorder()).
		BorderForeground(lipgloss.Color("240"))

	var b strings.Builder
	b.WriteString(strings.Repeat(" ", 83))
	b.WriteString("Current balance: ")
	b.WriteString(ac.FormatMoney(t.account.Balance))
	b.WriteString("\n")
	b.WriteString(style.Render(t.table.View()))
	b.WriteString("\n")
	b.WriteString(t.help.View(t.keys))
	b.WriteString("\n")

	return b.String()
}

func newTui(account *Account) Tui {
	columns := []table.Column{
		{Title: "Date", Width: 20},
		{Title: "Description", Width: 40},
		{Title: "Income", Width: 15},
		{Title: "Expense", Width: 15},
		{Title: "Balance", Width: 15},
	}

	style := table.DefaultStyles()
	style.Header = style.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

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

	keyMap := KeyMap{
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

		LineUp:     t.KeyMap.LineUp,
		LineDown:   t.KeyMap.LineDown,
		GotoTop:    t.KeyMap.GotoTop,
		GotoBottom: t.KeyMap.GotoBottom,

		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),

		Reload: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reload"),
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

	tui := Tui{
		keys:         keyMap,
		help:         help.New(),
		table:        t,
		account:      account,
		transactions: nil,
	}

	_, term_height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	tui.regenerateRows()
	height := len(tui.transactions)
	if height > term_height {
		height = term_height
	}

	tui.table.SetHeight(height)
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
		balance_str := ac.FormatMoney(balance)
		if balance < 0 {
			balance_str = fmt.Sprintf(("\u001b[31m%s"), balance_str)
		}

		rows = append(rows, table.Row{
			t.date.Format("January 2, 2006"),
			t.event.Description,
			income,
			expense,
			balance_str,
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
