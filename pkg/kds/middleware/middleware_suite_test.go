package middleware_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestMiddleware(t *testing.T) {
	test.RunSpecs(t, "KDS Middleware Suite")
}
