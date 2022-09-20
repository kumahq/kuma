package v3_test

import (
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
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
			configurer := &v3.RBACConfigurer{
				Rules:     given.rules,
				Mesh:      given.mesh,
				StatsName: given.stats,
			}
			res, err := listeners.NewFilterChainBuilder(envoy.APIV3).Build()
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
					Subset: []core_xds.Tag{},
					Conf: &v1alpha1.MeshTrafficPermission_Conf{
						Action: "ALLOW",
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
					Subset: []core_xds.Tag{},
					Conf: &v1alpha1.MeshTrafficPermission_Conf{
						Action: "DENY",
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
					Subset: []core_xds.Tag{
						{Key: "kuma.io/service", Value: "backend"},
						{Key: "version", Value: "v1"},
					},
					Conf: &v1alpha1.MeshTrafficPermission_Conf{
						Action: "ALLOW",
					},
				},
				{
					Subset: []core_xds.Tag{
						{Key: "kuma.io/service", Value: "web"},
						{Key: "kuma.io/zone", Value: "us-east"},
					},
					Conf: &v1alpha1.MeshTrafficPermission_Conf{
						Action: "ALLOW",
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
					Subset: []core_xds.Tag{
						{Key: "kuma.io/service", Value: "backend"},
						{Key: "version", Value: "v2", Not: true},
					},
					Conf: &v1alpha1.MeshTrafficPermission_Conf{
						Action: "ALLOW",
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
					Subset: []core_xds.Tag{
						{Key: "kuma.io/service", Value: "backend", Not: true},
						{Key: "version", Value: "v2"},
					},
					Conf: &v1alpha1.MeshTrafficPermission_Conf{
						Action: "ALLOW",
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
		Entry("shadow allow rule", testCase{
			stats: "shadow_allow_prefix",
			mesh:  "shadow_allow_mesh",
			rules: []*core_xds.Rule{
				{
					Subset: []core_xds.Tag{
						{Key: "kuma.io/service", Value: "backend"},
					},
					Conf: &v1alpha1.MeshTrafficPermission_Conf{
						Action: "DENY_WITH_SHADOW_ALLOW",
					},
				},
			},
			expected: `
filters:
  - name: envoy.filters.network.rbac
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.filters.network.rbac.v3.RBAC
      rules: {}
      shadowRules:
          policies:
              ShadowMeshTrafficPermission:
                  permissions:
                      - any: true
                  principals:
                      - authenticated:
                          principalName:
                              exact: spiffe://shadow_allow_mesh/backend
      statPrefix: shadow_allow_prefix.`,
		}),
	)
})
