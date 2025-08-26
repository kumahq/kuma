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

func Date(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}

func Number(v interface{}) string {
	return fmt.Sprintf("%d", v)
}

func Ago(m *time.Time, now time.Time) string {
	if m == nil {
		return "never"
	}
	d := now.Sub(*m)
	return Duration(d)
}

// TimeSince to calculate age of resources
func TimeSince(m, now time.Time) string {
	d := now.Sub(m)
	return Duration(d)
}

// Duration returns a readable representation of the provided time
func Duration(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds < -1 {
		return "never"
	} else if seconds < 0 {
		return "0s"
	} else if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if minutes := int(d.Minutes()); minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	} else if hours := int(d.Hours()); hours < 24 {
		return fmt.Sprintf("%dh", hours)
	} else if hours < 24*365 {
		return fmt.Sprintf("%dd", hours/24)
	}
	return fmt.Sprintf("%dy", int(d.Hours()/24/365))
}
