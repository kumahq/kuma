package model

import (
	"strings"
	"unicode"
)

func DisplayName(resType string) string {
	displayName := ""
	for i, c := range resType {
		if unicode.IsUpper(c) && i != 0 {
			displayName += " "
		}
		displayName += string(c)
	}
	return displayName
}

func PluralType(resType string) string {
	switch {
	case strings.HasSuffix(resType, "ay"):
		return resType + "s"
	case strings.HasSuffix(resType, "y"):
		return strings.TrimSuffix(resType, "y") + "ies"
	case strings.HasSuffix(resType, "s"), strings.HasSuffix(resType, "sh"), strings.HasSuffix(resType, "ch"):
		return resType + "es"
	default:
		return resType + "s"
	}
}
