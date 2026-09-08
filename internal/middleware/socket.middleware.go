package middleware

import (
	"context"
	"net/http"
)

type SocketAuth struct {
	SessionID  string
	GameLiveID *string
	Role       string
	TeamID     *string
}

const SocketAuthContextKey contextKey = "socketAuth"

func SocketAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sessionID := r.URL.Query().Get("sessionId")
		role := r.URL.Query().Get("role")
		teamID := r.URL.Query().Get("teamId")
		gameLiveID := r.URL.Query().Get("gameLiveId")

		if sessionID == "" || role == "" {
			http.Error(
				w,
				"Invalid socket auth",
				http.StatusUnauthorized,
			)
			return
		}

		auth := SocketAuth{
			SessionID: sessionID,
			Role:      role,
		}

		if teamID != "" {
			auth.TeamID = &teamID
		}

		if gameLiveID != "" {
			auth.GameLiveID = &gameLiveID
		}

		ctx := context.WithValue(
			r.Context(),
			SocketAuthContextKey,
			auth,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}