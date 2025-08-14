package spire_test

import (
	"context"
	"path/filepath"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/providers"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/providers/spire"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/core/xds/types"
	bldrs_core "github.com/kumahq/kuma/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/pkg/envoy/builders/tls"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
)

var _ = Describe("Spire Providers Test", func() {
	var spireProvider providers.IdentityProvider
	BeforeEach(func() {
		spireProvider = spire.NewSpireIdentityProvider("/run/socket/sockets", "socket", "my-zone")
	})

	Context("CreateIdentity", func() {
		It("should provide an identity for a dataplane", func() {
			meshIdentity := builders.MeshIdentity().
				WithName("matching-1").
				WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app":     "test-app",
						"version": "v1",
					},
				}).WithSpire().Build()
			expectedIdentity, err := bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(
				bldrs_tls.SdsSecretConfigSource(
					"spiffe://default.my-zone.mesh.local/ns/my-ns/sa/my-sa",
					bldrs_core.NewConfigSource().Configure(bldrs_core.ApiConfigSource(spire.SpireAgentClusterName)),
				),
			).Build()
			Expect(err).ToNot(HaveOccurred())
			expectedValidation, err := bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(
				bldrs_tls.SdsSecretConfigSource(
					spire.FederatedCASecretName,
					bldrs_core.NewConfigSource().Configure(bldrs_core.ApiConfigSource(spire.SpireAgentClusterName)),
				),
			).Build()
			Expect(err).ToNot(HaveOccurred())

			proxy := xds_builders.Proxy().
				WithMetadata(&core_xds.DataplaneMetadata{
					Features: types.Features{
						types.FeatureSpire: true,
					},
				}).
				WithDataplane(builders.Dataplane().
					WithName("web-01").
					WithAddress("192.168.0.2").
					WithLabels(map[string]string{
						metadata.KumaServiceAccount: "my-sa",
						mesh_proto.KubeNamespaceTag: "my-ns",
					}).
					WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")).
				Build()

			// create identity
			identity, err := spireProvider.CreateIdentity(context.TODO(), meshIdentity, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(identity.KRI).To(Equal(kri.From(meshIdentity)))
			Expect(identity.ManagementMode).To(Equal(core_xds.ExternalManagementMode))
			// we don't manage secrets
			Expect(identity.ExpirationTime).To(BeNil())
			Expect(identity.GenerationTime).To(BeNil())

			identitySource, err := bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(identity.IdentitySourceConfigurer()).Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(identitySource).To(Equal(expectedIdentity))

			validatorSource, err := bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(identity.ValidationSourceConfigurer()).Build()
			Expect(err).ToNot(HaveOccurred())
			Expect(validatorSource).To(Equal(expectedValidation))

			resources, err := util_yaml.GetResourcesToYaml(identity.AdditionalResources, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resources).To(matchers.MatchGoldenYAML(filepath.Join("testdata", "spire.cluster.golden.yaml")))
		})

		It("should not provide an identity for a dataplane without spire support", func() {
			meshIdentity := builders.MeshIdentity().
				WithName("matching-1").
				WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app":     "test-app",
						"version": "v1",
					},
				}).WithSpire().Build()

			proxy := xds_builders.Proxy().
				WithMetadata(&core_xds.DataplaneMetadata{
					Features: types.Features{},
				}).
				WithDataplane(builders.Dataplane().
					WithName("web-01").
					WithAddress("192.168.0.2").
					WithLabels(map[string]string{
						metadata.KumaServiceAccount: "my-sa",
						mesh_proto.KubeNamespaceTag: "my-ns",
					}).
					WithInboundOfTags(mesh_proto.ServiceTag, "web", mesh_proto.ProtocolTag, "http")).
				Build()

			// create identity
			identity, err := spireProvider.CreateIdentity(context.TODO(), meshIdentity, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(identity).To(BeNil())
		})
	})
})
