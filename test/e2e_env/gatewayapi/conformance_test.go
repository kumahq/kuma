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
	clientgo_kube "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi_v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance"
	conformanceapis "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	conformancegrpc "sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
	"sigs.k8s.io/yaml"

	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/version"
	. "github.com/kumahq/kuma/v3/test/framework"
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
	Init() // we need to manually call init on the config because it's not a Ginkgo test
	if Config.IPV6 {
		t.Skip("On IPv6 we run on kind which doesn't support load balancers")
	}

	g := NewWithT(t)

	cluster := NewK8sCluster(t, clusterName, Silent)
	opts := cluster.GetKubectlOptions()

	t.Cleanup(func() {
		var meshNamespaces []string
		clientset, err := k8s.GetKubernetesClientFromOptionsContextE(t, context.Background(), opts)
		if err == nil {
			if nsList, err := clientset.CoreV1().Namespaces().List(context.Background(),
				metav1.ListOptions{
					LabelSelector: fmt.Sprintf("%s=%s", metadata.KumaSidecarInjectionAnnotation, metadata.AnnotationEnabled),
				}); err == nil {
				for _, ns := range nsList.Items {
					meshNamespaces = append(meshNamespaces, ns.Name)
				}
			}
		}

		if t.Failed() {
			g.Expect(func() error { //nolint:unparam  // we need this return type to be included in the Expect block
				RegisterFailHandler(g.Fail)
				DebugKube(cluster, "default", meshNamespaces...)
				return nil
			}()).To(Succeed())
		}

		for _, ns := range meshNamespaces {
			g.Expect(cluster.DeleteNamespace(ns)).To(Succeed())
		}
		g.Expect(cluster.DeleteKuma()).To(Succeed())
		g.Expect(cluster.DismissCluster()).To(Succeed())
	})

	g.Expect(cluster.Install(GatewayAPICRDs)).To(Succeed())
	g.Eventually(func() error {
		return NewClusterSetup().Install(
			Kuma(config_core.Zone)).Setup(cluster)
	}, "90s", "3s").Should(Succeed())

	g.Eventually(func() error {
		return YamlK8s(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
`)(cluster)
	}, "30s", "3s").Should(Succeed())

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
		GRPCClient:           &conformancegrpc.DefaultClient{},
		ManifestFS:           []fs.FS{&conformance.Manifests},
		GatewayClassName:     "kuma",
		CleanupBaseResources: false,
		CleanupTestResources: true,
		Debug:                Config.Debug,
		NamespaceLabels: map[string]string{
			metadata.KumaSidecarInjectionAnnotation: metadata.AnnotationEnabled,
		},
		SkipTests: []string{
			"HTTPRouteNoBackendRefs",
		},
		// Left undeclared, with what Kuma does not do:
		//   - SupportMeshHTTPRouteBackendRequestHeaderModification: Kuma's BackendRef type
		//     (api/common/v1alpha1.BackendRef) has no Filters field, so
		//     rules[].backendRefs[].filters cannot be translated.
		//   - SupportMeshHTTPRouteNamedRouteRule: the upstream test is Provisional and
		//     asserts a backend path of /named for a /unnamed request with no rewrite in
		//     its own manifest.
		//   - SupportMeshHTTPRouteQueryParamMatching: requests that don't match any rule's
		//     query params still land on a backend (200) instead of getting a 404; Kuma
		//     doesn't have a deny-on-no-match fallback for MeshHTTPRoute.
		SupportedFeatures: []features.FeatureName{
			features.SupportHTTPRouteResponseHeaderModification,
			features.SupportHTTPRoute,
			features.SupportGRPCRoute,
			features.SupportHTTPRoute303RedirectStatusCode,
			features.SupportHTTPRoute307RedirectStatusCode,
			features.SupportHTTPRoute308RedirectStatusCode,
			features.SupportHTTPRouteParentRefPort,
			features.SupportMesh,
			features.SupportMeshClusterIPMatching,
			features.SupportMeshConsumerRoute,
			features.SupportMeshHTTPRouteRedirectPath,
			features.SupportMeshHTTPRouteRedirectPort,
			features.SupportMeshHTTPRouteSchemeRedirect,
			features.SupportMeshHTTPRouteRewritePath,
		},
		Implementation: implementation,
		ConformanceProfiles: []suite.ConformanceProfileName{
			suite.MeshHTTPConformanceProfileName,
			suite.MeshGRPCConformanceProfileName,
		},
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
