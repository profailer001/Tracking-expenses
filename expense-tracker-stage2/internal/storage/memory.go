package storage

import (
	"sync"

	"expense-tracker-stage2/internal/models"
)

type MemoryStore struct {
	mu       sync.Mutex
	expenses []models.Expense
	nextID   int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		expenses: make([]models.Expense, 0),
		nextID:   1,
	}
}

func (s *MemoryStore) Add(e models.Expense) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e.ID = s.nextID
	s.nextID++
	s.expenses = append(s.expenses, e)
}

func (s *MemoryStore) All() []models.Expense {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]models.Expense, len(s.expenses))
	copy(result, s.expenses)
	return result
}