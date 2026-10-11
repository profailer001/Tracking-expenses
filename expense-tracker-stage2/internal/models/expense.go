package models

import "time"

type Expense struct {
	ID int
	Amount float64
	Description string
	Date time.Time
}