package v1alpha1_test

import (
	"fmt"
	"maps"
	"os"
	"path"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
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
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
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

func identitySource(secretName string) func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
	return func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
		return bldrs_tls.SdsSecretConfigSource(
			secretName,
			bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
		)
	}
}

// workloadIdentity is the identity every proxy that MeshTLS runs on carries.
func workloadIdentity() *core_xds.WorkloadIdentity {
	return &core_xds.WorkloadIdentity{
		KRI:                      kri.Identifier{ResourceType: meshidentity_api.MeshIdentityType, Mesh: "default", Zone: "default", Name: "my-identity"},
		IdentitySourceConfigurer: identitySource("my-secret-name"),
	}
}

// kumaManagedWorkloadIdentity is managed by Kuma, so SAN matchers are derived
// from the trust domains of the mesh.
func kumaManagedWorkloadIdentity() *core_xds.WorkloadIdentity {
	identity := workloadIdentity()
	identity.ManagementMode = core_xds.KumaManagementMode
	return identity
}

// externalValidatorWorkloadIdentity delivers its own CA bundle instead of
// relying on the one Kuma generates.
func externalValidatorWorkloadIdentity() *core_xds.WorkloadIdentity {
	identity := workloadIdentity()
	identity.ExternalValidationSourceConfigurer = identitySource("ca-bundle")
	return identity
}

func trustDomains(names ...string) map[string][]xds_context.PEMBytes {
	cas := map[string][]xds_context.PEMBytes{}
	for i, name := range names {
		cas[name] = []xds_context.PEMBytes{xds_context.PEMBytes(fmt.Sprintf("ca-%d", i))}
	}
	return cas
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
			resourceSet.Add(getMeshServiceResources(proxy)...)

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
			caseName:    "no-workload-identity",
			meshBuilder: samples.MeshDefaultBuilder(),
		}),
		Entry("no policy = strict", testCase{
			caseName:         "no-policy",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: workloadIdentity(),
			noPolicy:         true,
		}),

		Entry("strict = a single kuma TLS filter chain", testCase{
			caseName:         "strict",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: workloadIdentity(),
		}),
		Entry("strict with tls version and ciphers", testCase{
			caseName:         "strict-with-tls-params",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: workloadIdentity(),
		}),
		Entry("strict with an external validator = identity delivers the CA bundle", testCase{
			caseName:         "strict-with-external-validator",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: externalValidatorWorkloadIdentity(),
		}),
		Entry("strict with MeshTrust = no SAN matchers, the identity isn't kuma managed", testCase{
			caseName:         "strict-with-mesh-trust",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: workloadIdentity(),
			casByTrustDomain: trustDomains("domain-1"),
		}),
		Entry("strict with MeshTrust and kuma managed identity = SAN matchers", testCase{
			caseName:         "strict-with-mesh-trust-kuma-managed",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaManagedWorkloadIdentity(),
			casByTrustDomain: trustDomains("domain-1"),
		}),
		Entry("strict with multiple MeshTrust and kuma managed identity = sorted SAN matchers", testCase{
			caseName:         "strict-with-multiple-mesh-trust-kuma-managed",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaManagedWorkloadIdentity(),
			// deliberately out of alphabetical order to verify SANs are sorted
			casByTrustDomain: trustDomains("domain-c", "domain-a", "domain-b"),
		}),
		Entry("strict with strict inbound ports feature = port filtering", testCase{
			caseName:         "strict-with-strict-inbound-ports",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: workloadIdentity(),
			features: xds_types.Features{
				xds_types.FeatureStrictInboundPorts: true,
			},
		}),
		Entry("strict with dualstack tproxy = ipv4 and ipv6 passthrough listeners", testCase{
			caseName:         "strict-with-dualstack-tproxy",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: workloadIdentity(),
			ipFamilyMode:     "dualstack",
		}),

		Entry("permissive = raw buffer, TLS and kuma TLS filter chains", testCase{
			caseName:         "permissive",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: workloadIdentity(),
		}),
		Entry("permissive with tls version and ciphers", testCase{
			caseName:         "permissive-with-tls-params",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: workloadIdentity(),
		}),
		Entry("permissive with an external validator = identity delivers the CA bundle", testCase{
			caseName:         "permissive-with-external-validator",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: externalValidatorWorkloadIdentity(),
		}),
		Entry("permissive with MeshTrust = no SAN matchers, the identity isn't kuma managed", testCase{
			caseName:         "permissive-with-mesh-trust",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: workloadIdentity(),
			casByTrustDomain: trustDomains("domain-1"),
		}),
		Entry("permissive with MeshTrust and kuma managed identity = SAN matchers", testCase{
			caseName:         "permissive-with-mesh-trust-kuma-managed",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaManagedWorkloadIdentity(),
			casByTrustDomain: trustDomains("domain-1"),
		}),
		Entry("permissive with multiple MeshTrust and kuma managed identity = sorted SAN matchers", testCase{
			caseName:         "permissive-with-multiple-mesh-trust-kuma-managed",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: kumaManagedWorkloadIdentity(),
			// deliberately out of alphabetical order to verify SANs are sorted
			casByTrustDomain: trustDomains("domain-c", "domain-a", "domain-b"),
		}),
		Entry("permissive with strict inbound ports feature = no port filtering", testCase{
			caseName:         "permissive-with-strict-inbound-ports",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: workloadIdentity(),
			features: xds_types.Features{
				xds_types.FeatureStrictInboundPorts: true,
			},
		}),
		Entry("permissive with dualstack tproxy = ipv4 and ipv6 passthrough listeners", testCase{
			caseName:         "permissive-with-dualstack-tproxy",
			meshBuilder:      samples.MeshDefaultBuilder(),
			workloadIdentity: workloadIdentity(),
			ipFamilyMode:     "dualstack",
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

// outgoingCluster is the outbound cluster the sidecar already has by the time
// MeshTLS runs. Its transport socket comes from the very builder the outbound
// generator uses, so the golden files show MeshTLS layering TLS params onto a
// real TLS context instead of an empty one. Without a workload identity there
// is no mTLS and therefore no transport socket to configure.
func outgoingCluster(proxy *core_xds.Proxy) *envoy_cluster.Cluster {
	builder := clusters.NewClusterBuilder(envoy_common.APIV3, outgoingMeshService.String())

	if proxy.WorkloadIdentity != nil {
		tlsContext, err := meshroute.UpstreamTLSContext(proxy, "outgoing", []string{"spiffe://default/outgoing"})
		Expect(err).ToNot(HaveOccurred())
		builder.Configure(clusters.UpstreamTLSContext(tlsContext))
	}

	return builder.MustBuild().(*envoy_cluster.Cluster)
}

func getMeshServiceResources(proxy *core_xds.Proxy) []*core_xds.Resource {
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
			Name:           outgoingMeshService.String(),
			Origin:         metadata.OriginOutbound,
			Resource:       outgoingCluster(proxy),
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
