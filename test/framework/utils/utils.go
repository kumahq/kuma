package utils

import (
	"fmt"
	"strings"
)

func ShellEscape(arg string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "\\'"))
}
