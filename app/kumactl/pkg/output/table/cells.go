package table

import (
	"fmt"
	"time"

	"github.com/Kong/kuma/app/kumactl/pkg/util"
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

// TimeSince to calculate age of resources
func TimeSince(m time.Time, now time.Time) string {
	d := now.Sub(m)
	return util.Duration(d)
}
