package manager_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestSecretManager(t *testing.T) {
	test.RunSpecs(t, "Secret Manager Suite")
}
