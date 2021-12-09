package gc_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestGC(t *testing.T) {
	test.RunSpecs(t, "GC Suite")
}
