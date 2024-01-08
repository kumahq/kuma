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

func MeshMetric() {
	const (
		namespace             = "meshmetric"
		mainMesh              = "main-metrics-mesh"
		mainPrometheusId      = "main-prometheus"
		secondaryMesh         = "secondary-metrics-mesh"
		secondaryPrometheusId = "secondary-prometheus"
	)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(mainMesh)).
			Install(MeshKubernetes(secondaryMesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
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

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, mainMesh, v1alpha1.MeshMetricResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(kubernetes.Cluster, secondaryMesh, v1alpha1.MeshMetricResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
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
}

func getServicesFrom(response MonitoringAssignmentResponse) []string {
	var services []string
	for _, assignment := range response.Resources {
		services = append(services, assignment.Service)
	}
	return services
}

type MonitoringAssignmentResponse struct {
	Resources []mads.MonitoringAssignment `json:"resources"`
}
