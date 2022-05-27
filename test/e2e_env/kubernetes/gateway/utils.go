package gateway

import (
	"fmt"
	"net/url"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

func proxyRequestToGateway(cluster Cluster, instance string, gateway string, port int, opts ...client.CollectResponsesOptsFn) {
	Logf("expecting 200 response from %q", gateway)
	Eventually(func(g Gomega) {
		target := fmt.Sprintf("http://%s:%d/%s",
			gateway, port, path.Join("test", url.PathEscape(GinkgoT().Name())),
		)

		response, err := client.CollectResponse(cluster, "demo-client", target, opts...)

		g.Expect(err).To(Succeed())
		g.Expect(response.Instance).To(Equal(instance))
	}, "60s", "1s").Should(Succeed())
}
