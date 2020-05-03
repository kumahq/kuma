package util_test

import (
	"testing"
	"time"

	"github.com/Kong/kuma/app/kumactl/pkg/util"
)

func TestDuration(t *testing.T) {
	testCases := []struct {
		d    time.Duration
		want string
	}{
		{d: -2 * time.Second, want: "nil"},
		{d: 0, want: "0s"},
		{d: time.Second, want: "1s"},
		{d: time.Minute - time.Millisecond, want: "59s"},
		{d: time.Minute, want: "1m"},
		{d: time.Hour - time.Millisecond, want: "59m"},
		{d: time.Hour, want: "1h"},
		{d: 3*time.Hour - time.Millisecond, want: "2h"},
		{d: 24 * time.Hour, want: "1d"},
		{d: 2*24*time.Hour - time.Millisecond, want: "1d"},
		{d: 365 * 24 * time.Hour, want: "1y"},
		{d: 6*365*24*time.Hour - time.Millisecond, want: "5y"},
		{d: 10 * 365 * 24 * time.Hour, want: "10y"},
	}
	for _, tt := range testCases {
		t.Run(tt.d.String(), func(t *testing.T) {
			if got := util.Duration(tt.d); got != tt.want {
				t.Errorf("Duration = %v, want %v", got, tt.want)
			}
		})
	}
}
