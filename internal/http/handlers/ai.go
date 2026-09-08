package handlers

import (
	"encoding/json"
	"net/http"

	"survey-battle-backend-go/internal/apperror"
	"survey-battle-backend-go/internal/http/requests"
	"survey-battle-backend-go/internal/response"
	"survey-battle-backend-go/internal/service"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// AIHandler mirrors controller.ts's `controller` object for the /ai routes.
type AIHandler struct {
	service *service.AIService
}

func NewAIHandler(svc *service.AIService) *AIHandler {
	return &AIHandler{service: svc}
}

// GenerateQuestionsByAI mirrors controller.generateQuestionsByAI.
func (h *AIHandler) GenerateQuestionsByAI(w http.ResponseWriter, r *http.Request) {
	var req requests.GenerateQuestionsRequest

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
