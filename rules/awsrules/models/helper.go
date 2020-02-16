package models

import "strings"

func truncateLongMessage(str string) string {
	limit := 80

	str = strings.Replace(str, "\r\n", "\n", -1)
	str = strings.Replace(str, "\n", "\\n", -1)

	r := []rune(str)
	if len(r) > limit {
		return string(r[0:limit]) + "..."
	}

	return str
}
