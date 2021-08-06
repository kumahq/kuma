package v3_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestServer(t *testing.T) {
	test.RunSpecs(t, "SDS Server Suite")
}
