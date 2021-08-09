package api_server_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestWs(t *testing.T) {
	test.RunSpecs(t, "API Server")
}
