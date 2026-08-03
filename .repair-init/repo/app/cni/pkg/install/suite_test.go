package install_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestInstall(t *testing.T) {
	test.RunSpecs(t, "Install CNI Suite")
}
