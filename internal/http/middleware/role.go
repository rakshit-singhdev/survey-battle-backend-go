package middleware

import "net/http"

// RequireRole mirrors authorize(role) from middleware/authorize.middleware.ts.
// It must run after AuthMiddleware has placed the authenticated User in the
// request context.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(User)

			if !ok || user.Role != role {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
