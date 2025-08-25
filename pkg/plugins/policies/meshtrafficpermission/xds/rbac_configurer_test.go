package xds_test

import (
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/xds"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("RBACConfigurer", func() {
	type testCase struct {
		inboundRules []*inbound.Rule
		stats        string
		expected     string
	}

	mtpOrigin := func(policyName string) common.Origin {
		return common.Origin{
			Resource: &test_model.ResourceMeta{
				Mesh: "default",
				Name: policyName,
				Labels: map[string]string{
					mesh_proto.ZoneTag:          "zone-1",
					mesh_proto.KubeNamespaceTag: "ns-1",
				},
			},
		}
	}

	DescribeTable("should generate proper envoy config",
		func(given testCase) {
			// given
			configurer := &xds.RBACConfigurer{
				InboundRules: given.inboundRules,
				StatsName:    given.stats,
			}
			res, err := listeners.NewFilterChainBuilder(envoy.APIV3, envoy.AnonymousResource).Build()
			Expect(err).ToNot(HaveOccurred())

			// when
			err = configurer.Configure(res.(*listenerv3.FilterChain))
			Expect(err).ToNot(HaveOccurred())

			// then
			actual, err := util_proto.ToYAML(res)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("allow all form mesh", testCase{
			stats: "allow_all_prefix",
			inboundRules: []*inbound.Rule{
				{
					Conf: &v1alpha1.Rule{
						Default: v1alpha1.RuleConf{
							Allow: &[]common_api.Match{
								{
									SpiffeId: &common_api.SpiffeIdMatch{
										Type:  common_api.PrefixMatchType,
										Value: "spiffeId://trust-domain.mesh/",
									},
								},
							},
						},
					},
					Origin: mtpOrigin("mtp-1"),
				},
			},
			expected: `
filters:
- name: envoy.filters.network.rbac
  typedConfig:
    '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
    matcher:
        matcherList:
            matchers:
                - onMatch:
                    action:
                        name: envoy.filters.rbac.action
                        typedConfig:
                            '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                            name: kri_mtp_default_zone-1_ns-1_mtp-1_
                  predicate:
                    singlePredicate:
                        input:
                            name: envoy.matching.inputs.uri_san
                            typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.matching.common_inputs.ssl.v3.UriSanInput
                        valueMatch:
                            prefix: spiffeId://trust-domain.mesh/
        onNoMatch:
            action:
                name: envoy.filters.rbac.action
                typedConfig:
                    '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                    action: DENY
                    name: default
    statPrefix: allow_all_prefix.`,
		}),
		Entry("deny all from mesh", testCase{
			stats: "deny_all_prefix",
			inboundRules: []*inbound.Rule{
				{
					Conf: &v1alpha1.Rule{
						Default: v1alpha1.RuleConf{
							Deny: &[]common_api.Match{
								{
									SpiffeId: &common_api.SpiffeIdMatch{
										Type:  common_api.PrefixMatchType,
										Value: "spiffeId://trust-domain.mesh/",
									},
								},
							},
						},
					},
					Origin: mtpOrigin("mtp-1"),
				},
			},
			expected: `
filters:
- name: envoy.filters.network.rbac
  typedConfig:
    '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
    matcher:
        matcherList:
            matchers:
                - onMatch:
                    action:
                        name: envoy.filters.rbac.action
                        typedConfig:
                            '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                            action: DENY
                            name: kri_mtp_default_zone-1_ns-1_mtp-1_
                  predicate:
                    singlePredicate:
                        input:
                            name: envoy.matching.inputs.uri_san
                            typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.matching.common_inputs.ssl.v3.UriSanInput
                        valueMatch:
                            prefix: spiffeId://trust-domain.mesh/
        onNoMatch:
            action:
                name: envoy.filters.rbac.action
                typedConfig:
                    '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                    action: DENY
                    name: default
    statPrefix: deny_all_prefix.`,
		}),
		Entry("allow multiple services", testCase{
			stats: "allow_2_services_prefix",
			inboundRules: []*inbound.Rule{
				{
					Conf: &v1alpha1.Rule{
						Default: v1alpha1.RuleConf{
							Allow: &[]common_api.Match{
								{
									SpiffeId: &common_api.SpiffeIdMatch{
										Type:  common_api.ExactMatchType,
										Value: "spiffeId://trust-domain.mesh/ns/backend/v1",
									},
								},
								{
									SpiffeId: &common_api.SpiffeIdMatch{
										Type:  common_api.ExactMatchType,
										Value: "spiffeId://trust-domain.mesh/ns/backend/v2",
									},
								},
							},
						},
					},
					Origin: mtpOrigin("mtp-1"),
				},
			},
			expected: `
filters:
- name: envoy.filters.network.rbac
  typedConfig:
    '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
    matcher:
        matcherList:
            matchers:
                - onMatch:
                    action:
                        name: envoy.filters.rbac.action
                        typedConfig:
                            '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                            name: kri_mtp_default_zone-1_ns-1_mtp-1_
                  predicate:
                    orMatcher:
                        predicate:
                            - singlePredicate:
                                input:
                                    name: envoy.matching.inputs.uri_san
                                    typedConfig:
                                        '@type': type.googleapis.com/envoy.extensions.matching.common_inputs.ssl.v3.UriSanInput
                                valueMatch:
                                    exact: spiffeId://trust-domain.mesh/ns/backend/v1
                            - singlePredicate:
                                input:
                                    name: envoy.matching.inputs.uri_san
                                    typedConfig:
                                        '@type': type.googleapis.com/envoy.extensions.matching.common_inputs.ssl.v3.UriSanInput
                                valueMatch:
                                    exact: spiffeId://trust-domain.mesh/ns/backend/v2
        onNoMatch:
            action:
                name: envoy.filters.rbac.action
                typedConfig:
                    '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                    action: DENY
                    name: default
    statPrefix: allow_2_services_prefix.`,
		}),
		Entry("rules from merged policies", testCase{
			stats: "rules_from_merged",
			inboundRules: []*inbound.Rule{
				{
					Conf: &v1alpha1.Rule{
						Default: v1alpha1.RuleConf{
							Deny: &[]common_api.Match{
								{
									SpiffeId: &common_api.SpiffeIdMatch{
										Type:  common_api.ExactMatchType,
										Value: "spiffeId://trust-domain.mesh/ns/backend/v1",
									},
								},
							},
						},
					},
					Origin: mtpOrigin("mtp-1"),
				},
				{
					Conf: &v1alpha1.Rule{
						Default: v1alpha1.RuleConf{
							Allow: &[]common_api.Match{
								{
									SpiffeId: &common_api.SpiffeIdMatch{
										Type:  common_api.PrefixMatchType,
										Value: "spiffeId://trust-domain.mesh/ns/backend",
									},
								},
							},
						},
					},
					Origin: mtpOrigin("mtp-2"),
				},
			},
			expected: `
filters:
- name: envoy.filters.network.rbac
  typedConfig:
    '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
    matcher:
        matcherList:
            matchers:
                - onMatch:
                    action:
                        name: envoy.filters.rbac.action
                        typedConfig:
                            '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                            action: DENY
                            name: kri_mtp_default_zone-1_ns-1_mtp-1_
                  predicate:
                    singlePredicate:
                        input:
                            name: envoy.matching.inputs.uri_san
                            typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.matching.common_inputs.ssl.v3.UriSanInput
                        valueMatch:
                            exact: spiffeId://trust-domain.mesh/ns/backend/v1
                - onMatch:
                    action:
                        name: envoy.filters.rbac.action
                        typedConfig:
                            '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                            name: kri_mtp_default_zone-1_ns-1_mtp-2_
                  predicate:
                    singlePredicate:
                        input:
                            name: envoy.matching.inputs.uri_san
                            typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.matching.common_inputs.ssl.v3.UriSanInput
                        valueMatch:
                            prefix: spiffeId://trust-domain.mesh/ns/backend
        onNoMatch:
            action:
                name: envoy.filters.rbac.action
                typedConfig:
                    '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                    action: DENY
                    name: default
    statPrefix: rules_from_merged.`,
		}),
		Entry("shadow deny rule", testCase{
			stats: "shadow_deny",
			inboundRules: []*inbound.Rule{
				{
					Conf: &v1alpha1.Rule{
						Default: v1alpha1.RuleConf{
							AllowWithShadowDeny: &[]common_api.Match{
								{
									SpiffeId: &common_api.SpiffeIdMatch{
										Type:  common_api.PrefixMatchType,
										Value: "spiffeId://trust-domain.mesh/",
									},
								},
							},
						},
					},
					Origin: mtpOrigin("mtp-1"),
				},
			},
			expected: `
filters:
- name: envoy.filters.network.rbac
  typedConfig:
    '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
    matcher:
        matcherList:
            matchers:
                - onMatch:
                    action:
                        name: envoy.filters.rbac.action
                        typedConfig:
                            '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                            name: kri_mtp_default_zone-1_ns-1_mtp-1_
                  predicate:
                    singlePredicate:
                        input:
                            name: envoy.matching.inputs.uri_san
                            typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.matching.common_inputs.ssl.v3.UriSanInput
                        valueMatch:
                            prefix: spiffeId://trust-domain.mesh/
        onNoMatch:
            action:
                name: envoy.filters.rbac.action
                typedConfig:
                    '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                    action: DENY
                    name: default
    shadowMatcher:
        matcherList:
            matchers:
                - onMatch:
                    action:
                        name: envoy.filters.rbac.action
                        typedConfig:
                            '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                            action: DENY
                            name: kri_mtp_default_zone-1_ns-1_mtp-1_
                  predicate:
                    singlePredicate:
                        input:
                            name: envoy.matching.inputs.uri_san
                            typedConfig:
                                '@type': type.googleapis.com/envoy.extensions.matching.common_inputs.ssl.v3.UriSanInput
                        valueMatch:
                            prefix: spiffeId://trust-domain.mesh/
        onNoMatch:
            action:
                name: envoy.filters.rbac.action
                typedConfig:
                    '@type': type.googleapis.com/envoy.config.rbac.v3.Action
                    action: DENY
                    name: default
    statPrefix: shadow_deny.`,
		}),
	)
})
