package handlers

import (
	"encoding/json"
	"net/http"

	"survey-battle-backend-go/internal/apperror"
	"survey-battle-backend-go/internal/config"
	"survey-battle-backend-go/internal/http/middleware"
	"survey-battle-backend-go/internal/response"
	"survey-battle-backend-go/internal/service"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// GenerateQuestionsRequest mirrors BuildPromptInputDto from
// prompt/build.prompt.ts, decoded from the POST /ai/generate-questions body.
type GenerateQuestionsRequest struct {
	Topic              string `json:"topic"`
	NumberOfQuestions  *int   `json:"numberOfQuestions"`
	AnswersPerQuestion *int   `json:"answersPerQuestion"`
}

func (r GenerateQuestionsRequest) ToBuildPromptInput() service.BuildPromptInput {
	return service.BuildPromptInput{
		Topic:              r.Topic,
		NumberOfQuestions:  r.NumberOfQuestions,
		AnswersPerQuestion: r.AnswersPerQuestion,
	}
}

// AIHandler mirrors controller.ts's `controller` object for the /ai routes.
type AIHandler struct {
	service *service.AIService
}

func NewAIHandler(svc *service.AIService) *AIHandler {
	return &AIHandler{service: svc}
}

// RegisterAIRoutes mirrors routes.ts's aiRouter: both routes require an
// access token and the host role (authorize(USER_ROLE.host) in TS).
func RegisterAIRoutes(mux *http.ServeMux, cfg *config.Config, handler *AIHandler) {
	authorized := func(next http.HandlerFunc) http.Handler {
		return middleware.AuthMiddleware(cfg)(
			middleware.RequireRole("host")(http.HandlerFunc(next)),
		)
	}

	mux.Handle("POST /ai/generate-questions", authorized(handler.GenerateQuestionsByAI))
	mux.Handle("GET /ai/analysis/survey/response/{id}", authorized(handler.AnalyzeSurveyResponseByAI))
}

// GenerateQuestionsByAI mirrors controller.generateQuestionsByAI.
func (h *AIHandler) GenerateQuestionsByAI(w http.ResponseWriter, r *http.Request) {
	var req GenerateQuestionsRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, "Invalid request body", http.StatusBadRequest, "", nil)
		return
	}

	result, err := h.service.GenerateQuestions(r.Context(), req.ToBuildPromptInput())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	response.Success(w, result, "Questions generated successfully", http.StatusOK)
}

// AnalyzeSurveyResponseByAI mirrors controller.analyzeSurveyResponseByAI.
// The "Questions generated successfully" message matches the TS controller
// verbatim, including its apparent copy-paste from the sibling handler.
func (h *AIHandler) AnalyzeSurveyResponseByAI(w http.ResponseWriter, r *http.Request) {
	surveyID, err := bson.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		response.Error(w, "Invalid survey id", http.StatusBadRequest, "", nil)
		return
	}

	result, err := h.service.AnalyzeSurveyResponses(r.Context(), surveyID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	response.Success(w, result, "Questions generated successfully", http.StatusOK)
}

func writeServiceError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*apperror.AppError); ok {
		response.Error(w, appErr.Message, appErr.StatusCode, appErr.Code, nil)
		return
	}

	response.Error(w, err.Error(), http.StatusInternalServerError, "", nil)
}
