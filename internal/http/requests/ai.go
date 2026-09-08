package requests

import "survey-battle-backend-go/internal/service"

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
