// +build kind

package cmd_test

import (
	"github.com/Kong/kuma/tools/test/framework"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegrationCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration CMD Suite")
}

var _ = Describe("Test K8s deployment with `kumactl install control-plane`", func() {

	t := framework.NewK8sTest(1, "")
	t.DeployKumaOnK8sCluster(1)
})
