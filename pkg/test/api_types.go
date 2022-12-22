package test

import (
	"time"

	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PointerOf[T any](value T) *T {
	return &value
}

func ParseDuration(duration string) *k8s.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic(err)
	}

	return &k8s.Duration{Duration: d}
}
