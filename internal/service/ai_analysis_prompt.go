package service

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RawQuestionStat mirrors RawQuestionStat from prompt/analysis.prompt.ts.
type RawQuestionStat struct {
	QuestionID string   `json:"questionId"`
	Title      string   `json:"title"`
	Answers    []string `json:"answers"`
}

// DefaultAnalysisTopN mirrors the topN = 5 default parameter on
// buildSurveyAnalysisPrompt in prompt/analysis.prompt.ts.
const DefaultAnalysisTopN = 5

func analysisResponseSchema() string {
	return strings.TrimSpace(`
{
  "questionStats": [
    {
      "questionId": "string - identical to input",
      "title": "string - identical to input",
      "answers": [
        {
          "main_answer": "string",
          "aliases": ["string"],
          "point": "integer > 0 (sum MUST be 100)"
        }
      ]
    }
  ]
}`)
}

func analysisStrictRules(topN int) string {
	return strings.TrimSpace(fmt.Sprintf(`
════════════════════════════════════════════════════════════════
STRICT RULES - NO VIOLATION ALLOWED
════════════════════════════════════════════════════════════════

[FORMAT]
R1. Return ONLY raw JSON (no markdown, no text)
R2. Root must contain ONLY "questionStats"

[COUNT]
R3. EXACTLY %d answers per question (NO MORE, NO LESS)

[COPY]
R4. questionId & title must be EXACT same as input

[GROUPING]
R5. Group by MEANING (not text)
    Examples:
    - "colddrik", "pespi" → Cold Drink
    - "burger", "burgur" → Burger
    - "momo", "veg momo" → Momos

[RANKING]
R6. Sort groups by frequency DESC
R7. Keep ONLY top %d

[PADDING]
R8. If < %d groups → generate realistic answers
R9. Generated answers MUST have empty aliases []

[ALIASES]
R10. Include ALL original values in aliases (no cleaning)

════════════════════════════════════════════════════════════════
`, topN, topN, topN))
}

func analysisPointRules() string {
	return strings.TrimSpace(`
════════════════════════════════════════════════════════════════
POINT RULES - PRODUCTION LEVEL SCORING
════════════════════════════════════════════════════════════════

R11. Total points MUST be EXACTLY 100

R12. Points MUST reflect popularity:
     more users → more points

R13. NO duplicate points (as much as possible)
     ❌ 20,20,20,20
     ✅ 35,25,20,12,8

R14. If duplicate happens:
     → adjust ±1 while keeping total = 100

R15. All points MUST be > 0

R16. Strict descending order:
     highest point first

R17. If same frequency:
     → assign slightly different points (21 vs 20)

R18. Generated answers:
     → MUST have lowest points
     → always less than smallest real answer

R19. After all calculations:
     → fix total to 100 by adjusting largest group

R20. NEVER assign random points

R21. NO duplicate main_answer allowed

════════════════════════════════════════════════════════════════
`)
}

func analysisFewShotExamples() string {
	return strings.TrimSpace(`
EXAMPLE:

Input:
pizza, pizza, burger, momo

Output:
Pizza 40
Burger 30
Momos 20
Sandwich 10

(Sum = 100, all unique, sorted)
`)
}

// BuildSurveyAnalysisPrompt mirrors buildSurveyAnalysisPrompt in
// prompt/analysis.prompt.ts, including its JSON.stringify(_, null, 2)
// formatting of the input block.
func BuildSurveyAnalysisPrompt(questionStats []RawQuestionStat, topN int) string {
	if questionStats == nil {
		questionStats = []RawQuestionStat{}
	}

	input, _ := json.MarshalIndent(struct {
		QuestionStats []RawQuestionStat `json:"questionStats"`
	}{QuestionStats: questionStats}, "", "  ")

	return strings.TrimSpace(fmt.Sprintf(`
You are a senior survey analyst.

Your job:
Convert raw messy answers into clean grouped answers with scoring.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
INPUT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
%s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TASK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Group similar answers
2. Normalize names
3. Rank by frequency
4. Keep EXACTLY %d answers
5. Assign points (sum = 100)
6. Sort DESC by points

%s

%s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
OUTPUT FORMAT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
%s

%s

FINAL INSTRUCTION:
Return ONLY JSON.
No explanation.
Start with { and end with }.
`,
		string(input),
		topN,
		analysisStrictRules(topN),
		analysisPointRules(),
		analysisResponseSchema(),
		analysisFewShotExamples(),
	))
}
