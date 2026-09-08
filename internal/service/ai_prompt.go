package service

import (
	"fmt"
	"strings"
)

// BuildPromptInput mirrors BuildPromptInputDto from prompt/build.prompt.ts.
type BuildPromptInput struct {
	Topic              string
	NumberOfQuestions  *int
	AnswersPerQuestion *int
}

func responseSchema(withAnswers bool) string {
	if !withAnswers {
		return strings.TrimSpace(`
{
  "questions": [
    {
      "title": "string - survey-style question",
      "order": 0
    }
  ]
}`)
	}

	return strings.TrimSpace(`
{
  "questions": [
    {
      "title": "string - survey-style question",
      "order": 0,
      "answers": [
        {
          "answer": "string",
          "aliases": ["string", "string", "string"],
          "point": 0
        }
      ]
    }
  ]
}`)
}

func strictRules(numberOfQuestions, answersPerQuestion int) string {
	tail := strings.TrimSpace(fmt.Sprintf(`
7.  Do NOT include an "answers" field on any question.
8.  No duplicate questions.`))

	if answersPerQuestion > 0 {
		tail = strings.TrimSpace(fmt.Sprintf(`
7.  Each answer object must have exactly three keys: "answer", "aliases", "point".
8.  Every question MUST have EXACTLY %d answer(s).
9.  The "point" values within each question MUST sum to exactly 100.
10. Sort answers in descending order by "point".
11. Each answer MUST have at least 3 aliases (short alternate phrasings).
12. No duplicate questions, and no duplicate answers within a question.`, answersPerQuestion))
	}

	return strings.TrimSpace(fmt.Sprintf(`
STRICT RULES - violating any single rule makes the entire output invalid:
1.  Return ONLY a raw JSON object. No markdown, no code fences, no prose, no extra keys.
2.  Top-level object must have exactly ONE key: "questions" (an array). Do NOT add any game title.
3.  "questions" MUST contain EXACTLY %d question object(s).
4.  Each question object must have a "title" and a 1-based "order" (1, 2, 3 ... in sequence).
5.  Write questions in a classic 100-person audience-survey style:
    • Start each with "Name something...", "We asked 100 people...", or "What is something...".
    • Answers must read like real everyday survey responses, not dictionary definitions.
6.  Do NOT include any of these fields anywhere - they are injected at runtime:
    "gameId", "source", "isPublic", "embedding", "_id", "status".
%s`, numberOfQuestions, tail))
}

// BuildPrompt mirrors buildPrompt in prompt/build.prompt.ts exactly,
// including its null/undefined coercion rules.
func BuildPrompt(payload BuildPromptInput) string {
	numberOfQuestions := 1
	if payload.NumberOfQuestions != nil {
		numberOfQuestions = *payload.NumberOfQuestions
	}
	if numberOfQuestions < 1 {
		numberOfQuestions = 1
	}

	answersPerQuestion := 0
	if payload.AnswersPerQuestion != nil {
		answersPerQuestion = *payload.AnswersPerQuestion
	}

	withAnswers := answersPerQuestion > 0

	answersPerQuestionLabel := "0 (questions only, no answers)"
	if withAnswers {
		answersPerQuestionLabel = fmt.Sprintf("%d", answersPerQuestion)
	}

	example := strings.TrimSpace(`
{
  "questions": [
    { "title": "Name something people do after waking up.", "order": 1 }
  ]
}`)

	if withAnswers {
		example = strings.TrimSpace(`
{
  "questions": [
    {
      "title": "Name something people do after waking up.",
      "order": 1,
      "answers": [
        { "answer": "Brush teeth", "aliases": ["brush", "toothbrush", "brushing"], "point": 40 },
        { "answer": "Check phone", "aliases": ["mobile", "phone", "check mobile"], "point": 35 },
        { "answer": "Drink water", "aliases": ["water", "glass of water", "hydrate"], "point": 25 }
      ]
    }
  ]
}`)
	}

	answersInstruction := ` without any "answers" field.`
	if withAnswers {
		answersInstruction = fmt.Sprintf(" with EXACTLY %d answer(s) per question.", answersPerQuestion)
	}

	return strings.TrimSpace(fmt.Sprintf(`
You are an expert survey-game question writer. Your ONLY output is a valid JSON object containing survey questions.

TASK:
- Topic                : "%s"
- Total questions      : %d
- Answers per question : %s

%s

REQUIRED OUTPUT SCHEMA (follow exactly - no extra fields):
%s

EXAMPLE (structure reference only - generate your own, do NOT copy this):
%s

Now generate EXACTLY %d question(s) on the topic "%s"%s
`,
		payload.Topic,
		numberOfQuestions,
		answersPerQuestionLabel,
		strictRules(numberOfQuestions, answersPerQuestion),
		responseSchema(withAnswers),
		example,
		numberOfQuestions,
		payload.Topic,
		answersInstruction,
	))
}
