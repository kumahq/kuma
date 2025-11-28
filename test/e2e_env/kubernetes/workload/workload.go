package workload

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v2/test/framework/envs/kubernetes"
)

func Workload() {
	const namespace = "workloads-ns"
	const mesh = "workloads"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(MeshKubernetes(mesh)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mesh, namespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	It("should use service account as workload when no workload labels configured", func() {
		// given
		const appName = "test-server-a"
		const serviceAccount = "test-server-sa"

		// when
		err := NewClusterSetup().
			Install(testserver.Install(
				testserver.WithName(appName),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithServiceAccount(serviceAccount),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then verify pod is created
		var podName string
		Eventually(func(g Gomega) {
			podName, err = PodNameOfApp(kubernetes.Cluster, appName, namespace)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())

		// and verify dataplane has workload annotation set to service account
		Eventually(func(g Gomega) {
			dpName := fmt.Sprintf("%s.%s", podName, namespace)
			dpYAML, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput(
				"get", "dataplane", dpName, "--mesh", mesh, "-oyaml",
			)
			g.Expect(err).ToNot(HaveOccurred(), "failed to get dataplane '%s'", dpName)

			dpRes, err := rest.YAML.UnmarshalCore([]byte(dpYAML))
			g.Expect(err).ToNot(HaveOccurred())

			dp, ok := dpRes.(*core_mesh.DataplaneResource)
			g.Expect(ok).To(BeTrue())

			// verify workload label is set to service account name
			// (stored as annotation in k8s, accessible through labels API)
			workloadLabel := dp.Meta.GetLabels()[metadata.KumaWorkload]
			g.Expect(workloadLabel).To(Equal(serviceAccount), "workload should equal service account name")
		}).Should(Succeed())
	})

	It("should use pod label as workload when workload labels configured", func() {
		// given
		const appName = "test-server-b"
		const appLabel = "test-server-app"

		// when deploy with pod labels
		err := NewClusterSetup().
			Install(testserver.Install(
				testserver.WithName(appName),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithPodLabels(map[string]string{
					"app.kubernetes.io/name": appLabel,
				}),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then verify pod is created
		var podName string
		Eventually(func(g Gomega) {
			podName, err = PodNameOfApp(kubernetes.Cluster, appName, namespace)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())

		// and verify dataplane has workload annotation set to "test-server" (from app.kubernetes.io/name label)
		Eventually(func(g Gomega) {
			dpName := fmt.Sprintf("%s.%s", podName, namespace)
			dpYAML, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput(
				"get", "dataplane", dpName, "--mesh", mesh, "-oyaml",
			)
			g.Expect(err).ToNot(HaveOccurred(), "failed to get dataplane '%s'", dpName)

			dpRes, err := rest.YAML.UnmarshalCore([]byte(dpYAML))
			g.Expect(err).ToNot(HaveOccurred())

			dp, ok := dpRes.(*core_mesh.DataplaneResource)
			g.Expect(ok).To(BeTrue())

			// verify workload is set from app label, not service account
			workloadLabel := dp.Meta.GetLabels()[metadata.KumaWorkload]
			g.Expect(workloadLabel).To(Equal(appLabel), "workload should equal app label value")
		}).Should(Succeed())
	})

	It("should reject pod creation with manual kuma.io/workload label", func() {
		// given pod YAML with manual kuma.io/workload label
		podYAML := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: manual-workload-pod
  namespace: %s
  labels:
    app: test
    kuma.io/mesh: %s
    kuma.io/sidecar-injection: enabled
    kuma.io/workload: manually-set-workload
spec:
  containers:
  - name: app
    image: nginx
    ports:
    - containerPort: 80
`, namespace, mesh)

		// when trying to create pod with manual workload label
		err := k8s.KubectlApplyFromStringE(
			kubernetes.Cluster.GetTesting(),
			kubernetes.Cluster.GetKubectlOptions(namespace),
			podYAML,
		)

		// then webhook should deny the creation
		Expect(err).To(HaveOccurred(), "pod with manual kuma.io/workload label should be rejected")
		Expect(err.Error()).To(ContainSubstring("cannot manually set kuma.io/workload label"))
	})

	It("should automatically create and delete Workload resource", func() {
		// given
		const appName = "workload-resource-test"
		const workloadName = "test-workload-resource"

		// when deploy test server with workload label
		err := NewClusterSetup().
			Install(testserver.Install(
				testserver.WithName(appName),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithPodLabels(map[string]string{
					"app.kubernetes.io/name": workloadName,
				}),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then verify pod is created
		Eventually(func(g Gomega) {
			_, err = PodNameOfApp(kubernetes.Cluster, appName, namespace)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())

		// and verify Workload resource is created
		Eventually(func(g Gomega) {
			// verify Workload resource exists and has correct content
			workloadK8sName := fmt.Sprintf("%s.%s", workloadName, namespace)
			workload, err := GetWorkload(kubernetes.Cluster, workloadK8sName, mesh)
			g.Expect(err).ToNot(HaveOccurred(), "failed to get workload '%s'", workloadK8sName)
			g.Expect(workload.GetMeta().GetName()).To(Equal(workloadK8sName))
			g.Expect(workload.GetMeta().GetMesh()).To(Equal(mesh))
		}).Should(Succeed())

		// when delete the deployment
		err = k8s.RunKubectlE(
			kubernetes.Cluster.GetTesting(),
			kubernetes.Cluster.GetKubectlOptions(namespace),
			"delete", "deployment", appName,
		)
		Expect(err).ToNot(HaveOccurred())

		// then verify Workload resource is deleted
		Eventually(func(g Gomega) {
			workloadK8sName := fmt.Sprintf("%s.%s", workloadName, namespace)
			_, err := GetWorkload(kubernetes.Cluster, workloadK8sName, mesh)
			g.Expect(err).To(HaveOccurred(), "workload should be deleted")
			g.Expect(err.Error()).To(ContainSubstring("No resources found in workloads mesh"))
		}, "2m").Should(Succeed())
	})

	FIt("should update Workload status as dataplanes are added and removed", func() {
		// given
		const workloadName = "status-test-workload"
		const appName1 = "test-server-1"
		const appName2 = "test-server-2"

		// when deploy first test server
		err := NewClusterSetup().
			Install(testserver.Install(
				testserver.WithName(appName1),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithPodLabels(map[string]string{
					"app.kubernetes.io/name": workloadName,
				}),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then verify Workload status shows 1 dataplane
		workloadK8sName := fmt.Sprintf("%s.%s", workloadName, namespace)
		Eventually(func(g Gomega) {
			workload, err := GetWorkload(kubernetes.Cluster, workloadK8sName, mesh)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(workload.Status.DataplaneProxies.Total).To(Equal(int32(1)))
		}, "2m").Should(Succeed())

		// when deploy second test server
		err = NewClusterSetup().
			Install(testserver.Install(
				testserver.WithName(appName2),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithPodLabels(map[string]string{
					"app.kubernetes.io/name": workloadName,
				}),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// then verify Workload status shows 2 dataplanes
		Eventually(func(g Gomega) {
			workload, err := GetWorkload(kubernetes.Cluster, workloadK8sName, mesh)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(workload.Status.DataplaneProxies.Total).To(Equal(int32(2)))
		}, "2m").Should(Succeed())

		// when delete first test server
		err = k8s.RunKubectlE(
			kubernetes.Cluster.GetTesting(),
			kubernetes.Cluster.GetKubectlOptions(namespace),
			"delete", "deployment", appName1,
		)
		Expect(err).ToNot(HaveOccurred())

		// then verify Workload status shows 1 dataplane again
		Eventually(func(g Gomega) {
			workload, err := GetWorkload(kubernetes.Cluster, workloadK8sName, mesh)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(workload.Status.DataplaneProxies.Total).To(Equal(int32(1)))
		}, "2m").Should(Succeed())
	})
}
