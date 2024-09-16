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
	loweredResType := strings.ToLower(resType)
	switch {
	case strings.HasSuffix(loweredResType, "ay"):
		return resType + "s"
	case strings.HasSuffix(loweredResType, "y"):
		return strings.TrimSuffix(resType, "y") + "ies"
	case strings.HasSuffix(loweredResType, "s"), strings.HasSuffix(loweredResType, "sh"), strings.HasSuffix(loweredResType, "ch"):
		return resType + "es"
	default:
		return resType + "s"
	}
}
