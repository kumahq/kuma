//nolint:staticcheck // SA1019 Test file: tests backward compatibility with deprecated core_rules.Rule
package v1alpha1_test

import (
	"fmt"
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
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
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

var _ = Describe("MeshTLS", func() {
	type testCase struct {
		caseName         string
		meshBuilder      *builders.MeshBuilder
		meshService      bool
		workloadIdentity *core_xds.WorkloadIdentity
		casByTrustDomain map[string][]xds_context.PEMBytes
		features         xds_types.Features
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
			if given.meshService {
				resourceSet.Add(getMeshServiceResources(secretsTracker, mesh)...)
			} else {
				resourceSet.Add(getResources(secretsTracker, mesh)...)
			}

			policy := getPolicy(given.caseName)

			proxyBuilder := xds_builders.Proxy().
				WithSecretsTracker(secretsTracker).
				WithWorkloadIdentity(given.workloadIdentity).
				WithApiVersion(envoy_common.APIV3).
				WithOutbounds(xds_types.Outbounds{&xds_types.Outbound{
					LegacyOutbound: builders.Outbound().
						WithService("outgoing").
						WithAddress("127.0.0.1").
						WithPort(27777).Build(),
				}}).
				WithDataplane(
					builders.Dataplane().
						WithName("test").
						WithMesh("default").
						WithAddress("127.0.0.1").
						WithTransparentProxying(15006, 15001, "ipv4").
						AddOutbound(
							builders.Outbound().
								WithAddress("127.0.0.1").
								WithPort(27777).
								WithService("outgoing"),
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
				WithPolicies(xds_builders.MatchedPolicies().WithFromPolicy(api.MeshTLSType, getRulesAsFromRules(pointer.Deref(policy.Spec.Rules))))

			if given.features != nil {
				proxyBuilder.WithMetadata(&core_xds.DataplaneMetadata{
					Features: given.features,
				})
			}

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
		Entry("strict with no mTLS on the mesh", testCase{
			caseName:    "strict-no-mtls",
			meshBuilder: samples.MeshDefaultBuilder(),
		}),
		Entry("permissive with no mTLS on the mesh", testCase{
			caseName:    "permissive-no-mtls",
			meshBuilder: samples.MeshDefaultBuilder(),
		}),
		Entry("strict with permissive mTLS on the mesh", testCase{
			caseName:    "strict-with-permissive-mtls",
			meshBuilder: samples.MeshMTLSBuilder().WithPermissiveMTLSBackends(),
		}),
		Entry("permissive with permissive mTLS on the mesh", testCase{
			caseName:    "permissive-with-permissive-mtls",
			meshBuilder: samples.MeshMTLSBuilder().WithPermissiveMTLSBackends(),
		}),
		Entry("strict with permissive mTLS on the mesh for MeshService", testCase{
			caseName:    "strict-with-permissive-mtls-meshservice",
			meshBuilder: samples.MeshMTLSBuilder().WithPermissiveMTLSBackends(),
			meshService: true,
		}),
		Entry("strict based on workload identity", testCase{
			caseName:    "strict-with-workload-identity",
			meshBuilder: samples.MeshMTLSBuilder(),
			meshService: true,
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI: kri.Identifier{ResourceType: meshidentity_api.MeshIdentityType, Mesh: "default", Zone: "default", Name: "my-identity"},
				IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"my-secret-name",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
			},
		}),
		Entry("permissive based on workload identity and custom functions", testCase{
			caseName:    "permissive-with-workload-identity-custom-functions",
			meshBuilder: samples.MeshMTLSBuilder(),
			meshService: true,
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
			caseName:    "strict-with-mesh-trust",
			meshBuilder: samples.MeshMTLSBuilder(),
			meshService: true,
			casByTrustDomain: map[string][]xds_context.PEMBytes{
				"domain-1": {
					xds_context.PEMBytes("123"),
				},
			},
		}),
		Entry("strict using external validator", testCase{
			caseName:    "strict-with-external-validator",
			meshBuilder: samples.MeshMTLSBuilder(),
			meshService: true,
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
			meshBuilder: samples.MeshMTLSBuilder(),
			meshService: true,
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
			meshBuilder: samples.MeshMTLSBuilder(),
			meshService: true,
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
		Entry("strict mode + strict mesh = no passthrough listeners", testCase{
			caseName:    "strict-with-strict-mtls",
			meshBuilder: samples.MeshMTLSBuilder(),
		}),
		Entry("permissive mode + strict mesh = passthrough listeners", testCase{
			caseName:    "permissive-with-strict-mtls",
			meshBuilder: samples.MeshMTLSBuilder(),
		}),
		Entry("workload identity without CA = passthrough listeners", testCase{
			caseName:    "strict-with-workload-identity-no-ca",
			meshBuilder: samples.MeshDefaultBuilder(),
			meshService: true,
			workloadIdentity: &core_xds.WorkloadIdentity{
				KRI: kri.Identifier{ResourceType: meshidentity_api.MeshIdentityType, Mesh: "default", Zone: "default", Name: "my-identity"},
				IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"my-secret-name",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
			},
		}),
		Entry("strict inbound ports feature = port filtering", testCase{
			caseName:    "strict-with-feature-strict-inbound-ports",
			meshBuilder: samples.MeshMTLSBuilder().WithPermissiveMTLSBackends(),
			features: xds_types.Features{
				xds_types.FeatureStrictInboundPorts: true,
			},
		}),
	)
})

func getMeshServiceResources(secretsTracker core_xds.SecretsTracker, mesh *builders.MeshBuilder) []*core_xds.Resource {
	return []*core_xds.Resource{
		{
			Name:   "inbound:127.0.0.1:17777",
			Origin: metadata.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
					Configure(
						listeners.HttpInboundRoutes(
							envoy_names.GetInboundRouteName("backend"),
							"backend",
							envoy_common.Routes{
								{
									Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
										envoy_common.WithService("backend"),
										envoy_common.WithWeight(100),
									)},
								},
							},
						),
					),
				)).MustBuild(),
		},
		{
			Name:   "inbound:127.0.0.1:17778",
			Origin: metadata.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP, true).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
				)).MustBuild(),
		},
		{
			Name:   "outbound",
			Origin: metadata.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "outgoing").
				Configure(clusters.ClientSideMTLS(secretsTracker, false, mesh.Build(), "outgoing", true, nil, false)).
				MustBuild(),
			Protocol: core_meta.ProtocolHTTP,
			ResourceOrigin: kri.Identifier{
				ResourceType: "MeshService",
				Mesh:         "default",
				Zone:         "zone-1",
				Namespace:    "backend-ns",
				Name:         "backend",
				SectionName:  "",
			},
		},
	}
}

func getResources(secretsTracker core_xds.SecretsTracker, mesh *builders.MeshBuilder) []*core_xds.Resource {
	return []*core_xds.Resource{
		{
			Name:   "inbound:127.0.0.1:17777",
			Origin: metadata.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17777, core_xds.SocketAddressProtocolTCP, true).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.HttpConnectionManager("127.0.0.1:17777", false, nil, true)).
					Configure(
						listeners.HttpInboundRoutes(
							envoy_names.GetInboundRouteName("backend"),
							"backend",
							envoy_common.Routes{
								{
									Clusters: []envoy_common.Cluster{envoy_common.NewCluster(
										envoy_common.WithService("backend"),
										envoy_common.WithWeight(100),
									)},
								},
							},
						),
					),
				)).MustBuild(),
		},
		{
			Name:   "inbound:127.0.0.1:17778",
			Origin: metadata.OriginInbound,
			Resource: listeners.NewInboundListenerBuilder(envoy_common.APIV3, "127.0.0.1", 17778, core_xds.SocketAddressProtocolTCP, true).
				Configure(listeners.FilterChain(listeners.NewFilterChainBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
					Configure(listeners.TcpProxyDeprecated("127.0.0.1:17778", envoy_common.NewCluster(envoy_common.WithName("frontend")))),
				)).MustBuild(),
		},
		{
			Name:   "outgoing",
			Origin: metadata.OriginOutbound,
			Resource: clusters.NewClusterBuilder(envoy_common.APIV3, "outgoing").
				Configure(clusters.ClientSideMTLS(secretsTracker, false, mesh.Build(), "outgoing", true, nil, false)).
				MustBuild(),
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
	var legacyRules []*core_rules.Rule
	var rules []*inbound.Rule

	for _, rule := range policyRules {
		legacyRules = append(legacyRules, &core_rules.Rule{
			Subset: subsetutils.Subset{},
			Conf:   rule.Default,
		})
		rules = append(rules, &inbound.Rule{
			Conf:   rule.Default,
			Origin: common.Origin{},
		})
	}

	return core_rules.FromRules{
		Rules: map[core_rules.InboundListener]core_rules.Rules{
			{Address: "127.0.0.1", Port: 17777}: legacyRules,
			{Address: "127.0.0.1", Port: 17778}: legacyRules,
		},
		InboundRules: map[core_rules.InboundListener][]*inbound.Rule{
			{Address: "127.0.0.1", Port: 17777}: rules,
			{Address: "127.0.0.1", Port: 17778}: rules,
		},
	}
}
