package validation

import (
    "fmt"
    "strings"
)

func Bool2Int(b bool) int {
    if b {
        return 1
    }
    return 0
}

func MustHaveOnlyOne(entity string, allowedValues... string) string {
    return fmt.Sprintf(`%s must have only one type defined: %s`, entity, strings.Join(allowedValues, ", "))
}

func MustBeDefined() string {
    return "must be defined"
}