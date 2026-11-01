package storage

import (
	"database/sql"
	"strings"
	"time"

	"expense-tracker-stage4/internal/models"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) Categories() ([]models.Category, error) {
	rows, err := s.db.Query(`SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (s *SQLiteStore) Create(e models.Expense) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO expenses (amount, description, date, category_id) VALUES (?, ?, ?, ?)`,
		e.Amount, e.Description, e.DateISO(), e.CategoryID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// All возвращает траты с названием категории одним JOIN-запросом.
// WHERE собирается динамически по заданным фильтрам.
func (s *SQLiteStore) All(f models.ExpenseFilter) ([]models.Expense, error) {
	var (
		conds []string
		args  []any
	)
	if f.CategoryID > 0 {
		conds = append(conds, "e.category_id = ?")
		args = append(args, f.CategoryID)
	}
	if !f.From.IsZero() {
		conds = append(conds, "e.date >= ?")
		args = append(args, f.From.Format("2006-01-02"))
	}
	if !f.To.IsZero() {
		conds = append(conds, "e.date <= ?")
		args = append(args, f.To.Format("2006-01-02"))
	}

	q := `SELECT e.id, e.amount, e.description, e.date, e.category_id, c.name
	      FROM expenses e
	      JOIN categories c ON c.id = e.category_id`
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY e.date DESC, e.id DESC"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Expense
	for rows.Next() {
		var (
			e       models.Expense
			dateStr string
		)
		if err := rows.Scan(&e.ID, &e.Amount, &e.Description, &dateStr,
			&e.CategoryID, &e.CategoryName); err != nil {
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

func (s *SQLiteStore) GetByID(id int64) (*models.Expense, error) {
	var (
		e       models.Expense
		dateStr string
	)
	err := s.db.QueryRow(
		`SELECT e.id, e.amount, e.description, e.date, e.category_id, c.name
		 FROM expenses e JOIN categories c ON c.id = e.category_id
		 WHERE e.id = ?`, id,
	).Scan(&e.ID, &e.Amount, &e.Description, &dateStr, &e.CategoryID, &e.CategoryName)

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
		`UPDATE expenses SET amount = ?, description = ?, date = ?, category_id = ? WHERE id = ?`,
		e.Amount, e.Description, e.DateISO(), e.CategoryID, e.ID,
	)
	return err
}

func (s *SQLiteStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM expenses WHERE id = ?`, id)
	return err
}