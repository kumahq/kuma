package manager_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestSecretManager(t *testing.T) {
	test.RunSpecs(t, "Secret Manager Suite")
}
