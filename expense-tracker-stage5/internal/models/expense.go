package models

import "time"

type Category struct {
	ID   int64
	Name string
}

type Expense struct {
	ID           int64
	Amount       float64
	Description  string
	Date         time.Time
	CategoryID   int64
	CategoryName string
	UserID       int64
}

type ExpenseFilter struct {
	CategoryID int64
	From       time.Time
	To         time.Time
	UserID     int64
}

func (e Expense) DateISO() string { return e.Date.Format("2006-01-02") }
func (e Expense) DateRU() string  { return e.Date.Format("02.01.2006") }