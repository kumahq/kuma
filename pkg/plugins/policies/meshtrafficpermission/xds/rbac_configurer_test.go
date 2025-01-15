package xds_test

import (
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("RBACConfigurer", func() {
	type testCase struct {
		rules    core_xds.Rules
		mesh     string
		stats    string
		expected string
	}

	DescribeTable("should generate proper envoy config",
		func(given testCase) {
			// given
			configurer := &xds.RBACConfigurer{
				Rules:     given.rules,
				Mesh:      given.mesh,
				StatsName: given.stats,
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
		Entry("allow all", testCase{
			stats: "allow_all_prefix",
			mesh:  "allow_all_mesh",
			rules: []*core_xds.Rule{
				{
					Subset: []subsetutils.Tag{},
					Conf: v1alpha1.Conf{
						Action: v1alpha1.Allow,
					},
				},
			},
			expected: `
filters:
  - name: envoy.filters.network.rbac
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
      rules:
          policies:
              MeshTrafficPermission:
                  permissions:
                      - any: true
                  principals:
                      - any: true
      statPrefix: allow_all_prefix.`,
		}),
		Entry("deny all", testCase{
			stats: "deny_all_prefix",
			mesh:  "deny_all_mesh",
			rules: []*core_xds.Rule{
				{
					Subset: []subsetutils.Tag{},
					Conf: v1alpha1.Conf{
						Action: v1alpha1.Deny,
					},
				},
			},
			expected: `
filters:
  - name: envoy.filters.network.rbac
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
      rules: {}
      statPrefix: deny_all_prefix.`,
		}),
		Entry("allow backend-v1 and web in us-east", testCase{
			stats: "allow_2_services_prefix",
			mesh:  "allow_2_service_mesh",
			rules: []*core_xds.Rule{
				{
					Subset: []subsetutils.Tag{
						{Key: "kuma.io/service", Value: "backend"},
						{Key: "version", Value: "v1"},
					},
					Conf: v1alpha1.Conf{
						Action: v1alpha1.Allow,
					},
				},
				{
					Subset: []subsetutils.Tag{
						{Key: "kuma.io/service", Value: "web"},
						{Key: "kuma.io/zone", Value: "us-east"},
					},
					Conf: v1alpha1.Conf{
						Action: v1alpha1.Allow,
					},
				},
			},
			expected: `
filters:
  - name: envoy.filters.network.rbac
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
      rules:
          policies:
              MeshTrafficPermission:
                  permissions:
                      - any: true
                  principals:
                      - andIds:
                          ids:
                              - authenticated:
                                  principalName:
                                      exact: spiffe://allow_2_service_mesh/backend
                              - authenticated:
                                  principalName:
                                      exact: kuma://version/v1
                      - andIds:
                          ids:
                              - authenticated:
                                  principalName:
                                      exact: spiffe://allow_2_service_mesh/web
                              - authenticated:
                                  principalName:
                                      exact: kuma://kuma.io/zone/us-east
      statPrefix: allow_2_services_prefix.`,
		}),
		Entry("allow rule with negation in kuma tag", testCase{
			stats: "allow_negation_prefix",
			mesh:  "allow_negation_mesh",
			rules: []*core_xds.Rule{
				{
					Subset: []subsetutils.Tag{
						{Key: "kuma.io/service", Value: "backend"},
						{Key: "version", Value: "v2", Not: true},
					},
					Conf: v1alpha1.Conf{
						Action: v1alpha1.Allow,
					},
				},
			},
			expected: `
filters:
  - name: envoy.filters.network.rbac
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
      rules:
          policies:
              MeshTrafficPermission:
                  permissions:
                      - any: true
                  principals:
                      - andIds:
                          ids:
                              - authenticated:
                                  principalName:
                                      exact: spiffe://allow_negation_mesh/backend
                              - notId:
                                  authenticated:
                                      principalName:
                                          exact: kuma://version/v2
      statPrefix: allow_negation_prefix.`,
		}),
		Entry("allow rule with negation in service tag", testCase{
			stats: "allow_negation_prefix",
			mesh:  "allow_negation_mesh",
			rules: []*core_xds.Rule{
				{
					Subset: []subsetutils.Tag{
						{Key: "kuma.io/service", Value: "backend", Not: true},
						{Key: "version", Value: "v2"},
					},
					Conf: v1alpha1.Conf{
						Action: v1alpha1.Allow,
					},
				},
			},
			expected: `
filters:
  - name: envoy.filters.network.rbac
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
      rules:
          policies:
              MeshTrafficPermission:
                  permissions:
                      - any: true
                  principals:
                      - andIds:
                          ids:
                              - notId:
                                  authenticated:
                                      principalName:
                                          exact: spiffe://allow_negation_mesh/backend
                              - authenticated:
                                  principalName:
                                      exact: kuma://version/v2
      statPrefix: allow_negation_prefix.`,
		}),
		Entry("no rules", testCase{
			stats: "no",
			mesh:  "nothing",
			rules: []*core_xds.Rule{},
			expected: `
filters:
  - name: envoy.filters.network.rbac
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
      rules: {}
      statPrefix: no.`,
		}),
		Entry("shadow deny rule", testCase{
			stats: "shadow_deny_prefix",
			mesh:  "shadow_deny_mesh",
			rules: []*core_xds.Rule{
				{
					Subset: []subsetutils.Tag{
						{Key: "kuma.io/service", Value: "backend"},
					},
					Conf: v1alpha1.Conf{
						Action: v1alpha1.AllowWithShadowDeny,
					},
				},
			},
			expected: `
filters:
  - name: envoy.filters.network.rbac
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
      rules:
          policies:
              MeshTrafficPermission:
                  permissions:
                      - any: true
                  principals:
                      - authenticated:
                          principalName:
                              exact: spiffe://shadow_deny_mesh/backend
      shadowRules:
          action: DENY
          policies:
              MeshTrafficPermission:
                  permissions:
                      - any: true
                  principals:
                      - authenticated:
                          principalName:
                              exact: spiffe://shadow_deny_mesh/backend
      statPrefix: shadow_deny_prefix.`,
		}),
	)
})
