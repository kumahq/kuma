package validators

import (
    "fmt"
    "strings"
)

func MustHaveOnlyOne(entity string, allowedValues ...string) string {
    return fmt.Sprintf(`%s must have only one type defined: %s`, entity, strings.Join(allowedValues, ", "))
}

func MustBeDefined() string {
    return "must be defined"
}

func MustNotBeEmpty() string {
    return "must not be empty"
}

