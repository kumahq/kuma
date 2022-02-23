package gateway

import (
	"fmt"
	"net/url"
	"path"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
)

// OptEnableMeshMTLS is a Kuma deployment option that enables mTLS on the default Mesh.
var OptEnableMeshMTLS = framework.WithMeshUpdate(
	"default",
	func(mesh *mesh_proto.Mesh) *mesh_proto.Mesh {
		mesh.Mtls = &mesh_proto.Mesh_Mtls{
			EnabledBackend: "builtin",
			Backends: []*mesh_proto.CertificateAuthorityBackend{
				{Name: "builtin", Type: "builtin"},
			},
		}
		return mesh
	},
)

// ProxySimpleRequests tests that basic HTTP requests are proxied to the echo-server.
func ProxySimpleRequests(cluster framework.Cluster, instance string, gateway string, opts ...client.CollectResponsesOptsFn) {
	framework.Logf("expecting 200 response from %q", gateway)
	Eventually(func(g Gomega) {
		target := fmt.Sprintf("http://%s/%s",
			gateway, path.Join("test", url.PathEscape(GinkgoT().Name())),
		)

		opts = append(opts, client.WithHeader("Host", "example.kuma.io"))
		response, err := client.CollectResponse(cluster, "gateway-client", target, opts...)

		g.Expect(err).To(Succeed())
		g.Expect(response.Instance).To(Equal(instance))
		g.Expect(response.Received.Headers["Host"]).To(ContainElement("example.kuma.io"))
	}, "60s", "1s").Should(Succeed())
}

// ProxySecureRequests tests that basic HTTPS requests are proxied to the echo-server.
func ProxySecureRequests(cluster framework.Cluster, instance string, gateway string, opts ...client.CollectResponsesOptsFn) {
	framework.Logf("expecting 200 response from %q", gateway)
	Eventually(func(g Gomega) {
		target := fmt.Sprintf("https://%s/%s",
			gateway, path.Join("https", "test", url.PathEscape(GinkgoT().Name())),
		)

		opts = append(opts,
			client.Insecure(),
			client.WithHeader("Host", "example.kuma.io"))
		response, err := client.CollectResponse(cluster, "gateway-client", target, opts...)

		g.Expect(err).To(Succeed())
		g.Expect(response.Instance).To(Equal(instance))
		g.Expect(response.Received.Headers["Host"]).To(ContainElement("example.kuma.io"))
	}, "60s", "1s").Should(Succeed())
}

// ProxyRequestsWithMissingPermission deletes the default TrafficPermission and expects
// that to cause proxying requests to fail.
//
// In mTLS mode, only the presence of TrafficPermission rules allow services to receive
// traffic, so removing the permission should cause requests to fail. We use this to
// prove that mTLS is enabled.
func ProxyRequestsWithMissingPermission(cluster framework.Cluster, gateway string, opts ...client.CollectResponsesOptsFn) {
	const PermissionName = "allow-all-default"

	framework.Logf("deleting TrafficPermission %q", PermissionName)
	if cluster.GetKubectlOptions() == nil {
		Expect(cluster.GetKumactlOptions().KumactlDelete(
			"traffic-permission", PermissionName, "default"),
		).To(Succeed())
	} else {
		Expect(k8s.RunKubectlE(cluster.GetTesting(), cluster.GetKubectlOptions(),
			"delete", "trafficpermission", PermissionName),
		).To(Succeed())
	}

	framework.Logf("expecting 503 response from %q", gateway)
	Eventually(func(g Gomega) {
		target := fmt.Sprintf("http://%s/%s",
			gateway, path.Join("test", url.PathEscape(GinkgoT().Name())),
		)

		opts = append(opts, client.WithHeader("Host", "example.kuma.io"))
		status, err := client.CollectFailure(cluster, "gateway-client", target, opts...)

		g.Expect(err).To(Succeed())
		g.Expect(status.ResponseCode).To(Equal(503))
	}, "30s", "1s").Should(Succeed())
}

// ProxyRequestsWithRateLimit tests that requests to gateway are rate-limited with a 429 response.
func ProxyRequestsWithRateLimit(cluster framework.Cluster, gateway string, opts ...client.CollectResponsesOptsFn) {
	framework.Logf("expecting 429 response from %q", gateway)
	Eventually(func(g Gomega) {
		target := fmt.Sprintf("http://%s/%s",
			gateway, path.Join("test", url.PathEscape(GinkgoT().Name())),
		)

		opts = append(opts,
			client.NoFail(),
			client.OutputFormat(`{ "received": { "status": %{response_code} } }`),
			client.WithHeader("Host", "example.kuma.io"),
		)
		response, err := client.CollectResponse(cluster, "gateway-client", target, opts...)

		g.Expect(err).To(Succeed())
		g.Expect(response.Received.StatusCode).To(Equal(429))
	}, "30s", "1s").Should(Succeed())
}

// GatewayClientAppUniversal runs an empty container that will
// function as a client for a gateway.
func GatewayClientAppUniversal(name string) framework.InstallFunc {
	return func(cluster framework.Cluster) error {
		return cluster.DeployApp(
			framework.WithName(name),
			framework.WithoutDataplane(),
			framework.WithVerbose(),
		)
	}
}

func GatewayProxyUniversal(name string) framework.InstallFunc {
	return func(cluster framework.Cluster) error {
		token, err := cluster.GetKuma().GenerateDpToken("default", "edge-gateway")
		if err != nil {
			return err
		}

		dataplaneYaml := `
type: Dataplane
mesh: default
name: {{ name }}
networking:
  address:  {{ address }}
  gateway:
    type: BUILTIN
    tags:
      kuma.io/service: edge-gateway
`
		return cluster.DeployApp(
			framework.WithKumactlFlow(),
			framework.WithName(name),
			framework.WithToken(token),
			framework.WithVerbose(),
			framework.WithYaml(dataplaneYaml),
		)
	}
}

func EchoServerApp(name string, service string, instance string) framework.InstallFunc {
	return func(cluster framework.Cluster) error {
		token, err := cluster.GetKuma().GenerateDpToken("default", service)
		if err != nil {
			return err
		}

		return framework.TestServerUniversal(
			name,
			"default",
			token,
			framework.WithArgs([]string{"echo", "--instance", instance}),
			framework.WithServiceName(service),
		)(cluster)
	}
}
