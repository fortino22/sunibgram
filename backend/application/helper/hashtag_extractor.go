package helper

import (
	"regexp"
	"strings"
)

var hashtagPattern = regexp.MustCompile(`#[a-zA-Z0-9_]+`)

func ExtractHashtags(content string) []string {
	matches := hashtagPattern.FindAllString(content, -1)
	if len(matches) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(matches))
	hashtags := make([]string, 0, len(matches))
	for _, match := range matches {
		name := strings.ToLower(match[1:])
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		hashtags = append(hashtags, name)
	}
	return hashtags
}
