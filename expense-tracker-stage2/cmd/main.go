package main

import (
	"log"
	"net/http"

	"expense-tracker-stage2/internal/handlers"
	"expense-tracker-stage2/internal/storage"
)

func main() {
	store := storage.NewMemoryStore()

	h, err := handlers.NewHandler(store)
	if err != nil {
		log.Fatalf("не удалось загрузить шаблоны: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.List)
	mux.HandleFunc("GET /expenses/new", h.CreateForm)
	mux.HandleFunc("POST /expenses", h.Create)
	mux.Handle("GET /static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("web/static"))))

	addr := ":8080"
	log.Printf("сервер запущен на http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}