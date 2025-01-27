package middleware

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
)

var SessionManager *scs.SessionManager

func init() {
	SessionManager = scs.New()
	SessionManager.Cookie.Persist = true
	SessionManager.Cookie.SameSite = http.SameSiteLaxMode
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !SessionManager.Exists(r.Context(), "userID") {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
