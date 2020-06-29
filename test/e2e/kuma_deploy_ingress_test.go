package e2e

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/Kong/kuma/test/framework"
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
          inline: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN3ekNDQWF1Z0F3SUJBZ0lKQVBUUnZLUUJPbDY0TUEwR0NTcUdTSWIzRFFFQkN3VUFNQkF4RGpBTUJnTlYKQkFNTUJVaGxiR3h2TUI0WERUSXdNRFl3TkRJek5URTFPRm9YRFRJd01EY3dOREl6TlRFMU9Gb3dFREVPTUF3RwpBMVVFQXd3RlNHVnNiRzh3Z2dFaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQkR3QXdnZ0VLQW9JQkFRQ2VNSjdFClVpNjFnMGFOZnprci80RnVHQXlyUXQ0d2Q5SkhNVWtjUXBTaGw5RFZxeVN0amNoajdyVmlMdXBNeW5ia1lGRysKNkxIRk5nYTNPSmFtWVA4R2RkM2VkaUZYOEFTakxCYlhVTzlGMGtTeDYraTdYbWg2MGN0Z09hMjR3L0hYaUtOZQp4NUkwdVlWbmkxd0dMd2pnVGhFNmIxdEx4WDgwNkV4eHIvd2t6bnFSUmRYREhGaUo3K2FGcXg4d3I1RTRqaWNaCnovbFNkSEVtWXNGTXJVTjhwYitwVzlWNVEyT0xhUTI1Z0xHbHRlVzVCdXBlYzBvYUUxR2hrdDdydzFnOERjbUoKZE1WbStsVkFtOVE1dzNNWGE1Mlo4M3JwL3pFTjZrdnV3VVRQY3BoVWozWmkrM1I3NlVWQVJ5c3dseFE3N0RwMAo5WnBLanRVdng5S2VsY1FsQWdNQkFBR2pJREFlTUE4R0ExVWRFd1FJTUFZQkFmOENBUUF3Q3dZRFZSMFBCQVFECkFnSUVNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUFiRGUvYkkyZ1hkQzRYdUxOU3BTd2U2cEpwN1R1bXNnRHIKYk01OEpMT0JPQU4xM3liK2RYYnFoNHVJNmRYMENJNDJIcTNVQWFEZ0VSeUZ2RkwzTTdPcUg2REx3dWxOemNFRgpBdlQyRmpTV3ozSXdUd1FKVW1DejRFeEJVK05JY0FLUElHM1Bxek45SE5BZ0hDdVJSREdDZ0tMMFhTVThNUk14CjBKUkxlQzh5R3BBbnpVZDhlREFjUnpFeGFGYXNpUXJmSjI3elVZK21aNjNueEJ3R3lqaXlsVWp1OTdEUS9BWFMKc3JVK0xNWTJtcDhCRm9SK2ZQMytmTlNzOWU4N254L29SWDltenBrOXJXd2dKQkliREZDQ3JubnBpTzNTMGp4TAp5RWZlaGZMYXAycU9Sd2JvZE1ZZ09TZHFGQzNOWVpHNlBSOTk2cGN6OTNFZUZGTlFwMzhVCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
        key:
          inline: LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRQ2VNSjdFVWk2MWcwYU4KZnprci80RnVHQXlyUXQ0d2Q5SkhNVWtjUXBTaGw5RFZxeVN0amNoajdyVmlMdXBNeW5ia1lGRys2TEhGTmdhMwpPSmFtWVA4R2RkM2VkaUZYOEFTakxCYlhVTzlGMGtTeDYraTdYbWg2MGN0Z09hMjR3L0hYaUtOZXg1STB1WVZuCmkxd0dMd2pnVGhFNmIxdEx4WDgwNkV4eHIvd2t6bnFSUmRYREhGaUo3K2FGcXg4d3I1RTRqaWNaei9sU2RIRW0KWXNGTXJVTjhwYitwVzlWNVEyT0xhUTI1Z0xHbHRlVzVCdXBlYzBvYUUxR2hrdDdydzFnOERjbUpkTVZtK2xWQQptOVE1dzNNWGE1Mlo4M3JwL3pFTjZrdnV3VVRQY3BoVWozWmkrM1I3NlVWQVJ5c3dseFE3N0RwMDlacEtqdFV2Cng5S2VsY1FsQWdNQkFBRUNnZ0VBRG9KUkhacVlGQ2Z0UWE4b2xFT0VJSS91SVlzcGkvS0JnK3dlVTR3N3k3SjgKQWcwSGVTK204SnVGWVhNQ0pIYnhmckxpN0lxMU8yeGdJMC82YVZvK0tkNkhzZzdOc2g0ZW5zUzlkNVJCemZxaQpPRnQxNWpHYmphQk9jZzM0UkJrY3huTU80UE9YRW1UdHVuaUt3VHB4S3ZtZUZPai95NnhhcFlTazlreDQ2UHN1ClVTSHZOMVhSb1VQbHFiYVREdVNZY2IwZERleHZmdG42UVkxZ3BGblFlb3JoN0xrOE84OGxzMW5FZ0V6R1ZiNlEKK1VTUkZ6Qm5PTCtqUFJvSzh0b2J3YkhXVFRNTHNrb1Vua1FXemowOTBQRlJOd3dmS3VlSE9sTHlweU1SamJUdgp4NnVCK3JSRnJJMFRjd2grenFJRGVxM3JIZy9ZL1krellEV0twU3FGUVFLQmdRRE1uOTVGT1E0dFJuT2pEbUI1CnVQSUIxamVPcElHcEFpZzgyVEM0SDJDUEFxdTNyQ3B1Wjh2SnRkd1UwQ0pIWVFCaXNpZDBZd1dtMHdoYjFONFIKYkVGL0l2U1FjYWdPbmV5UnRMUkdPSGJVRU1uRURXUmN3R0dQb2JIRkpNQnNEYVlOWm01WXBLMlR1dDJjZ1YyVwpVaG5NRVBOWERPMnU0Qzg2NDF3USt1NG9zUUtCZ1FERjZESFlkYjAxcHRKM1JUS0NKRkw5TkxPKzN4NGcwSzZmCk9BdG1WR3RLZFZHemdwTFkzUEJ5RSt4S3VLMmxPMEl4cGFtbi9kVzA4Zi9ZVnVUcytidnB0MG96TnRUN0NtVG4KYnhXaDdDeGhHQ0ZsV2hQRnFaQ1h3dXgyMjNuN1dUa2ZkQTduT0pDNTFleXlmbStIaUkyaGt3cXk4UDlRZ04rVgppRTNTcFFLdnRRS0JnUURGSXlROTVxa00zM29hMW9nRjNUTnlwNUlnMzhaM01EZlozNWs3V2lkcHdEWDFuMjNGCnJrUThVZlAxTFV4SkhtQUR0Z1dpOEorS1NIZ2VHT2ZWTzBtaWxlZXVuWUUraTlGVjB4VjNMWUQxOERLaXFoQk4KOTU1R2hZNUNFNVU2eEs5ODYzbFY5MW12SVBIT2pTZS80ZHN1cWduMmpPTVVmckdoOTFkRW1Ld0lFUUtCZ0ZzMApDRlNTM2VGdHdheEpiVjlnVWdaeVZTdHZNems0TW1FWnVOY3RyRXdpQ01iTE05VlE3RllHTEd5Njh2c2tkZnJmCk4zSTlubERHL1hxN2dNQmN6bWFFbTJOQ3I2QUpTRHNIakZhVXVsYjhnZGR0VFpOWDgxU2M5ZEJJa014dWI4NjQKODIxSE9oc0tKUXlWQzl6UDUwVkF1RHVDcUlaMi9aS3h2L3VGSTluSkFvR0JBS3pSS0RyTTJ0Qk5qUU9xM0FCZAowUGpFbUM3M3BENjN3WjF6TUdabis4OFFZbzhBM3FQbkZKYkNyTW1KWnBlc2pmU2M4K01SMS9EODBDSkZmNkRhCnJpbTlQZU9XQXArSEJnWm5ZWFhKRk9CdlU4VXRiQklvemlKUlhsTkEwaFhINy91YU5vUlBuVmJNdDZPNzFjWk4KTzVwODB0OE9TakRBaTRTMWcyY2pjdjYzCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K
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
			Install(Yaml(meshWithProvidedCA)).
			Install(Yaml(trafficPermission)).
			Install(Yaml(namespaceWithSidecarInjection(TestNamespace))).
			Install(EchoServer()).
			Install(Ingress(&ingress)).
			Setup(c1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma()).
			Install(Yaml(meshWithProvidedCA)).
			Install(Yaml(trafficPermission)).
			Install(Yaml(namespaceWithSidecarInjection(TestNamespace))).
			Install(DemoClient()).
			Install(Yaml(headlessService(TestNamespace, ingress.Port))).
			Install(Yaml(fakeDpForOutbound(TestNamespace, ingress.IP, ingress.Port))).
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
			"curl", "-v", fmt.Sprintf("%s:%d", ingress.IP, ingress.Port))
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
	})
})
