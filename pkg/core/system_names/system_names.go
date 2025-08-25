package system_names

import (
	"regexp"
	"strings"

	"github.com/kumahq/kuma/pkg/core/kri"
)

const (
	SystemPrefix          = "system_"
	sectionSeparator      = "_"
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

func AsSystemName[T string | kri.Identifier](name T) string {
	switch v := any(name).(type) {
	case kri.Identifier:
		if v != (kri.Identifier{}) {
			return SystemPrefix + v.String()
		}
		return ""
	case string:
		return SystemPrefix + v
	}
	panic("Unsupported type for AsSystemName: " + any(name).(string))
}

func MustBeSystemName(name string) string {
	if !cleanNameRegex.MatchString(name) {
		panic("Invalid system name: " + name + ". Only lowercase letters, numbers, hyphens, and underscores are allowed.")
	}
	return SystemPrefix + name
}

func JoinSections(sections ...string) string {
	return strings.Join(sections, sectionSeparator)
}

func JoinSectionParts(parts ...string) string {
	return strings.Join(parts, sectionPartsSeparator)
}

func GetNameOrDefault(predicate bool) func(name string, defaultName string) string {
	return func(name string, defaultName string) string {
		if predicate {
			return name
		}
		return defaultName
	}
}
