package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"expense-tracker-stage3/internal/models"
	"expense-tracker-stage3/internal/storage"
)

type Handler struct {
	store *storage.SQLiteStore
	tmpl  *template.Template
}

func NewHandler(store *storage.SQLiteStore) (*Handler, error) {
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Handler{store: store, tmpl: tmpl}, nil
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	expenses, err := h.store.All()
	if err != nil {
		log.Printf("list: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var total float64
	for _, e := range expenses {
		total += e.Amount
	}

	data := map[string]any{
		"Title":    "Список трат",
		"Expenses": expenses,
		"Total":    total,
	}
	h.render(w, data)
}

func (h *Handler) CreateForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":     "Новая трата",
		"IsEdit":    false,
		"Action":    "/expenses",
		"Expense":   models.Expense{Date: time.Now()},
	}
	h.render(w, data)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	e, ok := h.parseExpenseForm(w, r)
	if !ok {
		return
	}
	if _, err := h.store.Create(e); err != nil {
		log.Printf("create: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) EditForm(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	e, err := h.store.GetByID(id)
	if err != nil {
		log.Printf("get by id: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if e == nil {
		http.NotFound(w, r)
		return
	}
	data := map[string]any{
		"Title":   "Редактировать трату",
		"IsEdit":  true,
		"Action":  "/expenses/" + strconv.FormatInt(id, 10),
		"Expense": *e,
	}
	h.render(w, data)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	e, ok := h.parseExpenseForm(w, r)
	if !ok {
		return
	}
	e.ID = id
	if err := h.store.Update(e); err != nil {
		log.Printf("update: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) DeleteConfirm(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	e, err := h.store.GetByID(id)
	if err != nil {
		log.Printf("get by id: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if e == nil {
		http.NotFound(w, r)
		return
	}
	data := map[string]any{
		"Title":   "Удаление траты",
		"Expense": *e,
	}
	h.render(w, data)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if err := h.store.Delete(id); err != nil {
		log.Printf("delete: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) render(w http.ResponseWriter, data map[string]any) {
	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("render: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func parseID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (h *Handler) parseExpenseForm(w http.ResponseWriter, r *http.Request) (models.Expense, bool) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректная форма", http.StatusBadRequest)
		return models.Expense{}, false
	}

	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "сумма должна быть положительным числом", http.StatusBadRequest)
		return models.Expense{}, false
	}

	date, err := time.Parse("2006-01-02", r.FormValue("date"))
	if err != nil {
		http.Error(w, "некорректная дата", http.StatusBadRequest)
		return models.Expense{}, false
	}

	description := r.FormValue("description")
	if description == "" {
		http.Error(w, "описание не может быть пустым", http.StatusBadRequest)
		return models.Expense{}, false
	}

	return models.Expense{
		Amount:      amount,
		Description: description,
		Date:        date,
	}, true
}