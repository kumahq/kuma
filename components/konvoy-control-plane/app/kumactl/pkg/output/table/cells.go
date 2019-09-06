package table

import (
	"fmt"
	"time"
)

func Number(v interface{}) string {
	return fmt.Sprintf("%d", v)
}

func Ago(m *time.Time, now time.Time) string {
	if m == nil {
		return "never"
	}
	return now.Sub(*m).Truncate(time.Second).String()
}
