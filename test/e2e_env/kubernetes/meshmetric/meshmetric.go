package meshmetric

import (
	"encoding/json"
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mads "github.com/kumahq/kuma/api/observability/v1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshmetric/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/otelcollector"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func BasicMeshMetricForMesh(policyName string, mesh string) InstallFunc {
	meshMetric := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    sidecar:
      profiles:
        appendProfiles:
          - name: All
    backends:
      - type: Prometheus
        prometheus: 
          port: 8080
          path: /metrics
          tls:
            mode: Disabled
`, policyName, Config.KumaNamespace, mesh)
	return YamlK8s(meshMetric)
}

func BasicMeshMetricWithProfileForMesh(policyName string, mesh string) InstallFunc {
	meshMetric := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    sidecar:
      includeUnused: true
      profiles:
        appendProfiles:
          - name: Basic
          - name: None # check merging
        exclude:
          - type: Regex
            match: "envoy_cluster_lb_.*"
        include:
          - type: Exact
            match: "envoy_cluster_default_total_match_count"
    backends:
      - type: Prometheus
        prometheus: 
          port: 8080
          path: /metrics
          tls:
            mode: Disabled
`, policyName, Config.KumaNamespace, mesh)
	return YamlK8s(meshMetric)
}

func MeshMetricMultiplePrometheusBackends(policyName string, mesh string, firstPrometheus string, secondPrometheus string) InstallFunc {
	meshMetric := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    sidecar:
      profiles:
        appendProfiles:
          - name: All
    backends:
      - type: Prometheus
        prometheus: 
          clientId: %s
          port: 8080
          path: /metrics
          tls:
            mode: Disabled
      - type: Prometheus
        prometheus: 
          clientId: %s
          port: 8081
          path: /metrics
          tls:
            mode: Disabled
`, policyName, Config.KumaNamespace, mesh, firstPrometheus, secondPrometheus)
	return YamlK8s(meshMetric)
}

func MeshMetricWithSpecificPrometheusClientId(policyName string, mesh string, clientId string) InstallFunc {
	meshMetric := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    sidecar:
      profiles:
        appendProfiles:
          - name: All
    backends:
      - type: Prometheus
        prometheus: 
          clientId: %s
          port: 8080
          path: /metrics
          tls:
            mode: Disabled
`, policyName, Config.KumaNamespace, mesh, clientId)
	return YamlK8s(meshMetric)
}

func MeshMetricWithSpecificPrometheusBackendForMeshService(mesh string, clientId string, serviceName string) InstallFunc {
	meshMetric := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: mesh-metric-2
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: %s
  default:
    sidecar:
      profiles:
        appendProfiles:
          - name: All
    backends:
      - type: Prometheus
        prometheus: 
          clientId: %s
          port: 8080
          path: /metrics
          tls:
            mode: Disabled
`, Config.KumaNamespace, mesh, serviceName, clientId)
	return YamlK8s(meshMetric)
}

func MeshMetricWithApplicationForMesh(policyName, mesh, appName, path string) InstallFunc {
	meshMetric := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: %s
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    sidecar:
      profiles:
        appendProfiles:
          - name: All
    applications:
      - name: "%s"
        path: "%s"
        port: 80
    backends:
      - type: Prometheus
        prometheus: 
          port: 8080
          path: /metrics
          tls:
            mode: Disabled
`, policyName, Config.KumaNamespace, mesh, appName, path)
	return YamlK8s(meshMetric)
}

func MeshMetricWithOpenTelemetryBackend(mesh, openTelemetryEndpoint string) InstallFunc {
	meshMetric := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: otel-metrics
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    sidecar:
      profiles:
        appendProfiles:
          - name: All
    backends:
      - type: OpenTelemetry
        openTelemetry: 
          endpoint: %s
          refreshInterval: 30s
`, Config.KumaNamespace, mesh, openTelemetryEndpoint)
	return YamlK8s(meshMetric)
}

func MeshMetricWithOpenTelemetryAndIncludeUnused(mesh, openTelemetryEndpoint string) InstallFunc {
	meshMetric := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: otel-metrics
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    sidecar:
      includeUnused: false
      profiles:
        appendProfiles:
          - name: All
    backends:
      - type: OpenTelemetry
        openTelemetry:
          endpoint: %s
          refreshInterval: 30s
`, Config.KumaNamespace, mesh, openTelemetryEndpoint)
	return YamlK8s(meshMetric)
}

func MeshMetricWithOpenTelemetryAndPrometheusBackend(mesh, openTelemetryEndpoint string) InstallFunc {
	meshMetric := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: otel-metrics
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    sidecar:
      profiles:
        appendProfiles:
          - name: All
    backends:
      - type: OpenTelemetry
        openTelemetry: 
          endpoint: %s
          refreshInterval: 30s
      - type: Prometheus
        prometheus: 
          port: 8080
          path: /metrics
          tls:
            mode: Disabled
`, Config.KumaNamespace, mesh, openTelemetryEndpoint)
	return YamlK8s(meshMetric)
}

func MeshMetricWithMultipleOpenTelemetryBackends(mesh, primaryOpenTelemetryEndpoint string, secondaryOpenTelemetryEndpoint string) InstallFunc {
	meshMetric := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshMetric
metadata:
  name: otel-metrics
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  default:
    sidecar:
      profiles:
        appendProfiles:
          - name: All
    backends:
      - type: OpenTelemetry
        openTelemetry: 
          endpoint: %s
          refreshInterval: 30s
      - type: OpenTelemetry
        openTelemetry:
          endpoint: %s
          refreshInterval: 30s
`, Config.KumaNamespace, mesh, primaryOpenTelemetryEndpoint, secondaryOpenTelemetryEndpoint)
	return YamlK8s(meshMetric)
}

func MeshMetric() {
	const (
		namespace                                = "meshmetric"
		mainMesh                                 = "main-metrics-mesh"
		mainPrometheusId                         = "main-prometheus"
		secondaryMesh                            = "secondary-metrics-mesh"
		secondaryPrometheusId                    = "secondary-prometheus"
		observabilityNamespace                   = "observability"
		secondaryOpenTelemetryCollectorNamespace = "secondary-otel"
		primaryOtelCollectorName                 = "otel-collector"
		secondaryOtelCollectorName               = "secondary-otel-collector"
	)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(mainMesh)).
			Install(MeshKubernetes(secondaryMesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(observabilityNamespace)).
			Install(Namespace(secondaryOpenTelemetryCollectorNamespace)).
			Install(Parallel(
				otelcollector.Install(
					otelcollector.WithName(primaryOtelCollectorName),
					otelcollector.WithNamespace(observabilityNamespace),
					otelcollector.WithIPv6(Config.IPV6),
				),
				democlient.Install(democlient.WithNamespace(observabilityNamespace)),
				otelcollector.Install(
					otelcollector.WithName(secondaryOtelCollectorName),
					otelcollector.WithNamespace(secondaryOpenTelemetryCollectorNamespace),
					otelcollector.WithIPv6(Config.IPV6),
				),
				democlient.Install(democlient.WithNamespace(secondaryOpenTelemetryCollectorNamespace)),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).To(Succeed())

		for i := 0; i < 2; i++ {
			Expect(
				kubernetes.Cluster.Install(testserver.Install(
					testserver.WithName(fmt.Sprintf("test-server-%d", i)),
					testserver.WithMesh(mainMesh),
					testserver.WithNamespace(namespace),
				)),
			).To(Succeed())
		}
		for i := 2; i < 4; i++ {
			Expect(
				kubernetes.Cluster.Install(testserver.Install(
					testserver.WithName(fmt.Sprintf("test-server-%d", i)),
					testserver.WithMesh(secondaryMesh),
					testserver.WithNamespace(namespace),
				)),
			).To(Succeed())
		}
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mainMesh, namespace, observabilityNamespace)
		DebugKube(kubernetes.Cluster, secondaryMesh, secondaryOpenTelemetryCollectorNamespace)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, mainMesh, v1alpha1.MeshMetricResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(kubernetes.Cluster, secondaryMesh, v1alpha1.MeshMetricResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(observabilityNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(secondaryOpenTelemetryCollectorNamespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mainMesh)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(secondaryMesh)).To(Succeed())
	})

	It("Basic MeshMetric policy exposes Envoy metrics on correct port", func() {
		// given
		Expect(kubernetes.Cluster.Install(BasicMeshMetricForMesh("mesh-policy", mainMesh))).To(Succeed())
		podIp, err := PodIPOfApp(kubernetes.Cluster, "test-server-0", namespace)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "test-server-0", "http://"+net.JoinHostPort(podIp, "8080")+"/metrics",
				client.FromKubernetesPod(namespace, "test-server-0"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(BeNil())
			// metric from envoy
			g.Expect(stdout).To(ContainSubstring("envoy_http_downstream_rq_xx"))
		}).Should(Succeed())
	})

	It("Basic MeshMetric policy with profiles exposes correct Envoy metrics", func() {
		// given
		Expect(kubernetes.Cluster.Install(BasicMeshMetricWithProfileForMesh("mesh-policy", mainMesh))).To(Succeed())
		podIp, err := PodIPOfApp(kubernetes.Cluster, "test-server-0", namespace)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "test-server-0", "http://"+net.JoinHostPort(podIp, "8080")+"/metrics",
				client.FromKubernetesPod(namespace, "test-server-0"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(BeNil())
			// metric from envoy
			g.Expect(stdout).To(ContainSubstring("envoy_cluster_upstream_cx_active"))        // from basic
			g.Expect(stdout).To(ContainSubstring("envoy_cluster_default_total_match_count")) // from include
			g.Expect(stdout).To(Not(ContainSubstring("envoy_cluster_lb_healthy_panic")))     // from exclude
		}).Should(Succeed())
	})

	It("MeshMetric policy with multiple Prometheus backends", func() {
		// given
		Expect(kubernetes.Cluster.Install(MeshMetricMultiplePrometheusBackends("mesh-policy", mainMesh, mainPrometheusId, secondaryPrometheusId))).To(Succeed())
		podIp, err := PodIPOfApp(kubernetes.Cluster, "test-server-0", namespace)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			// main Prometheus backend
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "test-server-0", "http://"+net.JoinHostPort(podIp, "8080")+"/metrics",
				client.FromKubernetesPod(namespace, "test-server-0"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(BeNil())
			g.Expect(stdout).To(ContainSubstring("envoy_http_downstream_rq_xx"))

			// secondary Prometheus backend
			stdout, _, err = client.CollectResponse(
				kubernetes.Cluster, "test-server-0", "http://"+net.JoinHostPort(podIp, "8081")+"/metrics",
				client.FromKubernetesPod(namespace, "test-server-0"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(BeNil())
			// metric from envoy
			g.Expect(stdout).To(ContainSubstring("envoy_http_downstream_rq_xx"))
		}).Should(Succeed())
	})

	It("MeshMetric policy with dynamic configuration and application aggregation correctly exposes aggregated metrics", func() {
		// given
		Expect(kubernetes.Cluster.Install(MeshMetricWithApplicationForMesh("dynamic-config", mainMesh, "origin-app", "/metrics"))).To(Succeed())
		podIp, err := PodIPOfApp(kubernetes.Cluster, "test-server-0", namespace)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "test-server-0", "http://"+net.JoinHostPort(podIp, "8080")+"/metrics",
				client.FromKubernetesPod(namespace, "test-server-0"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(BeNil())
			// metric from envoy
			g.Expect(stdout).To(ContainSubstring("envoy_http_downstream_rq_xx"))
			g.Expect(stdout).To(ContainSubstring("origin-app"))
		}, "1m", "1s").Should(Succeed())

		// update policy config and check if changes was applied on DPP
		Expect(kubernetes.Cluster.Install(MeshMetricWithApplicationForMesh("dynamic-config", mainMesh, "updated-app", "/metrics"))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "test-server-0", "http://"+net.JoinHostPort(podIp, "8080")+"/metrics",
				client.FromKubernetesPod(namespace, "test-server-0"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(BeNil())
			// metric from envoy
			g.Expect(stdout).To(ContainSubstring("envoy_http_downstream_rq_xx"))
			g.Expect(stdout).ToNot(ContainSubstring("origin-app"))
			g.Expect(stdout).To(ContainSubstring("updated-app"))
		}, "1m", "1s").Should(Succeed())
	})

	It("MADS server response contains DPPs from all meshes when prometheus client id is empty", func() {
		// given
		Expect(kubernetes.Cluster.Install(BasicMeshMetricForMesh("main-mesh-policy", mainMesh))).To(Succeed())
		Expect(kubernetes.Cluster.Install(BasicMeshMetricForMesh("secondary-mesh-policy", secondaryMesh))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			assignment, err := kubernetes.Cluster.GetKuma().GetMonitoringAssignment(mainPrometheusId)
			g.Expect(err).ToNot(HaveOccurred())

			madsResponse := MonitoringAssignmentResponse{}
			g.Expect(json.Unmarshal([]byte(assignment), &madsResponse)).To(Succeed())
			// all DPPs from both meshes in single MADS response
			g.Expect(getServicesFrom(madsResponse)).To(ConsistOf(
				"test-server-0_meshmetric_svc_80", "test-server-1_meshmetric_svc_80", "test-server-2_meshmetric_svc_80", "test-server-3_meshmetric_svc_80",
			))
		}).Should(Succeed())

		// and same response for secondary backend
		Eventually(func(g Gomega) {
			assignment, err := kubernetes.Cluster.GetKuma().GetMonitoringAssignment(secondaryPrometheusId)
			g.Expect(err).ToNot(HaveOccurred())

			madsResponse := MonitoringAssignmentResponse{}
			g.Expect(json.Unmarshal([]byte(assignment), &madsResponse)).To(Succeed())
			// all DPPs from both meshes in single MADS response
			g.Expect(getServicesFrom(madsResponse)).To(ConsistOf(
				"test-server-0_meshmetric_svc_80", "test-server-1_meshmetric_svc_80", "test-server-2_meshmetric_svc_80", "test-server-3_meshmetric_svc_80",
			))
		}).Should(Succeed())
	})

	It("MADS server response contains DPPs from corresponding meshes when prometheus client id is set", func() {
		// given
		Expect(kubernetes.Cluster.Install(MeshMetricWithSpecificPrometheusClientId("main-mesh-policy", mainMesh, mainPrometheusId))).To(Succeed())
		Expect(kubernetes.Cluster.Install(MeshMetricWithSpecificPrometheusClientId("secondary-mesh-policy", secondaryMesh, secondaryPrometheusId))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			assignment, err := kubernetes.Cluster.GetKuma().GetMonitoringAssignment(mainPrometheusId)
			g.Expect(err).ToNot(HaveOccurred())

			madsResponse := MonitoringAssignmentResponse{}
			g.Expect(json.Unmarshal([]byte(assignment), &madsResponse)).To(Succeed())
			// all DPPs from primaryMesh for primary Prometheus backend
			g.Expect(getServicesFrom(madsResponse)).To(ConsistOf(
				"test-server-0_meshmetric_svc_80", "test-server-1_meshmetric_svc_80",
			))
		}).Should(Succeed())

		// and
		Eventually(func(g Gomega) {
			assignment, err := kubernetes.Cluster.GetKuma().GetMonitoringAssignment(secondaryPrometheusId)
			g.Expect(err).ToNot(HaveOccurred())

			madsResponse := MonitoringAssignmentResponse{}
			g.Expect(json.Unmarshal([]byte(assignment), &madsResponse)).To(Succeed())
			// all DPPs from secondaryMesh for secondary Prometheus backend
			g.Expect(getServicesFrom(madsResponse)).To(ConsistOf(
				"test-server-2_meshmetric_svc_80", "test-server-3_meshmetric_svc_80",
			))
		}).Should(Succeed())
	})

	It("override MADS response for single DPP in mesh", func() {
		// given
		Expect(kubernetes.Cluster.Install(MeshMetricWithSpecificPrometheusClientId("main-mesh-policy", mainMesh, mainPrometheusId))).To(Succeed())
		Expect(kubernetes.Cluster.Install(MeshMetricWithSpecificPrometheusBackendForMeshService(mainMesh, secondaryPrometheusId, "test-server-1_meshmetric_svc_80"))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			assignment, err := kubernetes.Cluster.GetKuma().GetMonitoringAssignment(mainPrometheusId)
			g.Expect(err).ToNot(HaveOccurred())

			madsResponse := MonitoringAssignmentResponse{}
			g.Expect(json.Unmarshal([]byte(assignment), &madsResponse)).To(Succeed())
			// two DPPs configured by Mesh targetRef
			g.Expect(getServicesFrom(madsResponse)).To(ConsistOf("test-server-0_meshmetric_svc_80"))
		}).Should(Succeed())

		// and
		Eventually(func(g Gomega) {
			assignment, err := kubernetes.Cluster.GetKuma().GetMonitoringAssignment(secondaryPrometheusId)
			g.Expect(err).ToNot(HaveOccurred())

			madsResponse := MonitoringAssignmentResponse{}
			g.Expect(json.Unmarshal([]byte(assignment), &madsResponse)).To(Succeed())
			// single DPP overridden by MeshService targetRef
			g.Expect(getServicesFrom(madsResponse)).To(ConsistOf("test-server-1_meshmetric_svc_80"))
		}).Should(Succeed())
	})

	XIt("MeshMetric with OpenTelemetry enabled", func() {
		// given
		openTelemetryCollector := otelcollector.From(kubernetes.Cluster, primaryOtelCollectorName)
		Expect(kubernetes.Cluster.Install(MeshMetricWithOpenTelemetryBackend(mainMesh, openTelemetryCollector.CollectorEndpoint()))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "demo-client", openTelemetryCollector.ExporterEndpoint(),
				client.FromKubernetesPod(observabilityNamespace, "demo-client"),
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("envoy_cluster_external_upstream_rq_time_bucket"))
		}, "3m", "5s").Should(Succeed())
	})

	XIt("MeshMetric with OpenTelemetry and usedonly/filter", func() {
		// given
		openTelemetryCollector := otelcollector.From(kubernetes.Cluster, primaryOtelCollectorName)
		Expect(kubernetes.Cluster.Install(MeshMetricWithOpenTelemetryAndIncludeUnused(mainMesh, openTelemetryCollector.CollectorEndpoint()))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "demo-client", openTelemetryCollector.ExporterEndpoint(),
				client.FromKubernetesPod(observabilityNamespace, "demo-client"),
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(Not(ContainSubstring("envoy_cluster_client_ssl_socket_factory_upstream_context_secrets_not_ready"))) // unused
			g.Expect(stdout).To(ContainSubstring("envoy_cluster_external_upstream_rq_time_bucket"))                                  // used
		}, "3m", "5s").Should(Succeed())
	})

	XIt("MeshMetric with OpenTelemetry and Prometheus enabled", func() {
		// given
		openTelemetryCollector := otelcollector.From(kubernetes.Cluster, primaryOtelCollectorName)
		testServerIp, err := PodIPOfApp(kubernetes.Cluster, "test-server-0", namespace)
		Expect(err).ToNot(HaveOccurred())
		Expect(kubernetes.Cluster.Install(MeshMetricWithOpenTelemetryAndPrometheusBackend(mainMesh, openTelemetryCollector.CollectorEndpoint()))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			// metrics from OpenTelemetry
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "demo-client", openTelemetryCollector.ExporterEndpoint(),
				client.FromKubernetesPod(observabilityNamespace, "demo-client"),
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("envoy_cluster_external_upstream_rq_time_bucket"))

			// metrics from Prometheus
			stdout, _, err = client.CollectResponse(
				kubernetes.Cluster, "test-server-0", "http://"+net.JoinHostPort(testServerIp, "8080")+"/metrics",
				client.FromKubernetesPod(namespace, "test-server-0"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).ToNot(BeNil())
			g.Expect(stdout).To(ContainSubstring("envoy_http_downstream_rq_xx"))
		}, "3m", "5s").Should(Succeed())
	})

	XIt("MeshMetric with multiple OpenTelemetry backends", func() {
		// given
		primaryOpenTelemetryCollector := otelcollector.From(kubernetes.Cluster, primaryOtelCollectorName)
		secondaryOpenTelemetryCollector := otelcollector.From(kubernetes.Cluster, secondaryOtelCollectorName)
		Expect(kubernetes.Cluster.Install(MeshMetricWithMultipleOpenTelemetryBackends(mainMesh, primaryOpenTelemetryCollector.CollectorEndpoint(), secondaryOpenTelemetryCollector.CollectorEndpoint()))).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			// primary collector
			stdout, _, err := client.CollectResponse(
				kubernetes.Cluster, "demo-client", primaryOpenTelemetryCollector.ExporterEndpoint(),
				client.FromKubernetesPod(observabilityNamespace, "demo-client"),
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("envoy_cluster_external_upstream_rq_time_bucket"))

			// secondary collector
			stdout, _, err = client.CollectResponse(
				kubernetes.Cluster, "demo-client", secondaryOpenTelemetryCollector.ExporterEndpoint(),
				client.FromKubernetesPod(secondaryOpenTelemetryCollectorNamespace, "demo-client"),
				client.WithVerbose(),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("envoy_cluster_external_upstream_rq_time_bucket"))
		}, "3m", "5s").Should(Succeed())
	})
}

func getServicesFrom(response MonitoringAssignmentResponse) []string {
	var services []string
	for _, assignment := range response.Resources {
		services = append(services, assignment.Service)
	}
	return services
}

type MonitoringAssignmentResponse struct {
	Resources []*mads.MonitoringAssignment `json:"resources"`
}
