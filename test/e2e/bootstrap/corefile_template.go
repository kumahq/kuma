package bootstrap

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/config/core"
	k8s_util "github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/util"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/deployments/democlient"
)

func CorefileTemplate() {
	var k8sCluster *K8sCluster
	appNamespace := "dns-app"
	appName := "demo-dp-app"
	expectedTestText := "# this dummy corefile template is loaded from control plane"
	configMapName := "corefile-template"
	configMap := func(ns string) string {
		return fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
 namespace: %s
 name: %s
data:
 %s: |
    .:{{ .CoreDNSPort }} {
    %s
    log
    }`, ns, configMapName, configMapName, expectedTestText)
	}

	dnsConfigDir := "/tmp/kuma-dp-config/coredns"
	minReplicas := 3
	BeforeAll(func() {
		k8sCluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
			WithTimeout(6 * time.Second).
			WithRetries(60).(*K8sCluster)

		Expect(NewClusterSetup().
			Install(Namespace(Config.KumaNamespace)).
			Install(YamlK8s(configMap(Config.KumaNamespace))).
			Install(E2EKuma(core.Zone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(fmt.Sprintf("kuma-%s", strings.ToLower(random.UniqueId()))),
				WithHelmOpt("controlPlane.envVars.KUMA_BOOTSTRAP_SERVER_PARAMS_COREFILE_TEMPLATE_PATH",
					dnsConfigDir+"/"+configMapName),
				WithHelmOpt("controlPlane.extraConfigMaps[0].name", configMapName),
				WithHelmOpt("controlPlane.extraConfigMaps[0].mountPath", dnsConfigDir),
				WithHelmOpt("controlPlane.extraConfigMaps[0].readonly", "false"),
				WithHelmOpt("controlPlane.autoscaling.enabled", "true"),
				WithHelmOpt("controlPlane.autoscaling.minReplicas", strconv.Itoa(minReplicas)),
				WithHelmOpt("controlPlane.envVars.KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_EXPERIMENTAL_PROXY", "false"),
			)).
			Install(MeshKubernetes("default")).
			Install(NamespaceWithSidecarInjection(appNamespace)).
			Setup(k8sCluster),
		).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(k8sCluster, "default", appNamespace, Config.KumaNamespace)
	})

	E2EAfterAll(func() {
		Expect(k8sCluster.TriggerDeleteNamespace(appNamespace)).To(Succeed())
		Expect(k8sCluster.DeleteKuma()).To(Succeed())
		Expect(k8sCluster.DismissCluster()).To(Succeed())
	})

	It("should deploy 3 CP replicas", func() {
		Expect(k8sCluster.WaitApp(Config.KumaServiceName, Config.KumaNamespace, minReplicas)).To(Succeed())
	})

	It("should use Corefile template from control plane at data plane", func() {
		Expect(NewClusterSetup().
			Install(democlient.Install(
				democlient.WithName(appName),
				democlient.WithNamespace(appNamespace),
				democlient.WithPodAnnotations(map[string]string{
					"kuma.io/sidecar-env-vars": fmt.Sprintf("KUMA_DNS_CONFIG_DIR=%s", dnsConfigDir),
				}),
			)).Setup(k8sCluster),
		).To(Succeed())
		dpPod, err := PodNameOfApp(k8sCluster, appName, appNamespace)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			stdout, _, err := k8sCluster.Exec(
				appNamespace, dpPod, k8s_util.KumaSidecarContainerName, "cat", dnsConfigDir+"/Corefile")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring(expectedTestText))
		}, "3m", "1s").Should(Succeed())
	})
}
