package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
)

// BasicAuth returns middleware that checks session cookie against env credentials
func BasicAuth(user, pass string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("adms_session")
			if err != nil || !validSession(cookie.Value, user, pass) {
				// Clear bad cookie
				http.SetCookie(w, &http.Cookie{Name: "adms_session", MaxAge: -1})
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ValidSession(token, user, pass string) bool {
	expected := sessionToken(user, pass)
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}

func validSession(token, user, pass string) bool {
	return ValidSession(token, user, pass)
}

func sessionToken(user, pass string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", user, pass)))
	return fmt.Sprintf("%x", h)
}

func SetSessionCookie(w http.ResponseWriter, user, pass string) {
	token := sessionToken(user, pass)
	http.SetCookie(w, &http.Cookie{
		Name:     "adms_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})
}
