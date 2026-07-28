package handler

import (
	"crypto/subtle"
	"html/template"
	"net/http"

	"adms-go/internal/middleware"
)

type AuthHandler struct {
	User     string
	Pass     string
	Template *template.Template
}

// LoginPage handles GET /login
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Already logged in? Redirect to devices
	if cookie, err := r.Cookie("adms_session"); err == nil {
		if middleware.ValidSession(cookie.Value, h.User, h.Pass) {
			http.Redirect(w, r, "/devices", http.StatusFound)
			return
		}
	}

	data := map[string]interface{}{
		"Error": r.URL.Query().Get("error"),
	}
	h.Template.ExecuteTemplate(w, "login.html", data)
}

// Login handles POST /login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	user := r.FormValue("username")
	pass := r.FormValue("password")

	if subtle.ConstantTimeCompare([]byte(user), []byte(h.User)) == 1 &&
		subtle.ConstantTimeCompare([]byte(pass), []byte(h.Pass)) == 1 {
		middleware.SetSessionCookie(w, h.User, h.Pass)
		http.Redirect(w, r, "/devices", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/login?error=invalid", http.StatusFound)
}

// Logout handles GET /logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "adms_session", MaxAge: -1, Path: "/"})
	http.Redirect(w, r, "/login", http.StatusFound)
}
