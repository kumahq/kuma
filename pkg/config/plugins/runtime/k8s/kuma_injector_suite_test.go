package k8s_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKumaInjector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KumaInjector Suite")
}
