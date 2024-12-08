package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/KonstantinGalanin/redditclone/internal/session"
)

func checkSession(r *http.Request, sm session.SessionManager) (*session.Session, error) {
	cookieSessionID, err := r.Cookie("session_id")
	if errors.Is(err, http.ErrNoCookie) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	sess, err := sm.Check(&session.SessionID{
		ID: cookieSessionID.Value,
	})
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func refreshSession(w http.ResponseWriter, sess *session.Session, sm session.SessionManager) {
	sessID, err := sm.Create(sess)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:    "session_id",
		Value:   sessID.ID,
		Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, &cookie)
}

func Session(sm session.SessionManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, err := checkSession(r, sm)
			if err != nil && sess != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
			}
			refreshSession(w, sess, sm)
			ctx := context.WithValue(r.Context(), "session", sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
