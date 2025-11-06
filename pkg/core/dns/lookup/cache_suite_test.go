package lookup_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestDNSCaching(t *testing.T) {
	test.RunSpecs(t, "DNS with cache Suite")
}
