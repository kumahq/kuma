package gatewayapi_test

import (
	"runtime"
	"testing"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	apis_gatewayapi "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/test/e2e/gateway/gatewayapi"
	. "github.com/kumahq/kuma/test/framework"
)

var clusterName = Kuma1
var minNodePort = 30080
var maxNodePort = 30089

// TestConformance runs as a `testing` test and not Ginkgo so we have to use an
// explicit `g` to use Gomega.
func TestConformance(t *testing.T) {
	if Config.IPV6 {
		t.Skip("On IPv6 we run on kind which doesn't support load balancers")
	}
	if runtime.GOARCH == "arm64" {
		t.Skip("On ARM64 it's not supported yet")
	}

	g := NewWithT(t)

	cluster := NewK8sCluster(t, clusterName, Silent)

	defer func() {
		g.Expect(cluster.DeleteKuma()).To(Succeed())
		g.Expect(cluster.DismissCluster()).To(Succeed())
	}()

	err := NewClusterSetup().
		Install(gatewayapi.GatewayAPICRDs).
		Install(Kuma(config_core.Standalone,
			WithCtlOpts(map[string]string{"--experimental-gatewayapi": "true"}),
		)).
		Install(YamlK8s(gatewayapi.GatewayClass)).
		Setup(cluster)
	g.Expect(err).ToNot(HaveOccurred())

	opts := cluster.GetKubectlOptions()

	configPath, err := opts.GetConfigPath(t)
	g.Expect(err).ToNot(HaveOccurred())

	config := k8s.LoadConfigFromPath(configPath)

	clientConfig, err := config.ClientConfig()
	g.Expect(err).ToNot(HaveOccurred())

	client, err := client.New(clientConfig, client.Options{})
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(apis_gatewayapi.AddToScheme(client.Scheme())).To(Succeed())

	var validUniqueListenerPorts []apis_gatewayapi.PortNumber
	for i := minNodePort; i <= maxNodePort; i++ {
		validUniqueListenerPorts = append(validUniqueListenerPorts, apis_gatewayapi.PortNumber(i))
	}

	conformanceSuite := suite.New(suite.Options{
		Client:           client,
		GatewayClassName: "kuma",
		Cleanup:          true,
		Debug:            false,
		NamespaceLabels: map[string]string{
			metadata.KumaSidecarInjectionAnnotation: metadata.AnnotationTrue,
		},
		ValidUniqueListenerPorts: validUniqueListenerPorts,
	})

	conformanceSuite.Setup(t)
	conformanceSuite.Run(t, tests.ConformanceTests)
}
