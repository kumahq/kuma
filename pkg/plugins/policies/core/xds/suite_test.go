package xds_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestOtel(t *testing.T) {
	test.RunSpecs(t, "Otel Backend Resolution")
}
