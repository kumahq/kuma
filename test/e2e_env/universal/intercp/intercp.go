package intercp

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func InterCP() {
	It("should run inter cp server", func() {
		Eventually(func(g Gomega) {
			_, _, err := universal.Cluster.GetKuma().Exec("nc", "-z", "localhost", "5683")
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})
}
