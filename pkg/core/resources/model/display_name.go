package model

import (
	"strings"
	"unicode"
)

func DisplayName(resType string) string {
	var displayName strings.Builder
	for i, c := range resType {
		if unicode.IsUpper(c) && i != 0 {
			displayName.WriteString(" ")
		}
		displayName.WriteRune(c)
	}
	return displayName.String()
}

func PluralType(resType string) string {
	loweredResType := strings.ToLower(resType)
	switch {
	case strings.HasSuffix(loweredResType, "ay"):
		return resType + "s"
	case strings.HasSuffix(loweredResType, "y"):
		return resType[:len(resType)-1] + "ies"
	case strings.HasSuffix(loweredResType, "s"), strings.HasSuffix(loweredResType, "sh"), strings.HasSuffix(loweredResType, "ch"):
		return resType + "es"
	default:
		return resType + "s"
	}
}
