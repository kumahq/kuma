package kubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2EKICKubernetes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KIC Kubernetes Suite")
}
