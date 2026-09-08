package service

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// AIQuestionSource mirrors QUESTION_SOURCE.ai from utils/common/enum.ts.
const AIQuestionSource = "ai"

// ParsedAnswer mirrors ParsedAnswer from parser/build.parser.ts.
type ParsedAnswer struct {
	Answer  string   `json:"answer"`
	Aliases []string `json:"aliases"`
	Point   float64  `json:"point"`
}

// ParsedQuestion mirrors ParsedQuestion from parser/build.parser.ts.
type ParsedQuestion struct {
	Title   string         `json:"title"`
	Order   int            `json:"order"`
	Source  string         `json:"source"`
	Answers []ParsedAnswer `json:"answers"`
}

// ParsedAIBuildResponse mirrors ParsedAiResponse from parser/build.parser.ts.
type ParsedAIBuildResponse struct {
	Questions []ParsedQuestion `json:"questions"`
}

func cleanStr(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(s)
}

func toPoint(v any) float64 {
	n, ok := toFloat(v)
	if !ok || math.IsNaN(n) || math.IsInf(n, 0) || n < 0 {
		return 0
	}

	return n
}

func toFloat(v any) (float64, bool) {
	switch value := v.(type) {
	case float64:
		return value, true
	case int:
		return float64(value), true
	case string:
		n, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return 0, false
		}

		return n, true
	default:
		return 0, false
	}
}

// cleanAliases trims, dedupes, and drops empty values, preserving first-seen
// order the same way `new Set(...)` does in cleanAliases (parser/build.parser.ts).
func cleanAliases(v any) []string {
	raw, ok := v.([]any)
	if !ok {
		return []string{}
	}

	seen := make(map[string]struct{}, len(raw))
	result := make([]string, 0, len(raw))

	for _, item := range raw {
		s := cleanStr(item)
		if s == "" {
			continue
		}

		if _, exists := seen[s]; exists {
			continue
		}

		seen[s] = struct{}{}
		result = append(result, s)
	}

	return result
}

func parseAnswers(raw any) []ParsedAnswer {
	items, ok := raw.([]any)
	if !ok {
		return []ParsedAnswer{}
	}

	answers := make([]ParsedAnswer, 0, len(items))

	for _, item := range items {
		a, _ := item.(map[string]any)

		answer := ParsedAnswer{
			Answer:  cleanStr(a["answer"]),
			Aliases: cleanAliases(a["aliases"]),
			Point:   toPoint(a["point"]),
		}

		if answer.Answer == "" {
			continue
		}

		answers = append(answers, answer)
	}

	sort.SliceStable(answers, func(i, j int) bool {
		return answers[i].Point > answers[j].Point
	})

	return answers
}

// ParseAIBuildResponse mirrors aiBuildResponseParser from
// parser/build.parser.ts.
func ParseAIBuildResponse(data any) ParsedAIBuildResponse {
	root, _ := data.(map[string]any)

	rawQuestions, _ := root["questions"].([]any)

	questions := make([]ParsedQuestion, 0, len(rawQuestions))
	order := 1

	for _, item := range rawQuestions {
		q, _ := item.(map[string]any)

		title := cleanStr(q["title"])
		if title == "" {
			continue
		}

		questions = append(questions, ParsedQuestion{
			Title:   title,
			Order:   order,
			Source:  AIQuestionSource,
			Answers: parseAnswers(q["answers"]),
		})
		order++
	}

	return ParsedAIBuildResponse{Questions: questions}
}
