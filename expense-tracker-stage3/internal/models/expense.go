package models

import "time"

type Expense struct {
	ID          int64
	Amount      float64
	Description string
	Date        time.Time
}

func (e Expense) DateISO() string {
	return e.Date.Format("2006-01-02")
}

func (e Expense) DateRU() string {
	return e.Date.Format("02.01.2006")
}