package storage

import (
	"database/sql"
	"strings"
	"time"

	"expense-tracker-stage5/internal/models"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// ---------- Категории ----------

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

// ---------- Траты ----------

func (s *SQLiteStore) Create(e models.Expense) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO expenses (amount, description, date, category_id, user_id)
		 VALUES (?, ?, ?, ?, ?)`,
		e.Amount, e.Description, e.DateISO(), e.CategoryID, e.UserID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *SQLiteStore) All(f models.ExpenseFilter) ([]models.Expense, error) {
	conds := []string{"e.user_id = ?"}
	args := []any{f.UserID}

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
	      JOIN categories c ON c.id = e.category_id
	      WHERE ` + strings.Join(conds, " AND ") + `
	      ORDER BY e.date DESC, e.id DESC`

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

func (s *SQLiteStore) GetByID(id int64, userID int64) (*models.Expense, error) {
	var (
		e       models.Expense
		dateStr string
	)
	err := s.db.QueryRow(
		`SELECT e.id, e.amount, e.description, e.date, e.category_id, c.name
		 FROM expenses e
		 JOIN categories c ON c.id = e.category_id
		 WHERE e.id = ? AND e.user_id = ?`, id, userID,
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
		`UPDATE expenses
		 SET amount = ?, description = ?, date = ?, category_id = ?
		 WHERE id = ? AND user_id = ?`,
		e.Amount, e.Description, e.DateISO(), e.CategoryID, e.ID, e.UserID,
	)
	return err
}

func (s *SQLiteStore) Delete(id int64, userID int64) error {
	_, err := s.db.Exec(`DELETE FROM expenses WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

// ---------- Пользователи ----------

func (s *SQLiteStore) CreateUser(email, passwordHash string) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO users (email, password_hash) VALUES (?, ?)`,
		email, passwordHash,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *SQLiteStore) UserByEmail(email string) (*models.User, error) {
	var u models.User
	err := s.db.QueryRow(
		`SELECT id, email, password_hash, created_at FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ---------- Сессии ----------

func (s *SQLiteStore) CreateSession(sessionID string, userID int64, expiresAt time.Time) error {
	_, err := s.db.Exec(
		`INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)`,
		sessionID, userID, expiresAt.UTC().Format(time.RFC3339),
	)
	return err
}

// SessionUser возвращает пользователя по активной сессии.
// Если сессия истекла — удаляет её и возвращает nil, nil.
func (s *SQLiteStore) SessionUser(sessionID string) (*models.User, error) {
	var (
		u         models.User
		expiresAt string
	)
	err := s.db.QueryRow(
		`SELECT u.id, u.email, u.password_hash, u.created_at, s.expires_at
		 FROM sessions s
		 JOIN users u ON u.id = s.user_id
		 WHERE s.session_id = ?`, sessionID,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return nil, err
	}
	if time.Now().After(t) {
		_, _ = s.db.Exec(`DELETE FROM sessions WHERE session_id = ?`, sessionID)
		return nil, nil
	}
	return &u, nil
}

func (s *SQLiteStore) DeleteSession(sessionID string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE session_id = ?`, sessionID)
	return err
}