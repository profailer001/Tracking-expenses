package storage

import (
	"database/sql"
	"time"

	"expense-tracker-stage3/internal/models"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) Create(e models.Expense) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO expenses (amount, description, date) VALUES (?, ?, ?)`,
		e.Amount, e.Description, e.DateISO(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *SQLiteStore) All() ([]models.Expense, error) {
	rows, err := s.db.Query(
		`SELECT id, amount, description, date FROM expenses ORDER BY date DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Expense
	for rows.Next() {
		var (
			e        models.Expense
			dateStr  string
		)
		if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &dateStr); err != nil {
			return nil, err
		}
		e.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

// GetByID возвращает (nil, nil), если записи нет.
func (s *SQLiteStore) GetByID(id int64) (*models.Expense, error) {
	var (
		e       models.Expense
		dateStr string
	)
	err := s.db.QueryRow(
		`SELECT id, amount, description, date FROM expenses WHERE id = ?`, id,
	).Scan(&e.ID, &e.Amount, &e.Description, &dateStr)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.Date, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *SQLiteStore) Update(e models.Expense) error {
	_, err := s.db.Exec(
		`UPDATE expenses SET amount = ?, description = ?, date = ? WHERE id = ?`,
		e.Amount, e.Description, e.DateISO(), e.ID,
	)
	return err
}

func (s *SQLiteStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM expenses WHERE id = ?`, id)
	return err
}