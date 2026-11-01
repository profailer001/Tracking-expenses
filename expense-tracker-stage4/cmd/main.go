package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	"expense-tracker-stage3/internal/handlers"
	"expense-tracker-stage3/internal/storage"
)

func main() {
	db, err := sql.Open("sqlite", "expenses.db")
	if err != nil {
		log.Fatalf("открытие БД: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping БД: %v", err)
	}

	mig, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		log.Fatalf("чтение миграции: %v", err)
	}
	if _, err := db.Exec(string(mig)); err != nil {
		log.Fatalf("применение миграции: %v", err)
	}

	store := storage.NewSQLiteStore(db)

	h, err := handlers.NewHandler(store)
	if err != nil {
		log.Fatalf("не удалось загрузить шаблоны: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.List)
	mux.HandleFunc("GET /expenses/new", h.CreateForm)
	mux.HandleFunc("POST /expenses", h.Create)
	mux.HandleFunc("GET /expenses/{id}/edit", h.EditForm)
	mux.HandleFunc("POST /expenses/{id}", h.Update)
	mux.HandleFunc("GET /expenses/{id}/delete", h.DeleteConfirm)
	mux.HandleFunc("POST /expenses/{id}/delete", h.Delete)

	mux.Handle("GET /static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("web/static"))))

	addr := ":8080"
	log.Printf("сервер запущен на http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}