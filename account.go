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

func (a *Account) save() {
	result, err := json.Marshal(a)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(a.config_path, result, 0); err != nil {
		log.Fatal(err)
	}
}

func (a *Account) predict(until time.Time) []Transaction {
	transactions := []Transaction{}

	for _, event := range a.Events {
		transactions = append(transactions, event.predict(until)...)
	}

	sort.Sort(byDate(transactions))
	return transactions
}