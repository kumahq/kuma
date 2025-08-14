package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/api/common/v1alpha1"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
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
			identity, found := meshidentity_api.Matched(given.labels, given.meshIdentities)

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
})
