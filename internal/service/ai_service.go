package service

import (
	"context"
	"encoding/json"
	"net/http"

	"survey-battle-backend-go/internal/apperror"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SurveyResponseProvider is the seam for survey-response data until the
// survey module is ported to Go; it stands in for
// serviceProvider.survey.getSurveyResponse from service.ts.
type SurveyResponseProvider interface {
	GetSurveyResponse(ctx context.Context, surveyID bson.ObjectID) ([]RawQuestionStat, error)
}

// AIService mirrors Service from service.ts.
type AIService struct {
	llm     LLM
	surveys SurveyResponseProvider
}

func NewAIService(llm LLM, surveys SurveyResponseProvider) *AIService {
	return &AIService{llm: llm, surveys: surveys}
}

// GenerateQuestions mirrors Service.generateQuestionsByAI in service.ts.
func (s *AIService) GenerateQuestions(ctx context.Context, input BuildPromptInput) (ParsedAIBuildResponse, error) {
	prompt := BuildPrompt(input)

	content, err := s.llm.Invoke(ctx, []LLMMessage{
		{Role: "system", Content: "You are a game data generator. Output only valid JSON."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return ParsedAIBuildResponse{}, err
	}

	if content == "" {
		return ParsedAIBuildResponse{}, apperror.New("AI returned empty response", http.StatusBadRequest)
	}

	var parsed any
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return ParsedAIBuildResponse{}, apperror.New("AI response parsing failed", http.StatusBadRequest)
	}

	result := ParseAIBuildResponse(parsed)

	if len(result.Questions) == 0 {
		return ParsedAIBuildResponse{}, apperror.New("AI returned no valid questions", http.StatusBadRequest)
	}

	return result, nil
}

// AnalyzeSurveyResponses mirrors Service.analyzeSurveyResponsesWithAI in
// service.ts.
func (s *AIService) AnalyzeSurveyResponses(ctx context.Context, surveyID bson.ObjectID) ([]AnalysisQuestion, error) {
	stats, err := s.surveys.GetSurveyResponse(ctx, surveyID)
	if err != nil {
		return nil, err
	}

	prompt := BuildSurveyAnalysisPrompt(stats, DefaultAnalysisTopN)

	content, err := s.llm.Invoke(ctx, []LLMMessage{
		{Role: "system", Content: "You are a survey-data analyst. You group raw answers by intent and return only valid JSON."},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, err
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, apperror.New("AI response parsing failed", http.StatusBadRequest)
	}

	return ParseSurveyAnalysis(parsed), nil
}
