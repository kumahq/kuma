package system_names

import (
	"regexp"
	"strings"
)

const SystemPrefix = "system_"

const separator = "_"

var cleanNameRegex = regexp.MustCompile(`[^a-z0-9-_]`)

func IsSystem(name string) bool {
	return strings.HasPrefix(SystemPrefix, name)
}

func AsSystemName(name string) string {
	return SystemPrefix + cleanName(name)
}

func Join(parts ...string) string {
	return strings.Join(parts, separator)
}

func cleanName(name string) string {
	return cleanNameRegex.ReplaceAllString(name, "")
}

func GetNameOrDefault(predicate bool) func(name string, defaultName string) string {
	return func(name string, defaultName string) string {
		if predicate {
			return name
		}
		return defaultName
	}
}
