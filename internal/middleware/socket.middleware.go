package middleware

import "net/http"

type SocketAuth struct {
	SessionID  string
	GameLiveID *string
	Role       string
	TeamId     *string
}

func SocketAuthMiddleWare(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sessionID := r.URL.Query().Get("sessionId")
		role := r.URL.Query().Get("role")
		teamId := r.URL.Query().Get("teamId")
		gameLiveId := r.URL.Query().Get("gameLiveId")

		if sessionID == "" || role == "" {
			http.Error(
				w,
				"Invalid socket auth",
				http.StatusUnauthorized,
			)
			return
		}

		auth := SocketAuth{
			SessionID:  sessionID,
			Role:       role,
		}

		if teamId != "" {
			auth.TeamId = &teamId
		}

		if gameLiveId != "" {
			auth.GameLiveID = &gameLiveId
		}

		next.ServeHTTP(w, r)
	})
}