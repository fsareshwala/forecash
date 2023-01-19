package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"time"
)

type Transaction struct {
	date  time.Time
	event *Event
}

func (t *Transaction) repeats() bool {
	return t.event.Frequency != Once
}

func (t *Transaction) isFirstOccurrence() bool {
	return t.date.Equal(t.event.Date)
}

type byDate []Transaction

func (t byDate) Len() int {
	return len(t)
}
func (t byDate) Swap(i int, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t byDate) Less(i int, j int) bool {
	return t[i].date.Before(t[j].date)
}

type Account struct {
	config_path string

	Balance float32
	Events  []Event
}

func newAccount(path string) Account {
	account_str, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var account Account
	if err := json.Unmarshal(account_str, &account); err != nil {
		log.Fatal(err)
	}

	account.config_path = path
	return account
}

func (a *Account) deleteEvent(i int) {
	last := len(a.Events) - 1
	a.Events[i] = a.Events[last]
	a.Events = a.Events[:last]
}

func (a *Account) save() {
	result, err := json.Marshal(a)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(a.config_path, result, 0); err != nil {
		log.Fatal(err)
	}
}

func (a *Account) reload() {
	account_str, err := os.ReadFile(a.config_path)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(account_str, a); err != nil {
		log.Fatal(err)
	}
}

func (a *Account) predict(until time.Time) []Transaction {
	transactions := []Transaction{}

	for i := range a.Events {
		event := &a.Events[i]
		transactions = append(transactions, event.predict(until)...)
	}

	sort.Sort(byDate(transactions))
	return transactions
}

func (a *Account) findEventIndex(tx *Transaction) int {
	for i, event := range a.Events {
		if *tx.event == event {
			return i
		}
	}

	return -1
}

func (a *Account) txComplete(tx *Transaction, update_balance bool) {
	if tx.repeats() && !tx.isFirstOccurrence() {
		// disallow marking done a future transaction generated by a repeating event
		return
	}

	if update_balance {
		a.Balance += tx.event.Amount
	}

	i := a.findEventIndex(tx)
	if !tx.repeats() {
		a.deleteEvent(i)
		return
	}

	tx.event.Date = tx.event.nextOccurrence(tx.event.Date)
}

func (a *Account) txDatePrevious(tx *Transaction) {
	if !tx.repeats() {
		tx.event.Date = tx.event.Date.AddDate(0, 0, -1)
	}
}

func (a *Account) txDateNext(tx *Transaction) {
	if !tx.repeats() {
		tx.event.Date = tx.event.Date.AddDate(0, 0, 1)
	}
}

func (a *Account) txSetToToday(tx *Transaction) {
	if tx.repeats() && !tx.isFirstOccurrence() {
		// doesn't make sense to place a future occurrence of a repeating event to be paid today when
		// there is an earlier occurrence
		return
	}

	if !tx.repeats() {
		// no need to do any splitting logic if the event never repeats
		tx.event.Date = time.Now().Round(0)
		return
	}

	new_event := *tx.event
	new_event.Frequency = Once
	new_event.Date = time.Now().Round(0)
	a.Events = append(a.Events, new_event)

	tx.event.Date = tx.event.nextOccurrence(tx.event.Date)

}
