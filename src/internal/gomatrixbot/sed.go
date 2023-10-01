package gomatrixbot

import (
	"fmt"
	"regexp"
	"strings"
)

func (mtrx *MtrxClient) sedThis(channelID string, content string) string {
	formatted := strings.Split(strings.TrimSpace(content), "/")
	if len(formatted) != 2 {
		return "Invalid sed. Use `.sed find/replace`"
	}

	msg, err := mtrx.searchRecentMessages(channelID, formatted[0])
	if err != nil {
		return fmt.Sprintf("Cannot find message matching %s", formatted[0])
	}

	return fmt.Sprintf(
		"%s: %s",
		msg.User,
		caseInsensitiveReplace(msg.Content, formatted[0], formatted[1]),
	)
}

func caseInsensitiveReplace(subject string, search string, replace string) string {
	searchRegex := regexp.MustCompile("(?i)" + search)
	return searchRegex.ReplaceAllString(subject, replace)
}
