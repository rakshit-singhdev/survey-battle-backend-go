package redis

import "fmt"

func PresenceTeamsKey(sessionID string) string {
	return fmt.Sprintf(
		"presence:%s:teams",
		sessionID,
	)
}

func PresenceCountKey(sessionID, role string) string {
	return fmt.Sprintf(
		"presence:%s:%s",
		sessionID,
		role,
	)
}
