package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	sessionCookieName = "kai_session"
	sessionDuration   = 12 * time.Hour
)

var (
	sessionStore = map[string]time.Time{}
	sessionMu    sync.Mutex
)

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func isValidSession(token string) bool {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	expiry, ok := sessionStore[token]
	if !ok {
		return false
	}

	if time.Now().After(expiry) {
		delete(sessionStore, token)
		return false
	}

	return true
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || !isValidSession(cookie.Value) {
			http.Redirect(w, r, "/login?next="+r.URL.Path, http.StatusFound)
			return
		}
		next(w, r)
	}
}

type LoginHandler struct{}

func NewLoginHandler() *LoginHandler { return &LoginHandler{} }

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.handlePost(w, r)
		return
	}
	http.ServeFile(w, r, "static/login.html")
}

func (h *LoginHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin"
	}

	if password != adminPassword {
		http.Redirect(w, r, "/login?error=1", http.StatusFound)
		return
	}

	token, err := generateToken()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	sessionMu.Lock()
	sessionStore[token] = time.Now().Add(sessionDuration)
	sessionMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionDuration.Seconds()),
	})

	next := r.URL.Query().Get("next")
	if next == "" || next == "/login" {
		next = "/requests"
	}

	http.Redirect(w, r, next, http.StatusFound)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		sessionMu.Lock()
		delete(sessionStore, cookie.Value)
		sessionMu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/login", http.StatusFound)
}
