package models

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

type Session struct {
	SessionID string
	UserID    int64
	ExpiresAt time.Time
}