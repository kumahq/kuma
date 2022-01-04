package gateway

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"
	"text/template"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/test/e2e/trafficroute/testutil"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func GatewayOnKubernetes() {
	var cluster *K8sCluster

	// ClientNamespace is a namespace to deploy gateway client
	// applications. Mesh sidecar injection is not enabled there.
	const ClientNamespace = "kuma-client"

	EchoServerApp := func(name string) InstallFunc {
		return testserver.Install(
			testserver.WithMesh("default"),
			testserver.WithName(name),
			testserver.WithArgs("echo", "--instance", "kubernetes"),
		)
	}

	// GatewayClientApp runs an empty container that will
	// function as a client for a gateway.
	GatewayClientApp := func(name string) InstallFunc {
		args := struct {
			Name      string
			Image     string
			Namespace string
		}{
			name, GetUniversalImage(), ClientNamespace,
		}

		out := &bytes.Buffer{}

		tmpl := template.Must(
			template.New("deployment").Parse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels:
    app: {{.Name}}
spec:
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
    spec:
      containers:
        - name: {{.Name}}
          image: {{.Image}}
          imagePullPolicy: IfNotPresent
          command: [ "sleep" ]
          args: [ "infinity" ]
`))
		Expect(tmpl.Execute(out, &args)).To(Succeed())

		return Combine(
			YamlK8s(out.String()),
			WaitPodsAvailable(ClientNamespace, name),
		)
	}

	SetupCluster := func(setup *ClusterSetup) {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		Expect(setup.Setup(cluster)).To(Succeed())

		out, err := k8s.RunKubectlAndGetOutputE(cluster.GetTesting(), cluster.GetKubectlOptions(),
			"api-resources", "--api-group=kuma.io", "--output=name")
		Expect(err).To(Succeed())

		// If the GatewayInstances CRD is installed, we can assume
		// that kuma-cp is built with support for builtin Gateways.
		// It's not a perfect test (the kumactl and kuma-cp builds
		// could be out of sync), but it ought to be reliable in CI.
		if !strings.Contains(out, "gatewayinstances.kuma.io") {
			Skip("kuma-cp builtin Gateway support is not enabled")
		}
	}

	// DeployCluster creates a Kuma cluster on Kubernetes using the
	// provided options, installing an echo service as well as a
	// gateway and a client container to send HTTP requests.
	DeployCluster := func(opt ...KumaDeploymentOption) {
		opt = append(opt, WithVerbose())

		SetupCluster(NewClusterSetup().
			Install(Kuma(config_core.Standalone, opt...)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(Namespace(ClientNamespace)).
			Install(EchoServerApp("echo-server")).
			Install(GatewayClientApp("gateway-client")),
		)
	}

	// GatewayInstanceAddress find the address of the gateway instance service.
	GatewayInstanceAddress := func(instanceName string) string {
		services, err := k8s.ListServicesE(cluster.GetTesting(), cluster.GetKubectlOptions(TestNamespace), metav1.ListOptions{})
		Expect(err).To(Succeed())

		// Find the service that is owned by the named GatewayInstance.
		for _, svc := range services {
			for _, ref := range svc.GetOwnerReferences() {
				if ref.Kind == "GatewayInstance" && ref.Name == instanceName {
					return svc.Spec.ClusterIP
				}
			}
		}

		return "0.0.0.0"
	}

	// Before each test, install the gateway and routes.
	JustBeforeEach(func() {
		Expect(k8s.KubectlApplyFromStringE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(TestNamespace), `
apiVersion: kuma.io/v1alpha1
kind: Gateway
metadata:
  name: edge-gateway
mesh: default
spec:
  selectors:
  - match:
      kuma.io/service: edge-gateway
  conf:
    listeners:
    - port: 8080
      protocol: HTTP
      hostname: example.kuma.io
      tags:
        hostname: example.kuma.io
`),
		).To(Succeed())

		Expect(k8s.KubectlApplyFromStringE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(TestNamespace), `
apiVersion: kuma.io/v1alpha1
kind: GatewayRoute
metadata:
  name: edge-gateway
mesh: default
spec:
  selectors:
  - match:
      kuma.io/service: edge-gateway
  conf:
    http:
      rules:
      - matches:
        - path:
            match: PREFIX
            value: /
        backends:
        - destination:
            kuma.io/service: echo-server_kuma-test_svc_80 # Matches the echo-server we deployed.
`),
		).To(Succeed())

		Expect(
			cluster.GetKumactlOptions().KumactlList("gateways", "default"),
		).To(ContainElement("edge-gateway"))
	})

	// Before each test, install the GatewayInstance to provision
	// dataplanes. Note that we expose the gateway inside the cluster
	// with a ClusterIP service. This makes it easier for the tests
	// to figure out the IP address to send requests to, since the
	// alternatives are a load balancer (we don't have one) or node port
	// (would need to inspect nodes).
	JustBeforeEach(func() {
		Expect(k8s.KubectlApplyFromStringE(
			cluster.GetTesting(),
			cluster.GetKubectlOptions(TestNamespace), `
apiVersion: kuma.io/v1alpha1
kind: GatewayInstance
metadata:
  name: edge-gateway
spec:
  replicas: 1
  serviceType: ClusterIP
  tags:
    kuma.io/service: edge-gateway
`),
		).To(Succeed())
	})

	// Before each test, verify that we have the Dataplanes that we expect to need.
	JustBeforeEach(func() {
		Expect(cluster.VerifyKuma()).To(Succeed())

		// Synchronize on the dataplanes coming up.
		Eventually(func(g Gomega) {
			dataplanes, err := cluster.GetKumactlOptions().KumactlList("dataplanes", "default")
			g.Expect(err).ToNot(HaveOccurred())
			// Dataplane names are generated, so we check for a partial match.
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring("echo-server")))
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring("edge-gateway")))
		}, "60s", "1s").Should(Succeed())
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteNamespace(ClientNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	// ProxySimpleRequests tests that basic HTTP requests are proxied to the echo-server.
	ProxySimpleRequests := func(prefix string, instance string) func() {
		return func() {
			Eventually(func(g Gomega) {
				target := fmt.Sprintf("http://%s/%s",
					net.JoinHostPort(GatewayInstanceAddress("edge-gateway"), "8080"),
					path.Join(prefix, "test", url.PathEscape(GinkgoT().Name())),
				)

				response, err := testutil.CollectResponse(
					cluster, "gateway-client", target,
					testutil.FromKubernetesPod(ClientNamespace, "gateway-client"),
					testutil.WithHeader("Host", "example.kuma.io"),
				)

				g.Expect(err).To(Succeed())
				g.Expect(response.Instance).To(Equal(instance))
				g.Expect(response.Received.Headers["Host"]).To(ContainElement("example.kuma.io"))
			}, "60s", "1s").Should(Succeed())
		}
	}

	Context("when MTLS is disabled", func() {
		BeforeEach(func() {
			DeployCluster(KumaK8sDeployOpts...)
		})

		It("should proxy simple HTTP requests", ProxySimpleRequests("/", "kubernetes"))
	})

	Context("when mTLS is enabled", func() {
		BeforeEach(func() {
			mtls := WithMeshUpdate("default", func(mesh *mesh_proto.Mesh) *mesh_proto.Mesh {
				mesh.Mtls = &mesh_proto.Mesh_Mtls{
					EnabledBackend: "builtin",
					Backends: []*mesh_proto.CertificateAuthorityBackend{
						{Name: "builtin", Type: "builtin"},
					},
				}
				return mesh
			})

			DeployCluster(append(KumaK8sDeployOpts, mtls)...)
		})

		It("should proxy simple HTTP requests", ProxySimpleRequests("/", "kubernetes"))
	})
}
