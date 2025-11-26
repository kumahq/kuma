package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/api/common/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

var _ = Describe("MeshIdentity Helper", func() {
	type testCase struct {
		labels               map[string]string
		meshIdentities       []*meshidentity_api.MeshIdentityResource
		expectedIdentityName string
	}
	DescribeTable("Matched",
		func(given testCase) {
			// when
			identity, found := meshidentity_api.BestMatched(given.labels, given.meshIdentities)

			// then
			if found {
				Expect(identity.Meta.GetName()).To(Equal(given.expectedIdentityName))
			} else {
				Expect(identity).To(BeNil())
			}
		},
		Entry("select the most specific identity", testCase{
			labels: map[string]string{
				"app": "test-app",
			},
			meshIdentities: []*meshidentity_api.MeshIdentityResource{
				builders.MeshIdentity().WithName("not-matching-1").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app":     "test-app",
						"version": "v1",
					},
				}).Build(),
				builders.MeshIdentity().WithName("matching-all").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{},
				}).Build(),
				builders.MeshIdentity().WithName("matching-specific").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app": "test-app",
					},
				}).Build(),
			},
			expectedIdentityName: "matching-specific",
		}),
		Entry("select the most specific identity with 2 tags", testCase{
			labels: map[string]string{
				"app":     "test-app",
				"version": "v1",
			},
			meshIdentities: []*meshidentity_api.MeshIdentityResource{
				builders.MeshIdentity().WithName("matching-v1").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app":     "test-app",
						"version": "v1",
					},
				}).Build(),
				builders.MeshIdentity().WithName("matching-all").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{},
				}).Build(),
				builders.MeshIdentity().WithName("matching-specific").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app": "test-app",
					},
				}).Build(),
			},
			expectedIdentityName: "matching-v1",
		}),
		Entry("select matching all", testCase{
			labels: map[string]string{
				"app":     "demo-app",
				"version": "v1",
			},
			meshIdentities: []*meshidentity_api.MeshIdentityResource{
				builders.MeshIdentity().WithName("not-matching-1").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app":     "test-app",
						"version": "v1",
					},
				}).Build(),
				builders.MeshIdentity().WithName("matching-all").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{},
				}).Build(),
				builders.MeshIdentity().WithName("matching-specific").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app": "test-app",
					},
				}).Build(),
			},
			expectedIdentityName: "matching-all",
		}),
		Entry("select matching with the ascending name", testCase{
			labels: map[string]string{
				"app":     "demo-app",
				"version": "v1",
			},
			meshIdentities: []*meshidentity_api.MeshIdentityResource{
				builders.MeshIdentity().WithName("b-matching-v1").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app":     "demo-app",
						"version": "v1",
					},
				}).Build(),
				builders.MeshIdentity().WithName("a-matching-v1").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app":     "demo-app",
						"version": "v1",
					},
				}).Build(),
			},
			expectedIdentityName: "a-matching-v1",
		}),
		Entry("select matching with the ascending name when 3 tags", testCase{
			labels: map[string]string{
				"app":     "test-app",
				"version": "v1",
				"env":     "prod",
			},
			meshIdentities: []*meshidentity_api.MeshIdentityResource{
				builders.MeshIdentity().WithName("b-matching-v1").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app":     "test-app",
						"version": "v1",
					},
				}).Build(),
				builders.MeshIdentity().WithName("a-matching-v1").WithSelector(&v1alpha1.LabelSelector{
					MatchLabels: &map[string]string{
						"app": "test-app",
						"env": "prod",
					},
				}).Build(),
			},
			expectedIdentityName: "a-matching-v1",
		}),
		Entry("no identities", testCase{
			labels: map[string]string{
				"app":     "demo-app",
				"version": "v1",
			},
			meshIdentities: []*meshidentity_api.MeshIdentityResource{},
		}),
	)

	DescribeTable("UsesWorkloadLabel",
		func(pathTemplate *string, expected bool) {
			// given
			mi := &meshidentity_api.MeshIdentity{}
			if pathTemplate != nil {
				mi.SpiffeID = &meshidentity_api.SpiffeID{
					Path: pathTemplate,
				}
			}

			// when
			result := mi.UsesWorkloadLabel()

			// then
			Expect(result).To(Equal(expected))
		},
		Entry("nil SpiffeID", nil, false),
		Entry("uses workload label", pointer.To("/ns/{{ .Namespace }}/workload/{{ label \"kuma.io/workload\" }}"), true),
		Entry("uses workload label with extra spaces", pointer.To("/workload/{{  label  \"kuma.io/workload\"  }}"), true),
		Entry("does not use workload label", pointer.To("/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"), false),
		Entry("uses different label", pointer.To("/ns/{{ .Namespace }}/label/{{ label \"app\" }}"), false),
		Entry("empty path", pointer.To(""), false),
	)
})
