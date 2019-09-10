package table

import (
	"fmt"
	"time"
)

func Check(selected bool) string {
	if selected {
		return "*"
	}
	return ""
}

func OnOff(on bool) string {
	if on {
		return "on"
	} else {
		return "off"
	}
}

func Number(v interface{}) string {
	return fmt.Sprintf("%d", v)
}

func Ago(m *time.Time, now time.Time) string {
	if m == nil {
		return "never"
	}
	return now.Sub(*m).Truncate(time.Second).String()
}
