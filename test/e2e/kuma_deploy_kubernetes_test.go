package e2e

import (
	"fmt"

	"github.com/Kong/kuma/test/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeEach(func() {

})

var _ = Describe("Test K8s deployment with `kumactl install control-plane`", func() {

	It("Deploy on Single K8s cluster and verify the Kuma CP REST API is accessible", func(done Done) {
		clusters, err := framework.NewK8sClusters(
			[]string{framework.Kuma1},
			framework.Verbose)
		Expect(err).ToNot(HaveOccurred())
		c := clusters.GetCluster(framework.Kuma1)

		err = c.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		err = c.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		logs, err := c.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(logs)

		_ = c.DeleteKuma()

		// completed
		close(done)
	}, 180)

	It("Deploy on Two K8s cluster and verify the Kuma CP REST API is accessible. Use the Clusters Interface.", func(done Done) {
		clusters, err := framework.NewK8sClusters(
			[]string{framework.Kuma1, framework.Kuma2},
			framework.Verbose)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		err = clusters.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		logs, err := clusters.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(logs)

		_ = clusters.DeleteKuma()

		// completed
		close(done)
	}, 180)
})
