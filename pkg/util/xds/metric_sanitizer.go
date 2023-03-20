package xds

import (
	"regexp"
)

var illegalChars = regexp.MustCompile(`[^a-zA-Z_\-0-9]`)

// We need to sanitize metrics in order to  not break statsd and prometheus format.
// StatsD only allow [a-zA-Z_\-0-9.] characters, everything else is removed
// Extra dots breaks many regexes that converts statsd metric to prometheus one with tags
func SanitizeMetric(metric string) string {
	return illegalChars.ReplaceAllString(metric, "_")
}
