package server_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestServer(t *testing.T) {
	test.RunSpecs(t, "Dataplane Token Server Suite")
}
