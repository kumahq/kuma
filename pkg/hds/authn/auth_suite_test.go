package authn_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestAuthn(t *testing.T) {
	test.RunSpecs(t, "HDS Authn Suite")
}
