package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	meshidentity_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
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

	Describe("GetSpiffeID", func() {
		type spiffeIDTestCase struct {
			name        string
			environment config_core.EnvironmentType
			labels      map[string]string
			expectedID  string
		}

		DescribeTable("should generate correct SPIFFE ID",
			func(tc spiffeIDTestCase) {
				// given
				meshIdentity := builders.MeshIdentity().
					WithName("test-identity").
					WithMesh("test-mesh").
					Build()

				meta := &test_model.ResourceMeta{
					Mesh:   "test-mesh",
					Name:   "test-dp",
					Labels: tc.labels,
				}

				trustDomain := "test-mesh.test-zone.mesh.local"

				// when
				spiffeID, err := meshIdentity.Spec.GetSpiffeID(trustDomain, meta, tc.environment)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(spiffeID).To(Equal(tc.expectedID))
			},
			Entry("Kubernetes environment with service account", spiffeIDTestCase{
				name:        "kubernetes with service account",
				environment: config_core.KubernetesEnvironment,
				labels: map[string]string{
					mesh_proto.KubeNamespaceTag: "test-namespace",
					metadata.KumaServiceAccount: "test-sa",
				},
				expectedID: "spiffe://test-mesh.test-zone.mesh.local/ns/test-namespace/sa/test-sa",
			}),
			Entry("Universal environment with workload", spiffeIDTestCase{
				name:        "universal with workload",
				environment: config_core.UniversalEnvironment,
				labels: map[string]string{
					metadata.KumaWorkload: "my-workload",
				},
				expectedID: "spiffe://test-mesh.test-zone.mesh.local/workload/my-workload",
			}),
		)

		It("should use custom SPIFFE ID path template for Kubernetes", func() {
			// given
			customPath := "/custom/{{ .Namespace }}/{{ .ServiceAccount }}"
			meshIdentity := builders.MeshIdentity().
				WithName("test-identity").
				WithMesh("test-mesh").
				WithSpiffeID("", customPath).
				Build()

			meta := &test_model.ResourceMeta{
				Mesh: "test-mesh",
				Name: "test-dp",
				Labels: map[string]string{
					mesh_proto.KubeNamespaceTag: "custom-namespace",
					metadata.KumaServiceAccount: "custom-sa",
				},
			}

			trustDomain := "test-mesh.test-zone.mesh.local"

			// when
			spiffeID, err := meshIdentity.Spec.GetSpiffeID(trustDomain, meta, config_core.KubernetesEnvironment)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(spiffeID).To(Equal("spiffe://test-mesh.test-zone.mesh.local/custom/custom-namespace/custom-sa"))
		})

		It("should use custom SPIFFE ID path template for Universal", func() {
			// given
			customPath := "/app/{{ .Workload }}"
			meshIdentity := builders.MeshIdentity().
				WithName("test-identity").
				WithMesh("test-mesh").
				WithSpiffeID("", customPath).
				Build()

			meta := &test_model.ResourceMeta{
				Mesh: "test-mesh",
				Name: "test-dp",
				Labels: map[string]string{
					metadata.KumaWorkload: "my-app",
				},
			}

			trustDomain := "test-mesh.test-zone.mesh.local"

			// when
			spiffeID, err := meshIdentity.Spec.GetSpiffeID(trustDomain, meta, config_core.UniversalEnvironment)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(spiffeID).To(Equal("spiffe://test-mesh.test-zone.mesh.local/app/my-app"))
		})
	})

	DescribeTable("SpiffeIDPathLabels",
		func(env config_core.EnvironmentType, pathTemplate *string, expected []string) {
			// given
			mi := &meshidentity_api.MeshIdentity{}
			if pathTemplate != nil {
				mi.SpiffeID = &meshidentity_api.SpiffeID{
					Path: pathTemplate,
				}
			}
			meta := &test_model.ResourceMeta{
				Mesh: "default",
				Name: "dp-1",
				Labels: map[string]string{
					metadata.KumaWorkload: "backend",
					"app":                 "backend",
				},
			}

			// when
			read := mi.SpiffeIDPathLabels(meta, env)

			// then
			if expected == nil {
				Expect(read).To(BeEmpty())
			} else {
				Expect(read).To(Equal(expected))
			}
		},
		Entry("nil SpiffeID on Kubernetes uses default ns/sa template", config_core.KubernetesEnvironment, nil,
			[]string{mesh_proto.KubeNamespaceTag, metadata.KumaServiceAccount}),
		Entry("nil SpiffeID on Universal falls back to default workload template", config_core.UniversalEnvironment, nil,
			[]string{metadata.KumaWorkload}),
		Entry("label function", config_core.UniversalEnvironment, pointer.To(`/workload/{{ label "kuma.io/workload" }}`),
			[]string{metadata.KumaWorkload}),
		Entry("root-scoped .Workload", config_core.UniversalEnvironment, pointer.To("/workload/{{ $.Workload }}"),
			[]string{metadata.KumaWorkload}),
		Entry(".Workload with trim markers", config_core.UniversalEnvironment, pointer.To("/workload/{{- .Workload -}}"),
			[]string{metadata.KumaWorkload}),
		Entry(".Workload through a pipeline", config_core.UniversalEnvironment, pointer.To(`/workload/{{ .Workload | printf "%s" }}`),
			[]string{metadata.KumaWorkload}),
		Entry("label read via a variable", config_core.UniversalEnvironment, pointer.To(`/workload/{{ $w := label "kuma.io/workload" }}{{ $w }}`),
			[]string{metadata.KumaWorkload}),
		Entry("label with a computed key", config_core.UniversalEnvironment, pointer.To(`/workload/{{ label (printf "kuma.io/%s" "workload") }}`),
			[]string{metadata.KumaWorkload}),
		Entry("other label", config_core.UniversalEnvironment, pointer.To(`/app/{{ label "app" }}`),
			[]string{"app"}),
		Entry("label missing on the dataplane is still reported, and so are reads after it", config_core.UniversalEnvironment, pointer.To(`/{{ label "missing" }}/{{ .Workload }}`),
			[]string{metadata.KumaWorkload, "missing"}),
		Entry("only the branch taken for the dataplane", config_core.UniversalEnvironment, pointer.To(`/{{ if eq .Workload "backend" }}{{ label "app" }}{{ else }}{{ label "other" }}{{ end }}`),
			[]string{"app", metadata.KumaWorkload}),
		Entry("static path", config_core.UniversalEnvironment, pointer.To("/static"), nil),
		Entry("empty path on Universal", config_core.UniversalEnvironment, pointer.To(""), nil),
	)
})
