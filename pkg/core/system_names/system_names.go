package system_names

import (
	"strings"
)

const SystemPrefix = "system_"

func IsSystem(name string) bool {
	return strings.HasPrefix(SystemPrefix, name)
}

func AsSystemName(name string) string {
	return SystemPrefix + name
}
