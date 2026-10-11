package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"expense-tracker-stage2/internal/models"
	"expense-tracker-stage2/internal/storage"
)

type Handler struct {
	store *storage.MemoryStore
	tmpl  *template.Template
}

func NewHandler(store *storage.MemoryStore) (*Handler, error) {
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Handler{store: store, tmpl: tmpl}, nil
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":    "Список трат",
		"Expenses": h.store.All(),
		"Total":    h.totalAmount(),
	}
	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("render list: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) CreateForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title": "Новая трата",
	}
	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("render form: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректная форма", http.StatusBadRequest)
		return
	}

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "сумма должна быть положительным числом", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", r.FormValue("date"))
	if err != nil {
		http.Error(w, "некорректная дата", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "описание не может быть пустым", http.StatusBadRequest)
		return
	}

	h.store.Add(models.Expense{
		Amount:      amount,
		Description: description,
		Date:        date,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) totalAmount() float64 {
	var sum float64
	for _, e := range h.store.All() {
		sum += e.Amount
	}
	return sum
}