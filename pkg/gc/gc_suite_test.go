package gc_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestGC(t *testing.T) {
	test.RunSpecs(t, "GC Suite")
}
