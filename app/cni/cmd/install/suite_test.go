package install_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestInstall(t *testing.T) {
	test.RunSpecs(t, "Install CNI Suite")
}
