package routes

import (
	"net/http"

	"survey-battle-backend-go/internal/config"
	"survey-battle-backend-go/internal/http/handlers"
	"survey-battle-backend-go/internal/http/middleware"
)

// RegisterAIRoutes mirrors routes.ts's aiRouter: both routes require an
// access token and the host role (authorize(USER_ROLE.host) in TS).
func RegisterAIRoutes(mux *http.ServeMux, cfg *config.Config, handler *handlers.AIHandler) {
	authorized := func(next http.HandlerFunc) http.Handler {
		return middleware.AuthMiddleware(cfg)(
			middleware.RequireRole("host")(http.HandlerFunc(next)),
		)
	}

	mux.Handle("POST /ai/generate-questions", authorized(handler.GenerateQuestionsByAI))
	mux.Handle("GET /ai/analysis/survey/response/{id}", authorized(handler.AnalyzeSurveyResponseByAI))
}
