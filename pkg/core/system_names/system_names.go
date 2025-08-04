package system_names

import (
	"regexp"
	"strings"
)

const (
	SystemPrefix = "system_"
	sectionSeparator = "_"
	sectionPartsSeparator = "-"
)

var cleanNameRegex = regexp.MustCompile(`([a-z0-9-]*_?)+`)

func IsSystem(name string) bool {
	return strings.HasPrefix(SystemPrefix, name)
}

func CleanName(name string) string {
	matches := cleanNameRegex.FindAllString(name, -1)
	return strings.Join(matches, sectionPartsSeparator)
}

func AsSystemName(name string) string {
	return SystemPrefix + name
}

func MustBeSystemName(name string) string {
	if !cleanNameRegex.MatchString(name) {
		panic("Invalid system name: " + name + ". Only lowercase letters, numbers, hyphens, and underscores are allowed.")
	}
	return SystemPrefix + name
}

func Join(parts ...string) string {
	return strings.Join(parts, sectionSeparator)
}

func GetNameOrDefault(predicate bool) func(name string, defaultName string) string {
	return func(name string, defaultName string) string {
		if predicate {
			return name
		}
		return defaultName
	}
}
