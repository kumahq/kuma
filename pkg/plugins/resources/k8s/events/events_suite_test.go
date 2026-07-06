package events_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestKubernetesEvents(t *testing.T) {
	test.RunSpecs(t, "Kubernetes Events Suite")
}
