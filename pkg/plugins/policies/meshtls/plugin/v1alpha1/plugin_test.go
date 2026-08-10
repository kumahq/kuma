package v1alpha1_test

import (
	"fmt"
	"maps"
	"os"
	"path"

	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	bldrs_common "github.com/kumahq/kuma/v3/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v3/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/v3/pkg/envoy/builders/tls"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtls/api/v1alpha1"
	plugin "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtls/plugin/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_yaml "github.com/kumahq/kuma/v3/pkg/util/yaml"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

// kumaWorkloadIdentity is the identity every proxy that MeshTLS runs on carries.
func kumaWorkloadIdentity() *core_xds.WorkloadIdentity {
	return &core_xds.WorkloadIdentity{
		KRI: kri.Identifier{ResourceType: meshidentity_api.MeshIdentityType, Mesh: "default", Zone: "default", Name: "my-identity"},
		IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
			return bldrs_tls.SdsSecretConfigSource(
				"my-secret-name",
				bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
			)
		},
	}
}

var _ = Describe("MeshTLS", func() {
	type testCase struct {
		caseName         string
		meshBuilder      *builders.MeshBuilder
		workloadIdentity *core_xds.WorkloadIdentity
		casByTrustDomain map[string][]xds_context.PEMBytes
		features         xds_types.Features
		// noPolicy exercises a mesh without any MeshTLS policy, where the mode
		// falls back to Strict
		noPolicy bool
		// ipFamilyMode of the transparent proxy, "ipv4" when empty
		ipFamilyMode string
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// given
			mesh := given.meshBuilder
			context := *xds_builders.Context().
				WithMeshBuilder(mesh).
				WithCAsByTrustDomain(given.casByTrustDomain).
				Build()
			resourceSet := core_xds.NewResourceSet()
			secretsTracker := envoy_common.NewSecretsTracker("default", nil)
			resourceSet.Add(getMeshServiceResources()...)

			fromRules := core_rules.FromRules{}
			if !given.noPolicy {
				fromRules = getRulesAsFromRules(pointer.Deref(getPolicy(given.caseName).Spec.Rules))
			}

			ipFamilyMode := given.ipFamilyMode
			if ipFamilyMode == "" {
				ipFamilyMode = "ipv4"
			}

			proxyBuilder := xds_builders.Proxy().
				WithSecretsTracker(secretsTracker).
				WithWorkloadIdentity(given.workloadIdentity).
				WithApiVersion(envoy_common.APIV3).
				WithDataplane(
					builders.Dataplane().
						WithName("test").
						WithMesh("default").
						WithAddress("127.0.0.1").
						WithTransparentProxying(15006, 15001, ipFamilyMode).
						AddOutbound(
							builders.Outbound().
								WithAddress("127.0.0.1").
								WithPort(27777).
								WithMeshService("outgoing", 80),
						).
						AddInbound(
							builders.Inbound().
								WithAddress("127.0.0.1").
								WithPort(17777).
								WithService("backend"),
						).
						AddInbound(
							builders.Inbound().
								WithAddress("127.0.0.1").
								WithPort(17778).
								WithService("frontend"),
						),
				).
				WithPolicies(xds_builders.MatchedPolicies().WithFromPolicy(api.MeshTLSType, fromRules))

			features := xds_types.Features{}
			maps.Copy(features, given.features)
			proxyBuilder.WithMetadata(&core_xds.DataplaneMetadata{Features: features})

			proxy := proxyBuilder.Build()

			plugin := plugin.NewPlugin().(core_plugins.PolicyPlugin)

			// when
			Expect(plugin.Apply(resourceSet, context, proxy)).To(Succeed())

			// then
			resource, err := util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.ListenerType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.listeners.golden.yaml", given.caseName)))
			resource, err = util_yaml.GetResourcesToYaml(resourceSet, envoy_resource.ClusterType)
			Expect(err).ToNot(HaveOccurred())
			Expect(resource).To(matchers.MatchGoldenYAML(fmt.Sprintf("testdata/%s.clusters.golden.yaml", given.caseName)))
		},
		Entry("no workload identity = plugin does not run", testCase{
			caseName:    "strict-no-mtls",
			meshBuilder: samples.MeshDefaultBuilder(),
		}),
		Entry("strict with SNI matchers", testCase{
			caseName:         "strict-with-permissive-mtls",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaWorkloadIdentity(),
		}),
		Entry("permissive with SNI matchers", testCase{
			caseName:         "permissive-with-permissive-mtls",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaWorkloadIdentity(),
		}),
		Entry("strict based on workload identity", testCase{
			caseName:         "strict-with-workload-identity",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaWorkloadIdentity(),
		}),
		Entry("permissive based on workload identity and custom functions", testCase{
			caseName:    "permissive-with-workload-identity-custom-functions",
			meshBuilder: samples.MeshDefaultBuilder(),
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI: kri.Identifier{ResourceType: meshidentity_api.MeshIdentityType, Mesh: "default", Zone: "default", Name: "my-identity"},
				IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"my-secret-name",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
				ExternalValidationSourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"ca-bundle",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
			},
		}),
		Entry("strict with MeshTrust", testCase{
			caseName:         "strict-with-mesh-trust",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaWorkloadIdentity(),
			casByTrustDomain: map[string][]xds_context.PEMBytes{
				"domain-1": {
					xds_context.PEMBytes("123"),
				},
			},
		}),
		Entry("strict using external validator", testCase{
			caseName:    "strict-with-external-validator",
			meshBuilder: samples.MeshDefaultBuilder(),
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI: kri.Identifier{ResourceType: meshidentity_api.MeshIdentityType, Mesh: "default", Zone: "default", Name: "my-identity"},
				IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"my-secret-name",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
				ExternalValidationSourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"ca-bundle",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
			},
			casByTrustDomain: map[string][]xds_context.PEMBytes{
				"domain-1": {
					xds_context.PEMBytes("123"),
				},
			},
		}),
		Entry("strict with MeshTrust and kuma managed identity", testCase{
			caseName:    "strict-with-mesh-trust-kuma-managed",
			meshBuilder: samples.MeshDefaultBuilder(),
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI:            kri.Identifier{ResourceType: meshidentity_api.MeshIdentityType, Mesh: "default", Zone: "default", Name: "my-identity"},
				ManagementMode: core_xds.KumaManagementMode,
				IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"my-secret-name",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
			},
			casByTrustDomain: map[string][]xds_context.PEMBytes{
				"domain-1": {
					xds_context.PEMBytes("123"),
				},
			},
		}),
		Entry("strict with multiple MeshTrust and kuma managed identity", testCase{
			caseName:    "strict-with-multiple-mesh-trust-kuma-managed",
			meshBuilder: samples.MeshDefaultBuilder(),
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI:            kri.Identifier{ResourceType: meshidentity_api.MeshIdentityType, Mesh: "default", Zone: "default", Name: "my-identity"},
				ManagementMode: core_xds.KumaManagementMode,
				IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"my-secret-name",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
			},
			// deliberately out of alphabetical order to verify SANs are sorted
			casByTrustDomain: map[string][]xds_context.PEMBytes{
				"domain-c": {xds_context.PEMBytes("123")},
				"domain-a": {xds_context.PEMBytes("456")},
				"domain-b": {xds_context.PEMBytes("789")},
			},
		}),
		Entry("strict mode = no passthrough listeners", testCase{
			caseName:         "strict-with-strict-mtls",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaWorkloadIdentity(),
		}),
		Entry("permissive mode = passthrough listeners", testCase{
			caseName:         "permissive-with-strict-mtls",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaWorkloadIdentity(),
		}),
		Entry("workload identity without CA = passthrough listeners", testCase{
			caseName:         "strict-with-workload-identity-no-ca",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaWorkloadIdentity(),
		}),
		Entry("tls version on both ends of the range with workload identity", testCase{
			caseName:         "strict-with-workload-identity-tls-version",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaWorkloadIdentity(),
		}),
		Entry("dualstack tproxy = ipv4 and ipv6 passthrough listeners", testCase{
			caseName:         "permissive-with-dualstack-tproxy",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaWorkloadIdentity(),
			ipFamilyMode:     "dualstack",
		}),
		Entry("no policy = strict", testCase{
			caseName:         "no-policy-with-strict-mtls",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaWorkloadIdentity(),
			noPolicy:         true,
		}),
		Entry("strict inbound ports feature with workload identity = port filtering", testCase{
			caseName:         "strict-with-workload-identity-strict-inbound-ports",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaWorkloadIdentity(),
			features: xds_types.Features{
				xds_types.FeatureStrictInboundPorts: true,
			},
		}),
	)
})

// outgoingMeshService identifies the destination the dataplane's only outbound
// points at, and gives the outbound cluster its unified name.
var outgoingMeshService = kri.Identifier{
	ResourceType: "MeshService",
	Mesh:         "default",
	Zone:         "zone-1",
	Namespace:    "backend-ns",
	Name:         "outgoing",
	SectionName:  "80",
}

func getMeshServiceResources() []*core_xds.Resource {
	return []*core_xds.Resource{
		{
			Name:   "inbound:127.0.0.1:17777",
			Origin: metadata.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
					Configure(
						listeners.HttpInboundRoute(
							envoy_names.GetInboundRouteName("backend"),
							"backend",
							plugins_xds.NewClusterBuilder().WithService("backend").Build(),
						),
					),
				)).MustBuild(),
		},
		{
			Name:   "inbound:127.0.0.1:17778",
			Origin: metadata.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP, true).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.TcpProxyDeprecated("127.0.0.1:17778", plugins_xds.NewClusterBuilder().WithName("frontend").Build())),
				)).MustBuild(),
		},
		{
			Name:   outgoingMeshService.String(),
			Origin: metadata.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, outgoingMeshService.String()).
				Configure(clusters.UpstreamTLSContext(&envoy_tls.UpstreamTlsContext{
					CommonTlsContext: &envoy_tls.CommonTlsContext{},
					Sni:              "outgoing",
				})).
				MustBuild(),
			Protocol:       core_meta.ProtocolHTTP,
			ResourceOrigin: outgoingMeshService,
		},
	}
}

func getPolicy(caseName string) *api.MeshTLSResource {
	// setup
	meshTLS := api.NewMeshTLSResource()

	// when
	contents, err := os.ReadFile(path.Join("testdata", caseName+".policy.yaml"))
	Expect(err).ToNot(HaveOccurred())
	err = core_model.FromYAML(contents, &meshTLS.Spec)
	Expect(err).ToNot(HaveOccurred())

	meshTLS.SetMeta(&test_model.ResourceMeta{
		Name: "name",
		Mesh: core_model.DefaultMesh,
	})
	// and
	verr := meshTLS.Validate()
	Expect(verr).ToNot(HaveOccurred())

	return meshTLS
}

func getRulesAsFromRules(policyRules []api.Rule) core_rules.FromRules {
	var rules []*inbound.Rule

	for _, rule := range policyRules {
		rules = append(rules, &inbound.Rule{
			Conf:   rule.Default,
			Origin: common.Origin{},
		})
	}

	return core_rules.FromRules{
		InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
			{Address: "127.0.0.1", Port: 17777}: rules,
			{Address: "127.0.0.1", Port: 17778}: rules,
		},
	}
}
