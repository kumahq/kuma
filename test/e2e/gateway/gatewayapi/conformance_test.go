package gatewayapi_test

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"
	clientgo_kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	apis_gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"
	conformanceapis "sigs.k8s.io/gateway-api/conformance/apis/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
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
	// this is like job-0
	if os.Getenv("CIRCLE_NODE_INDEX") != "" && os.Getenv("CIRCLE_NODE_INDEX") != "0" {
		t.Skip("Conformance tests are only run on job 0")
	}
	if Config.IPV6 {
		t.Skip("On IPv6 we run on kind which doesn't support load balancers")
	}

	g := NewWithT(t)

	cluster := NewK8sCluster(t, clusterName, Silent)

	t.Cleanup(func() {
		g.Expect(cluster.DeleteKuma()).To(Succeed())
		g.Expect(cluster.DismissCluster()).To(Succeed())
	})

	g.Expect(cluster.Install(GatewayAPICRDs)).To(Succeed())
	g.Eventually(func() error {
		return NewClusterSetup().
			Install(Kuma(config_core.Standalone,
				WithCtlOpts(map[string]string{"--experimental-gatewayapi": "true"}),
			)).
			Setup(cluster)
	}, "90s", "3s").Should(Succeed())

	opts := cluster.GetKubectlOptions()

	configPath, err := opts.GetConfigPath(t)
	g.Expect(err).ToNot(HaveOccurred())

	config := k8s.LoadConfigFromPath(configPath)

	clientConfig, err := config.ClientConfig()
	g.Expect(err).ToNot(HaveOccurred())

	client, err := client.New(clientConfig, client.Options{})
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(apis_gatewayapi.AddToScheme(client.Scheme())).To(Succeed())

	clientset, err := clientgo_kube.NewForConfig(clientConfig)
	g.Expect(err).ToNot(HaveOccurred())

	suiteOpts := suite.Options{
		Client:               client,
		RESTClient:           clientset.CoreV1().RESTClient().(*rest.RESTClient),
		RestConfig:           clientConfig,
		Clientset:            clientset,
		GatewayClassName:     "kuma",
		CleanupBaseResources: true,
		Debug:                false,
		NamespaceLabels: map[string]string{
			metadata.KumaSidecarInjectionAnnotation: metadata.AnnotationTrue,
		},
		SupportedFeatures: sets.New(
			suite.SupportGateway,
			suite.SupportHTTPResponseHeaderModification,
			suite.SupportHTTPRoute,
			suite.SupportHTTPRouteHostRewrite,
			suite.SupportHTTPRouteMethodMatching,
			suite.SupportHTTPRoutePathRedirect,
			suite.SupportHTTPRoutePathRewrite,
			suite.SupportHTTPRoutePortRedirect,
			suite.SupportHTTPRouteQueryParamMatching,
			suite.SupportHTTPRouteRequestMirror,
			suite.SupportHTTPRouteSchemeRedirect,
			suite.SupportMesh,
		),
	}

	conformanceSuite, err := suite.NewExperimentalConformanceTestSuite(
		suite.ExperimentalConformanceOptions{
			Options:             suiteOpts,
			Implementation:      implementation,
			ConformanceProfiles: sets.New(suite.HTTPConformanceProfileName, suite.MeshConformanceProfileName),
		},
	)
	g.Expect(err).ToNot(HaveOccurred())

	conformanceSuite.Setup(t)

	var passingTests []suite.ConformanceTest
	for _, test := range tests.ConformanceTests {
		// This is an easy way to enable/disable single tests when upgrading/debugging
		switch test.ShortName {
		}
		passingTests = append(passingTests, test)
	}

	g.Expect(conformanceSuite.Run(t, passingTests)).To(Succeed())

	rep, err := conformanceSuite.Report()
	g.Expect(err).ToNot(HaveOccurred())
	repYaml, err := yaml.Marshal(rep)
	g.Expect(err).ToNot(HaveOccurred())

	t.Log("Gateway API CONFORMANCE REPORT:")
	t.Logf("\n%s", string(repYaml))
}
