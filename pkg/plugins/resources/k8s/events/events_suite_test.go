package events_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestKubernetesEvents(t *testing.T) {
	test.RunSpecs(t, "Kubernetes Events Suite")
}
