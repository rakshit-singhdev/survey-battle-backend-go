package redis

import (
	"fmt"
	"strings"
)

func GameStateKey(sessionID string) string {
	return "game:" + sessionID
}

func GameLockKey(sessionID string) string {
	return "lock:game:" + sessionID
}

func BuzzerLockKey(sessionID, roundID string) string {
	return fmt.Sprintf("buzz:%s:%s", sessionID, roundID)
}

func RosterCacheKey(sessionID string) string {
	return "roster:" + sessionID
}

func GameDefCacheKey(gameID string) string {
	return "gamedef:" + gameID
}

func SurveyDefCacheKey(surveyID string) string {
	return "surveydef:" + surveyID
}

func SurveyCodeCacheKey(code string) string {
	return "surveycode:" + strings.ToUpper(code)
}
