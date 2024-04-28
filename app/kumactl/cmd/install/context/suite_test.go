package context

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestInstall(t *testing.T) {
	test.RunSpecs(t, "kubectl install Suite")
}
