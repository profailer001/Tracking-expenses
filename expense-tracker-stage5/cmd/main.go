package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	"expense-tracker-stage5/internal/auth"
	"expense-tracker-stage5/internal/handlers"
	"expense-tracker-stage5/internal/storage"
)

func main() {
	// DSN с PRAGMA foreign_keys на уровне каждого соединения.
	dsn := "file:expenses.db?_pragma=foreign_keys(1)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("открытие БД: %v", err)
	}
	defer db.Close()

	// modernc.org/sqlite + PRAGMA-параметры: одно соединение надёжнее.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		log.Fatalf("ping БД: %v", err)
	}

	migrations := []string{
		"migrations/001_init.sql",
		"migrations/002_users.sql",
	}
	for _, path := range migrations {
		mig, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("чтение миграции %s: %v", path, err)
		}
		if _, err := db.Exec(string(mig)); err != nil {
			log.Fatalf("применение миграции %s: %v", path, err)
		}
	}

	// ALTER TABLE не поддерживает IF NOT EXISTS — проверяем колонку вручную.
	if err := ensureColumn(db, "expenses", "user_id",
		`ALTER TABLE expenses ADD COLUMN user_id INTEGER REFERENCES users(id)`); err != nil {
		log.Fatalf("миграция user_id: %v", err)
	}

	store := storage.NewSQLiteStore(db)

	h, err := handlers.NewHandler(store)
	if err != nil {
		log.Fatalf("не удалось загрузить шаблоны: %v", err)
	}

	requireAuth := auth.RequireAuth(store)

	mux := http.NewServeMux()

	// Публичные маршруты.
	mux.HandleFunc("GET /register", h.RegisterForm)
	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("GET /login", h.LoginForm)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /logout", h.Logout)

	// Защищённые маршруты.
	mux.Handle("GET /{$}", requireAuth(http.HandlerFunc(h.List)))
	mux.Handle("GET /expenses/new", requireAuth(http.HandlerFunc(h.CreateForm)))
	mux.Handle("POST /expenses", requireAuth(http.HandlerFunc(h.Create)))
	mux.Handle("GET /expenses/{id}/edit", requireAuth(http.HandlerFunc(h.EditForm)))
	mux.Handle("POST /expenses/{id}", requireAuth(http.HandlerFunc(h.Update)))
	mux.Handle("GET /expenses/{id}/delete", requireAuth(http.HandlerFunc(h.DeleteConfirm)))
	mux.Handle("POST /expenses/{id}/delete", requireAuth(http.HandlerFunc(h.Delete)))

	// Статика.
	mux.Handle("GET /static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("web/static"))))

	addr := ":8080"
	log.Printf("сервер запущен на http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// ensureColumn добавляет колонку, если её ещё нет в таблице.
func ensureColumn(db *sql.DB, table, column, alterSQL string) error {
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid       int
			name      string
			ctype     string
			notnull   int
			dfltValue sql.NullString
			pk        int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = db.Exec(alterSQL)
	return err
}