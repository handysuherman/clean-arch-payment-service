package helper

import (
	"regexp"
	"strings"
)

func Slugify(text string) string {
	reSpace := regexp.MustCompile(`\s+`)
	text = reSpace.ReplaceAllString(text, "-")

	nonAlphaNum := regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	text = nonAlphaNum.ReplaceAllString(text, "-")

	consecutiveDash := regexp.MustCompile(`-{2,}`)
	text = consecutiveDash.ReplaceAllString(text, "-")

	text = strings.Trim(text, "-")
	// text = strings.ToLower(text)

	return text
}
