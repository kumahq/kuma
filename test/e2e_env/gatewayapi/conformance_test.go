package gatewayapi_test

import (
	"context"
	"fmt"
	"io/fs"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/gomega"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	clientgo_kube "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance"
	conformanceapis "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
	"sigs.k8s.io/yaml"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/version"
	. "github.com/kumahq/kuma/test/framework"
)

var clusterName = Kuma1

var implementation = conformanceapis.Implementation{
	Organization: "kumahq",
	Project:      "kuma",
	URL:          "https://github.com/kumahq/kuma",
	Version:      version.Build.Version,
	Contact:      []string{"@kumahq/kuma-maintainers"},
}

// TestConformance runs as a `testing` test and not Ginkgo so we have to use an
// explicit `g` to use Gomega.
func TestConformance(t *testing.T) {
	if Config.IPV6 {
		t.Skip("On IPv6 we run on kind which doesn't support load balancers")
	}

	g := NewWithT(t)

	cluster := NewK8sCluster(t, clusterName, Silent)
	opts := cluster.GetKubectlOptions()

	t.Cleanup(func() {
		if t.Failed() || Config.Debug {
			var namespaces []string
			clientset, err := k8s.GetKubernetesClientFromOptionsE(t, opts)
			if err == nil {
				if nsList, err := clientset.CoreV1().Namespaces().List(context.Background(),
					metav1.ListOptions{
						LabelSelector: fmt.Sprintf("%s=%s", metadata.KumaSidecarInjectionAnnotation, metadata.AnnotationEnabled),
					}); err == nil {
					for _, ns := range nsList.Items {
						namespaces = append(namespaces, ns.Name)
					}
				}
			}

			if len(namespaces) > 0 {
				g.Expect(func() error { //nolint:unparam  // we need this return type to be included in the Expect block
					RegisterFailHandler(g.Fail)
					DebugKube(cluster, "default", namespaces...)
					return nil
				}()).To(Succeed())
			}
		}

		g.Expect(cluster.DeleteKuma()).To(Succeed())
		g.Expect(cluster.DismissCluster()).To(Succeed())
	})

	g.Expect(cluster.Install(GatewayAPICRDs)).To(Succeed())
	g.Eventually(func() error {
		return NewClusterSetup().Install(Kuma(config_core.Zone)).Setup(cluster)
	}, "90s", "3s").Should(Succeed())

	configPath, err := opts.GetConfigPath(t)
	g.Expect(err).ToNot(HaveOccurred())

	config := k8s.LoadConfigFromPath(configPath)

	clientConfig, err := config.ClientConfig()
	g.Expect(err).ToNot(HaveOccurred())

	client, err := client.New(clientConfig, client.Options{})
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(gatewayapi.Install(client.Scheme())).To(Succeed())
	g.Expect(gatewayapi_v1.Install(client.Scheme())).To(Succeed())
	g.Expect(apiextensionsv1.AddToScheme(client.Scheme())).To(Succeed())

	clientset, err := clientgo_kube.NewForConfig(clientConfig)
	g.Expect(err).ToNot(HaveOccurred())

	options := suite.ConformanceOptions{
		Client:               client,
		RestConfig:           clientConfig,
		Clientset:            clientset,
		GatewayClassName:     "kuma",
		CleanupBaseResources: true,
		Debug:                Config.Debug,
		NamespaceLabels: map[string]string{
			metadata.KumaSidecarInjectionAnnotation: metadata.AnnotationEnabled,
		},
		ManifestFS: []fs.FS{&conformance.Manifests},
		SupportedFeatures: sets.New(
			features.SupportGateway,
			features.SupportGatewayHTTPListenerIsolation,
			features.SupportGatewayPort8080,
			features.SupportReferenceGrant,
			features.SupportHTTPRouteResponseHeaderModification,
			features.SupportHTTPRoute,
			features.SupportHTTPRouteHostRewrite,
			features.SupportHTTPRouteMethodMatching,
			features.SupportHTTPRouteParentRefPort,
			features.SupportHTTPRoutePathRedirect,
			features.SupportHTTPRoutePathRewrite,
			features.SupportHTTPRoutePortRedirect,
			features.SupportHTTPRouteQueryParamMatching,
			features.SupportHTTPRouteRequestMirror,
			features.SupportHTTPRouteSchemeRedirect,
			features.SupportMesh,
			features.SupportMeshConsumerRoute,
		),
		Implementation:      implementation,
		ConformanceProfiles: sets.New(suite.GatewayHTTPConformanceProfileName, suite.MeshHTTPConformanceProfileName),
		// We are seeing flaky runs which are related to headless service cases, so ignoring them temporarily
		// See https://github.com/kumahq/kuma/pull/11463
		SkipTests: []string{tests.HTTPRouteServiceTypes.ShortName},
	}

	conformanceSuite, err := suite.NewConformanceTestSuite(options)
	g.Expect(err).ToNot(HaveOccurred())

	conformanceSuite.Setup(t, tests.ConformanceTests)
	g.Expect(conformanceSuite.Run(t, tests.ConformanceTests)).To(Succeed())

	rep, err := conformanceSuite.Report()
	g.Expect(err).ToNot(HaveOccurred())
	repYaml, err := yaml.Marshal(rep)
	g.Expect(err).ToNot(HaveOccurred())

	t.Log("Gateway API CONFORMANCE REPORT:")
	t.Logf("\n%s", string(repYaml))
}
