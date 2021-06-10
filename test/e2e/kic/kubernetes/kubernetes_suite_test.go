package kubernetes_test

import (
	"testing"

	"github.com/kumahq/kuma/test/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2EKICKubernetes(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "KIC Kubernetes Suite")
	} else {
		t.SkipNow()
	}
}
