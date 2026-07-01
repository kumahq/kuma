package events

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKubernetesEvents(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubernetes Events Suite")
}
