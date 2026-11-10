package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"expense-tracker-stage5/internal/auth"
	"expense-tracker-stage5/internal/models"
	"expense-tracker-stage5/internal/storage"
)

const sessionDuration = 7 * 24 * time.Hour

type Handler struct {
	store    *storage.SQLiteStore
	listTmpl *template.Template
	formTmpl *template.Template
	authTmpl *template.Template
	delTmpl  *template.Template
}

func NewHandler(store *storage.SQLiteStore) (*Handler, error) {
	listTmpl, err := template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/index.html",
	)
	if err != nil {
		return nil, err
	}
	formTmpl, err := template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/form.html",
	)
	if err != nil {
		return nil, err
	}
	authTmpl, err := template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/auth.html",
	)
	if err != nil {
		return nil, err
	}
	delTmpl, err := template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/delete.html",
	)
	if err != nil {
		return nil, err
	}
	return &Handler{
		store:    store,
		listTmpl: listTmpl,
		formTmpl: formTmpl,
		authTmpl: authTmpl,
		delTmpl:  delTmpl,
	}, nil
}

// ---------- Список и CRUD трат ----------

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	cats, err := h.store.Categories()
	if err != nil {
		log.Printf("categories: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	filter := models.ExpenseFilter{UserID: userID}

	if v := q.Get("category"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.CategoryID = id
		}
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.From = t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.To = t
		}
	}

	expenses, err := h.store.All(filter)
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
		"Title":         "Список трат",
		"Expenses":      expenses,
		"Categories":    cats,
		"Total":         total,
		"Filter":        filter,
		"FilterCatStr":  q.Get("category"),
		"FilterFromStr": q.Get("from"),
		"FilterToStr":   q.Get("to"),
		"Auth":          true,
	}
	h.render(w, h.listTmpl, data)
}

func (h *Handler) CreateForm(w http.ResponseWriter, r *http.Request) {
	cats, err := h.store.Categories()
	if err != nil {
		log.Printf("categories: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"Title":      "Новая трата",
		"Action":     "/expenses",
		"Expense":    models.Expense{Date: time.Now()},
		"Categories": cats,
		"Auth":       true,
	}
	h.render(w, h.formTmpl, data)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	e, ok := h.parseExpenseForm(w, r)
	if !ok {
		return
	}
	e.UserID = userID
	if _, err := h.store.Create(e); err != nil {
		log.Printf("create: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) EditForm(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	e, err := h.store.GetByID(id, userID)
	if err != nil {
		log.Printf("get by id: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if e == nil {
		http.NotFound(w, r)
		return
	}
	cats, err := h.store.Categories()
	if err != nil {
		log.Printf("categories: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"Title":      "Редактировать трату",
		"Action":     "/expenses/" + strconv.FormatInt(id, 10),
		"Expense":    *e,
		"Categories": cats,
		"Auth":       true,
	}
	h.render(w, h.formTmpl, data)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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
	e.UserID = userID
	if err := h.store.Update(e); err != nil {
		log.Printf("update: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) DeleteConfirm(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	e, err := h.store.GetByID(id, userID)
	if err != nil {
		log.Printf("get by id: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if e == nil {
		http.NotFound(w, r)
		return
	}
	data := map[string]any{"Title": "Удаление траты", "Expense": *e, "Auth": true}
	h.render(w, h.delTmpl, data)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if err := h.store.Delete(id, userID); err != nil {
		log.Printf("delete: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ---------- Регистрация / вход / выход ----------

func (h *Handler) RegisterForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":  "Регистрация",
		"Action": "/register",
		"Mode":   "register",
		"Auth":   false,
	}
	h.render(w, h.authTmpl, data)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректная форма", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")

	renderErr := func(msg string) {
		data := map[string]any{
			"Title":  "Регистрация",
			"Action": "/register",
			"Mode":   "register",
			"Error":  msg,
			"Email":  email,
			"Auth":   false,
		}
		w.WriteHeader(http.StatusBadRequest)
		h.render(w, h.authTmpl, data)
	}

	if email == "" || password == "" {
		renderErr("Заполните все поля")
		return
	}
	if len(password) < 6 {
		renderErr("Пароль должен содержать минимум 6 символов")
		return
	}

	existing, err := h.store.UserByEmail(email)
	if err != nil {
		log.Printf("user by email: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if existing != nil {
		renderErr("Пользователь с таким email уже существует")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	userID, err := h.store.CreateUser(email, string(hash))
	if err != nil {
		log.Printf("create user: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := h.startSession(w, userID); err != nil {
		log.Printf("session: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":  "Вход",
		"Action": "/login",
		"Mode":   "login",
		"Auth":   false,
	}
	h.render(w, h.authTmpl, data)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректная форма", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")

	renderErr := func(msg string) {
		data := map[string]any{
			"Title":  "Вход",
			"Action": "/login",
			"Mode":   "login",
			"Error":  msg,
			"Email":  email,
			"Auth":   false,
		}
		w.WriteHeader(http.StatusUnauthorized)
		h.render(w, h.authTmpl, data)
	}

	u, err := h.store.UserByEmail(email)
	if err != nil {
		log.Printf("user by email: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if u == nil {
		renderErr("Неверный email или пароль")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		renderErr("Неверный email или пароль")
		return
	}

	if err := h.startSession(w, u.ID); err != nil {
		log.Printf("session: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(auth.CookieName); err == nil && c.Value != "" {
		if err := h.store.DeleteSession(c.Value); err != nil {
			log.Printf("delete session: %v", err)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ---------- Вспомогательное ----------

func (h *Handler) startSession(w http.ResponseWriter, userID int64) error {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return err
	}
	sid := hex.EncodeToString(buf)
	expires := time.Now().Add(sessionDuration)
	if err := h.store.CreateSession(sid, userID, expires); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName,
		Value:    sid,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (h *Handler) render(w http.ResponseWriter, tmpl *template.Template, data map[string]any) {
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("render: %v", err)
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

	categoryID, err := strconv.ParseInt(r.FormValue("category_id"), 10, 64)
	if err != nil || categoryID <= 0 {
		http.Error(w, "выберите категорию", http.StatusBadRequest)
		return models.Expense{}, false
	}

	return models.Expense{
		Amount:      amount,
		Description: description,
		Date:        date,
		CategoryID:  categoryID,
	}, true
}