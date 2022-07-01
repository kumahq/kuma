package attachment_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestRouteAttachment(t *testing.T) {
	test.RunSpecs(t, "Gateway API route attachment support")
}
