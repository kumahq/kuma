package authn_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestAuthn(t *testing.T) {
	test.RunSpecs(t, "HDS Authn Suite")
}
