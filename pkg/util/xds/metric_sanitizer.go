package xds

import (
	"regexp"
	"strings"
)

var (
	whitespaces  = regexp.MustCompile(`\s+`)
	illegalChars = regexp.MustCompile(`[^a-zA-Z_\-0-9{}=]`)
)

func SanitizeMetric(metric string) string {
	result := whitespaces.ReplaceAllString(metric, "_")
	result = strings.ReplaceAll(result, "/", "_")
	result = illegalChars.ReplaceAllString(result, "_")
	return result
}
