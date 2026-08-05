package manager_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestSecretManager(t *testing.T) {
	test.RunSpecs(t, "Secret Manager Suite")
}
