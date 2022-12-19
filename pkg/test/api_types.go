package test

import (
	"time"

	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PointerUint32(val uint32) *uint32 {
	return &val
}

func PointerBool(val bool) *bool {
	return &val
}

func ParseDuration(duration string) *k8s.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic(err)
	}

	return &k8s.Duration{Duration: d}
}
