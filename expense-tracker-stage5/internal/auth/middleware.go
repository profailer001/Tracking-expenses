package auth

import (
	"log"
	"net/http"

	"expense-tracker-stage5/internal/storage"
)

const CookieName = "session_id"

// RequireAuth пропускает запрос дальше, только если есть валидная сессия.
// ID пользователя кладётся в контекст запроса.
func RequireAuth(store *storage.SQLiteStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(CookieName)
			if err != nil || c.Value == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			u, err := store.SessionUser(c.Value)
			if err != nil {
				log.Printf("session lookup: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if u == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), u.ID)))
		})
	}
}