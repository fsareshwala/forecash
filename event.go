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

func (f Frequency) toString() string {
	switch f {
	case Once:
		return "Once"
	case Daily:
		return "Daily"
	case Weekly:
		return "Weekly"
	case Biweekly:
		return "Biweekly"
	case Monthly:
		return "Monthly"
	case Yearly:
		return "Yearly"
	}

	return "Unknown"
}

type Event struct {
	Date        time.Time
	Description string
	Amount      float32
	Frequency   Frequency
}

func (e *Event) nextOccurrence(from time.Time) time.Time {
	switch e.Frequency {
	case Daily:
		return from.AddDate(0, 0, 1)
	case Weekly:
		return from.AddDate(0, 0, 7)
	case Biweekly:
		return from.AddDate(0, 0, 14)
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
		t := Transaction{
			date:  now,
			event: e,
		}

		t.calculateHash()
		transactions = append(transactions, t)

		now = e.nextOccurrence(now)
	}

	return transactions

}
