package lookup_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestDNSCaching(t *testing.T) {
	test.RunSpecs(t, "DNS with cache Suite")
}
