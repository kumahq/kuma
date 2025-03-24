package container_patch

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"

	k8s_util "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func ContainerPatch() {
	const namespace = "container-patch"
	const mesh = "container-patch"
	const appName = "test-service"
	const appNameWithPatch = "test-service-patched"

	containerPatch := func(ns string) string {
		return fmt.Sprintf(`apiVersion: kuma.io/v1alpha1
kind: ContainerPatch
metadata:
  namespace: %s
  name: container-patch-1
spec:
  sidecarPatch:
    - op: add
      path: /securityContext/privileged
      value: "true"`, ns)
	}
	containerPatch2 := func(ns string) string {
		return fmt.Sprintf(`apiVersion: kuma.io/v1alpha1
kind: ContainerPatch
metadata:
  namespace: %s
  name: container-patch-2
spec:
  initPatch:
    - op: remove
      path: /securityContext/runAsUser`, ns)
	}
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(YamlK8s(containerPatch(Config.KumaNamespace))).
			Install(YamlK8s(containerPatch2(Config.KumaNamespace))).
			Install(MeshKubernetes(mesh)).
			Install(Parallel(
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(mesh),
					testserver.WithName(appNameWithPatch),
					testserver.WithPodAnnotations(
						map[string]string{"kuma.io/container-patches": "container-patch-1,container-patch-2"},
					),
				),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithMesh(mesh),
					testserver.WithName(appName),
				),
			)).
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

	It("should apply container patch to kubernetes configuration", func() {
		// when
		// pod without container patch
		podName, err := PodNameOfApp(kubernetes.Cluster, appName, namespace)
		Expect(err).ToNot(HaveOccurred())
		pod, err := k8s.GetPodE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), podName)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(pod.Spec.InitContainers).To(
			Or(HaveLen(1), HaveLen(2)),
		)
		// should have default value *int64 = 0
		Expect(pod.Spec.InitContainers[0].SecurityContext.RunAsUser).To(Equal(new(int64)))
		Expect(pod.Spec.Containers).To(
			Or(HaveLen(2), HaveLen(1)),
		)
		beSidecarWithoutPrivileged := And(
			WithTransform(func(c kube_core.Container) string { return c.Name }, BeEquivalentTo(k8s_util.KumaSidecarContainerName)),
			WithTransform(func(c kube_core.Container) *bool { return c.SecurityContext.Privileged }, BeNil()),
		)
		if len(pod.Spec.Containers) == 2 {
			// kuma-sidecar is the first container
			Expect(pod.Spec.Containers[0]).To(beSidecarWithoutPrivileged)
		} else {
			Expect(pod.Spec.InitContainers).To(HaveLen(2))
			// kuma-sidecar is the second init container
			Expect(pod.Spec.InitContainers[1]).To(beSidecarWithoutPrivileged)
		}

		// when
		// pod with patch
		podName, err = PodNameOfApp(kubernetes.Cluster, appNameWithPatch, namespace)
		Expect(err).ToNot(HaveOccurred())
		pod, err = k8s.GetPodE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(namespace), podName)
		Expect(err).ToNot(HaveOccurred())

		// then
		pointerTrue := new(bool)
		*pointerTrue = true
		Expect(pod.Spec.InitContainers).To(
			Or(HaveLen(1), HaveLen(2)),
		)
		// should doesn't have defined RunAsUser
		Expect(pod.Spec.InitContainers[0].SecurityContext.RunAsUser).To(BeNil())
		Expect(pod.Spec.Containers).To(
			Or(HaveLen(2), HaveLen(1)),
		)
		beSidecarWithPrivileged := And(
			WithTransform(func(c kube_core.Container) string { return c.Name }, BeEquivalentTo(k8s_util.KumaSidecarContainerName)),
			WithTransform(func(c kube_core.Container) *bool { return c.SecurityContext.Privileged }, Equal(pointerTrue)),
		)
		if len(pod.Spec.Containers) == 2 {
			// kuma-sidecar is the first container
			Expect(pod.Spec.Containers[0]).To(beSidecarWithPrivileged)
		} else {
			Expect(pod.Spec.InitContainers).To(HaveLen(2))
			// kuma-sidecar is the second init container
			Expect(pod.Spec.InitContainers[1]).To(beSidecarWithPrivileged)
		}
	})

	It("should reject ContainerPatch in non-system namespace", func() {
		// when
		err := k8s.KubectlApplyFromStringE(kubernetes.Cluster.GetTesting(), kubernetes.Cluster.GetKubectlOptions(), containerPatch(namespace))

		// then
		Expect(err).To(HaveOccurred())
	})
}
