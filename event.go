package main

import (
	"time"
)

type Frequency int

const (
	Once Frequency = iota
	Daily
	Weekly
	Biweekly
	Monthly
	Yearly
)

type Event struct {
	Date        time.Time
	Description string
	Amount      float32
	Frequency   Frequency
}

func (e *Event) nextOccurrence(from time.Time) time.Time {
	switch e.Frequency {
	case Daily:
		return from.Add(time.Hour * 24)
	case Weekly:
		return from.Add(time.Hour * 24 * 7)
	case Biweekly:
		return from.Add(time.Hour * 24 * 7 * 2)
	case Monthly:
		return from.AddDate(0, 1, 0)
	case Yearly:
		return from.AddDate(1, 0, 0)
	}

	// events that don't repeat have no next occurrence so return some really far out date
	return e.Date.AddDate(100, 0, 0)
}

func (e *Event) predict(until time.Time) []Transaction {
	transactions := []Transaction{}

	now := e.Date
	for now.Before(until) {
		transactions = append(transactions, Transaction{
			date:  now,
			event: e,
		})

		now = e.nextOccurrence(now)
	}

	return transactions

}
