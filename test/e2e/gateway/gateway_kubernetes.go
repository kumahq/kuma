package gateway

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"text/template"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func GatewayOnKubernetes() {
	var cluster *K8sCluster

	// ClientNamespace is a namespace to deploy gateway client
	// applications. Mesh sidecar injection is not enabled there.
	const ClientNamespace = "kuma-client"

	const GatewayPort = "8080"

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
			name, Config.GetUniversalImage(), ClientNamespace,
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

	// DeployCluster creates a Kuma cluster on Kubernetes using the
	// provided options, installing an echo service as well as a
	// gateway and a client container to send HTTP requests.
	DeployCluster := func(opt ...KumaDeploymentOption) {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		opt = append(opt, WithVerbose(), WithCtlOpts(map[string]string{"--experimental-meshgateway": "true"}))

		err := NewClusterSetup().
			Install(Kuma(config_core.Standalone, opt...)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(Namespace(ClientNamespace)).
			Install(EchoServerApp("echo-server")).
			Install(GatewayClientApp("gateway-client")).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	}

	// GatewayAddress find the address of the gateway instance service named instanceName.
	GatewayAddress := func(instanceName string) string {
		services, err := k8s.ListServicesE(cluster.GetTesting(), cluster.GetKubectlOptions(TestNamespace), metav1.ListOptions{})
		Expect(err).To(Succeed())

		// Find the service that is owned by the named GatewayInstance.
		for _, svc := range services {
			for _, ref := range svc.GetOwnerReferences() {
				if ref.Kind == "MeshGatewayInstance" && ref.Name == instanceName {
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
kind: MeshGateway
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
kind: MeshGatewayRoute
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
			cluster.GetKumactlOptions().KumactlList("meshgateways", "default"),
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
kind: MeshGatewayInstance
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

	Context("when mTLS is disabled", func() {
		BeforeEach(func() {
			DeployCluster()
		})

		It("should proxy simple HTTP requests", func() {
			ProxySimpleRequests(cluster, "kubernetes",
				net.JoinHostPort(GatewayAddress("edge-gateway"), GatewayPort),
				client.FromKubernetesPod(ClientNamespace, "gateway-client"))
		})
	})

	Context("when mTLS is enabled", func() {
		BeforeEach(func() {
			DeployCluster(OptEnableMeshMTLS)
		})

		It("should proxy simple HTTP requests", func() {
			ProxySimpleRequests(cluster, "kubernetes",
				net.JoinHostPort(GatewayAddress("edge-gateway"), GatewayPort),
				client.FromKubernetesPod(ClientNamespace, "gateway-client"))
		})

		// In mTLS mode, only the presence of TrafficPermission rules allow services to receive
		// traffic, so removing the permission should cause requests to fail. We use this to
		// prove that mTLS is enabled
		It("should fail without TrafficPermission", func() {
			ProxyRequestsWithMissingPermission(cluster,
				net.JoinHostPort(GatewayAddress("edge-gateway"), GatewayPort),
				client.FromKubernetesPod(ClientNamespace, "gateway-client"))
		})
	})

	Context("when targeting an external service", func() {
		BeforeEach(func() {
			DeployCluster()
			err := NewClusterSetup().
				Install(testserver.Install(
					testserver.WithName("es-test-server"),
					testserver.WithNamespace(ClientNamespace), // reuse client namespace to not wait for delete separate namespace
					testserver.WithArgs("echo", "--instance", "es-test-server"),
				)).
				Install(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
mesh: default
metadata:
  name: external-service
spec:
  tags:
    kuma.io/service: external-service
    kuma.io/protocol: http
  networking:
    address: es-test-server.kuma-client.svc.cluster.local:80`)).
				Install(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: MeshGatewayRoute
metadata:
  name: edge-gateway-external-service
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
            value: /external
        backends:
        - destination:
            kuma.io/service: external-service
`)).
				Setup(cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should proxy simple HTTP requests", func() {
			ProxySimpleRequests(cluster, "es-test-server",
				net.JoinHostPort(GatewayAddress("edge-gateway"), GatewayPort),
				client.FromKubernetesPod(ClientNamespace, "gateway-client"),
				client.WithPathPrefix("/external"),
			)
		})
	})

	Context("when targeting a HTTPS gateway", func() {
		BeforeEach(func() {
			DeployCluster()
		})

		JustBeforeEach(func() {
			cert, key, err := CreateCertsFor("example.kuma.io")
			Expect(err).To(Succeed())
			secretData := base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{key, cert}, "\n")))

			err = NewClusterSetup().
				Install(YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
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
      protocol: HTTPS
      hostname: example.kuma.io
      tls:
        mode: TERMINATE
        certificates:
        - secret: example-kuma-io-certificate
      tags:
        hostname: example.kuma.io
`)).
				Install(YamlK8s(fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: example-kuma-io-certificate
  namespace: %s
  labels:
    kuma.io/mesh: default 
data:
  value: %s
type: system.kuma.io/secret
`, Config.KumaNamespace, secretData))).
				Setup(cluster)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should proxy simple HTTPS requests", func() {
			ProxySecureRequests(cluster, "kubernetes",
				net.JoinHostPort("example.kuma.io", GatewayPort),
				client.Resolve("example.kuma.io", 8080, GatewayAddress("edge-gateway")),
				client.FromKubernetesPod(ClientNamespace, "gateway-client"),
			)
		})
	})

	Context("when a rate limit is configured", func() {
		BeforeEach(func() {
			DeployCluster()
		})

		JustBeforeEach(func() {
			Expect(k8s.KubectlApplyFromStringE(
				cluster.GetTesting(),
				cluster.GetKubectlOptions(), `
apiVersion: kuma.io/v1alpha1
kind: RateLimit
metadata:
  name: echo-rate-limit
mesh: default
spec:
  sources:
  - match:
      kuma.io/service: edge-gateway
  destinations:
  - match:
      kuma.io/service: echo-server_kuma-test_svc_80 # Matches the echo-server we deployed.
  conf:
    http:
      requests: 5
      interval: 10s
`),
			).To(Succeed())
		})

		It("should be rate limited", func() {
			ProxyRequestsWithRateLimit(cluster,
				net.JoinHostPort(GatewayAddress("edge-gateway"), GatewayPort),
				client.FromKubernetesPod(ClientNamespace, "gateway-client"))
		})
	})
}
