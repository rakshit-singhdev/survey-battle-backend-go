package service

import "sort"

// AnalysisAnswer mirrors the per-answer object built by surveyAnalysisParser
// in parser/survey.analysis.parser.ts.
type AnalysisAnswer struct {
	Answer  string   `json:"answer"`
	Aliases []string `json:"aliases"`
	Point   float64  `json:"point"`
}

// AnalysisQuestion mirrors the per-question object built by
// surveyAnalysisParser in parser/survey.analysis.parser.ts.
type AnalysisQuestion struct {
	QuestionID string           `json:"questionId"`
	Title      string           `json:"title"`
	Answers    []AnalysisAnswer `json:"answers"`
}

func stringOrEmpty(v any) string {
	s, _ := v.(string)
	return s
}

func rawAliases(v any) []string {
	raw, ok := v.([]any)
	if !ok {
		return []string{}
	}

	aliases := make([]string, 0, len(raw))
	for _, item := range raw {
		aliases = append(aliases, stringOrEmpty(item))
	}

	return aliases
}

func rawPoint(v any) float64 {
	n, _ := toFloat(v)
	return n
}

func ParseSurveyAnalysis(data map[string]any) []AnalysisQuestion {
	rawQuestions, _ := data["questionStats"].([]any)

	questions := make([]AnalysisQuestion, 0, len(rawQuestions))

	for _, item := range rawQuestions {
		q, _ := item.(map[string]any)

		rawAnswers, _ := q["answers"].([]any)
		answers := make([]AnalysisAnswer, 0, len(rawAnswers))

		for _, answerItem := range rawAnswers {
			a, _ := answerItem.(map[string]any)

			answers = append(answers, AnalysisAnswer{
				Answer:  stringOrEmpty(a["main_answer"]),
				Aliases: rawAliases(a["aliases"]),
				Point:   rawPoint(a["point"]),
			})
		}

		sort.SliceStable(answers, func(i, j int) bool {
			return answers[i].Point > answers[j].Point
		})

		questions = append(questions, AnalysisQuestion{
			QuestionID: stringOrEmpty(q["questionId"]),
			Title:      stringOrEmpty(q["title"]),
			Answers:    answers,
		})
	}

	return questions
}
