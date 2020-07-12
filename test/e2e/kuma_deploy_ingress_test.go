package e2e

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Ingress", func() {
	const meshWithProvidedCA = `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabledBackend: ca-1
    backends:
    - name: ca-1
      type: provided
      conf:
        cert:
          inline: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN3ekNDQWF1Z0F3SUJBZ0lKQU5CNHVGbHM4eFRaTUEwR0NTcUdTSWIzRFFFQkN3VUFNQkF4RGpBTUJnTlYKQkFNTUJVaGxiR3h2TUI0WERUSXdNRGN3TnpFME5EWXhNMW9YRFRNd01EY3dOVEUwTkRZeE0xb3dFREVPTUF3RwpBMVVFQXd3RlNHVnNiRzh3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRREQxWHFnCjd1cFRYZGEzV3VuL2V5anZ0OEMrMThiRmJ6bnp5ZDN0MmFKazlDMVhXSjBYVHI0R1hCdVNGL2tCZExUSjNrTWsKSmlqV25OR3Z5THBwME1MdFVQT2V5QUxmNnRCZ2VvMk5STUlyeVdZWTNWbDBBZUc2WnlRSUdhZmdPbkk5aE9sRApFVjhsVDE5YngzQVlmMDNlam14RUwzMjhyV1AwbksrdGYzZHpQLzQralRoYVhOWHlEaHprTE9UYmZudkpVbkVzCnQvNytFaUNIem5qd3BFT3d3Y0FYcmlJd25VZUQ2dUpNeVhFM3dhaWVvNkEralFOMUV5bXYxaXUzRnBCV3BWTXEKWjNFTHY2aW5mdGU5S0V5NGFERkpLQWdQMTZPc01jWDR2ckZYQVFnSDdkZ1VCNmJkR1dMUFVLN2hrd2FXWWZjRAphUzMzTnRET3grb2U2Z014QWdNQkFBR2pJREFlTUE4R0ExVWRFd1FJTUFZQkFmOENBUUF3Q3dZRFZSMFBCQVFECkFnSUVNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUREeFJTZEZLUzRuZkUxa3Q1NjFHc3N5ZW9QekV6dVNFV3IKL0hrWU9mRUM2eGlkcDNYYmJPV1VMWDJrRWpWb1VKNUNUaExraDdYaUQ2TmtIaHhPYjQyQWFZd21KOE9JUkFPYQpuREdyTkg1RDRxOE5vMmhYV2tnaTVJeER3UkpLaE5jSHEyTkUvU0VIL3kreVJxelFyS0Y5V2x6SWtqMUxTUXBpCmRNWmhYd2F2UVJrN2Z2aFB1SXJTcWpxU0xaUzRaNUNKMlRMN3VCa3RERm5oam41cWJFakkvSWhLczMwZGNOaXQKcnFtNEVHMU83TVY5enBlVnR3WDlpZVhBaTIvU2s2Njk1SW4waWV5V1h4T1FoSGg0T3A3WXdKa1RYUEtQUVBaSwpnLzIzc2hqVzB5YmFFWjAwTTFJVGo3ZFVRTVRCSDB1THAwMHgyZTlFY2tRcHBxTnlSTVJRCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
        key:
          inline: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRREQxWHFnN3VwVFhkYTMKV3VuL2V5anZ0OEMrMThiRmJ6bnp5ZDN0MmFKazlDMVhXSjBYVHI0R1hCdVNGL2tCZExUSjNrTWtKaWpXbk5Hdgp5THBwME1MdFVQT2V5QUxmNnRCZ2VvMk5STUlyeVdZWTNWbDBBZUc2WnlRSUdhZmdPbkk5aE9sREVWOGxUMTliCngzQVlmMDNlam14RUwzMjhyV1AwbksrdGYzZHpQLzQralRoYVhOWHlEaHprTE9UYmZudkpVbkVzdC83K0VpQ0gKem5qd3BFT3d3Y0FYcmlJd25VZUQ2dUpNeVhFM3dhaWVvNkEralFOMUV5bXYxaXUzRnBCV3BWTXFaM0VMdjZpbgpmdGU5S0V5NGFERkpLQWdQMTZPc01jWDR2ckZYQVFnSDdkZ1VCNmJkR1dMUFVLN2hrd2FXWWZjRGFTMzNOdERPCngrb2U2Z014QWdNQkFBRUNnZ0VCQU1GTmpkZ2hQS2ZCcnRvYUlYUVBhOThEc0h3d25ZSHhRbkVEeDg2cHpvUjgKQ2UxNENNZ2k3NnR6YTd1UGNqa2ZxL3kvS2VNYXo2RFg5cHJmTmpLUTRIaEVPZFYzZEc3MloyMTBTeGt3ejhGTQo4VHlGOFhCekV3OWVFOUR6RWlSaFRMYXc1VmRRWkd4OXBwRC9rZ1I4Vks3a1FyWWpjcWUxTno4VEVzM2RUbGt4CkM0ZDBZVnNGTEpjUWFmdU13SGNhQmxBMG8zV21Wd25tcEFHSHRSVVpZeFdQejltaXZEU2w0dENEaTFabHp3U3gKUFJlNHNRTDYrdk1JcWVmWDFjYVFuUlpOOURMd1ZCaTZCOEdHbXZyNG5TM0w1SitOWEtFV2JsMHlmZTMzdU1MWgpHWDI1SmExc0p6S09kK29iSnZPZTJ1WmUxaHVnTVFyQWFiUGFsVDJ2bStFQ2dZRUE5bGNwQ3UzRkFsOUM5cnFLClNSeVFad3lpMW05cXRUbGZFMTZRWUZjTWlTMXVpeThzajFVbEU3blNITktoNFFBV1Q3ZFhVdFFBQWJ4Y1diWkYKNGp3S1VEYXZoMnhtdHZGUDhTb0lydWlFTVpKWUtOK1crbjAxZTZWMWlJYUVScW1aNUdhWUZua2djNUwwUUVqcgpLK2FoNXpvQzdnSFFKNGRoVUpIS1RLVFhpRlVDZ1lFQXk0TlJqeFh1NmNIZHRHL1g0ODg1Y0ZzNmxsNlNiNWR6CjJHQ2pjdTFKMlBUNzZ3anVJOHNOM09pVktMQVNtWVhycjcxTnZXOFA3RzVYYTh2LytQS0NwYnpYTWgxcmhmOHYKaU9mMXZmSFhBVFBvdUJLZnYvV3llSDBDUGFaeFVXRTdqMFkrRmZzcmdHSTAvRy8wUjdwYjQ2dmtGampQTHI0VwpweGNlSy9pMEcyMENnWUE0b2tNNlV2MjNGT1dWU2IrZkhXVUplL3MzNTNlVjRIRytSMEJVRmM4NC9tdnFyZGJGCndTSjhEWDJEeU4wVW1HdUl1akxtUlAwWGFSR21RbVNBcGFNTlcvVXc0amdmR1ExeStXSHpyRnN2OW1BMFRXc3QKZlhtOVNvWGg5R01XeDhrc25IV2N2UTQ3NCs0cGxWb1R4cnMwS0w4aHJ1TUhJM1c1Q3p1Q01XZW4zUUtCZ1FDNwp5bThsODMxRUltb3NKOUExSEhES0pzU0hJTGxMVTV2SUhGUjJwbE13YWM5VDhDZWV5NjM5SEhrVzFISTFUQWhSCllBTXVQQitiY2E0bGdGYXhKMFk3SFdnTmpHdzlkMTRybks5OEdIN25Vemo1TWVaTFFiTHZ6NXFUdk5SdjNhTVIKOENVMkwxRFM1TXd6N0RLalJXbXBTbUhyeDN3V2k3MW5iY09mbTV6R2VRS0JnRVVLZ25uUmJrbU1lZHhXN3hZcApmVXlFN2tPNjRnZ05MMk9zTnpoZjhmZlpETk5BNlFVMnZVMHVEQmIrYkhMeWU2Q0ZlSWwzdXowVWM1K09rMXpWCmZmZ3l3bE9YOXZsTWVZRkZlN0I2Tk55Vk83K1A2bHJhSTliVi9yWmZlOGQ2cmEzS3RaMGtvcFM0Sys4U1FrdXEKK3Uxd0Q4RWx6eS9QbG95cVVHN2ptamxrCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K
`

	const trafficPermission = `
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  namespace: default
  name: everything
spec:
  sources:
  - match:
      service: '*'
  destinations:
  - match:
      service: '*'
`

	namespaceWithSidecarInjection := func(namespace string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s
  labels:
    kuma.io/sidecar-injection: "enabled"
`, namespace)
	}

	headlessService := func(namespace string, port int32) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Service
metadata:
  name: echo-server
  namespace: %s
spec:
  clusterIP: None
  selector:
    app: test-app
  ports:
    - protocol: TCP
      port: %d
      targetPort: %d
`, namespace, port, port)
	}

	fakeDpForOutbound := func(namespace, ingressIP string, port int32) string {
		return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Dataplane
mesh: default
metadata:
  name: ingress-dp
  namespace: %s
spec:
  networking:
    address: %s
    inbound:
    - port: %d
      tags:
        service: echo-server_%s_svc_80
`, namespace, ingressIP, port, namespace)
	}

	var c1 Cluster
	var c2 Cluster
	var ingress IngressDesc

	BeforeEach(func() {
		clusters, err := NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		c1 = clusters.GetCluster(Kuma1)
		c2 = clusters.GetCluster(Kuma2)

		err = NewClusterSetup().
			Install(Kuma()).
			Install(YamlK8s(meshWithProvidedCA)).
			Install(YamlK8s(trafficPermission)).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(EchoServerK8s()).
			Install(Ingress(&ingress)).
			Setup(c1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma()).
			Install(YamlK8s(meshWithProvidedCA)).
			Install(YamlK8s(trafficPermission)).
			Install(YamlK8s(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClientK8s()).
			Install(YamlK8s(headlessService(TestNamespace, ingress.Port))).
			Install(YamlK8s(fakeDpForOutbound(TestNamespace, ingress.IP, ingress.Port))).
			Setup(c2)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		_ = c1.DeleteKuma()
		_ = c2.DeleteKuma()
		_ = k8s.KubectlDeleteFromStringE(c1.GetTesting(), c1.GetKubectlOptions(), namespaceWithSidecarInjection(TestNamespace))
		_ = k8s.KubectlDeleteFromStringE(c2.GetTesting(), c2.GetKubectlOptions(), namespaceWithSidecarInjection(TestNamespace))
	})

	It("should deploy ingress", func() {
		pods, err := k8s.ListPodsE(
			c2.GetTesting(),
			c2.GetKubectlOptions(TestNamespace),
			kube_meta.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		_, stderr, err := c2.ExecWithRetries(TestNamespace, pods[0].GetName(), "demo-client",
			"curl", "-v", "-m", "3", fmt.Sprintf("%s:%d", ingress.IP, ingress.Port))
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
	})
})
