package httpclient_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestHTTPClient(t *testing.T) {
	test.RunSpecs(t, "HTTP Client Suite")
}
