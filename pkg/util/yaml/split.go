package yaml

import (
	"regexp"
	"strings"
)

var sep = regexp.MustCompile("(?:^|\\s*\n)---\\s*")

// SplitYAML takes YAMLs separated by `---` line and splits it into multiple YAMLs. Empty entries are ignored
func SplitYAML(yamls string) []string {
	var result []string
	// Making sure that any extra whitespace in YAML stream doesn't interfere in splitting documents correctly.
	trimYAMLs := strings.TrimSpace(yamls)
	docs := sep.Split(trimYAMLs, -1)
	for _, doc := range docs {
		if doc == "" {
			continue
		}
		doc = strings.TrimSpace(doc)
		result = append(result, doc)
	}
	return result
}
