package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
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

	DescribeTable("RenderSpiffeID",
		func(env config_core.EnvironmentType, pathTemplate *string, expectedID string, expectedRead []string) {
			// given
			mi := meshidentity_api.NewMeshIdentityResource()
			mi.SetMeta(&test_model.ResourceMeta{Mesh: "default", Name: "identity"})
			if pathTemplate != nil {
				mi.Spec.SpiffeID = &meshidentity_api.SpiffeID{Path: pathTemplate}
			}
			meta := &test_model.ResourceMeta{
				Mesh: "default",
				Name: "dp-1",
				Labels: map[string]string{
					metadata.KumaWorkload:       "backend",
					"app":                       "backend",
					mesh_proto.KubeNamespaceTag: "demo",
					metadata.KumaServiceAccount: "backend-sa",
				},
			}

			// when
			spiffeID, read, err := mi.RenderSpiffeID(meta, "zone-1", env)

			// then
			if expectedID == "" {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).ToNot(HaveOccurred())
				Expect(spiffeID).To(Equal(expectedID))
			}
			if expectedRead == nil {
				Expect(read).To(BeEmpty())
			} else {
				Expect(read).To(Equal(expectedRead))
			}
		},
		Entry("nil SpiffeID on Kubernetes uses default ns/sa template", config_core.KubernetesEnvironment, nil,
			"spiffe://default.zone-1.mesh.local/ns/demo/sa/backend-sa", []string{mesh_proto.KubeNamespaceTag, metadata.KumaServiceAccount}),
		Entry("nil SpiffeID on Universal falls back to default workload template", config_core.UniversalEnvironment, nil,
			"spiffe://default.zone-1.mesh.local/workload/backend", []string{metadata.KumaWorkload}),
		Entry("label function", config_core.UniversalEnvironment, pointer.To(`/workload/{{ label "kuma.io/workload" }}`),
			"spiffe://default.zone-1.mesh.local/workload/backend", []string{metadata.KumaWorkload}),
		Entry("root-scoped .Workload", config_core.UniversalEnvironment, pointer.To("/workload/{{ $.Workload }}"),
			"spiffe://default.zone-1.mesh.local/workload/backend", []string{metadata.KumaWorkload}),
		Entry(".Workload with trim markers", config_core.UniversalEnvironment, pointer.To("/workload/{{- .Workload -}}"),
			"spiffe://default.zone-1.mesh.local/workload/backend", []string{metadata.KumaWorkload}),
		Entry(".Workload through a pipeline", config_core.UniversalEnvironment, pointer.To(`/workload/{{ .Workload | printf "%s" }}`),
			"spiffe://default.zone-1.mesh.local/workload/backend", []string{metadata.KumaWorkload}),
		Entry("label read via a variable", config_core.UniversalEnvironment, pointer.To(`/workload/{{ $w := label "kuma.io/workload" }}{{ $w }}`),
			"spiffe://default.zone-1.mesh.local/workload/backend", []string{metadata.KumaWorkload}),
		Entry("label with a computed key", config_core.UniversalEnvironment, pointer.To(`/workload/{{ label (printf "kuma.io/%s" "workload") }}`),
			"spiffe://default.zone-1.mesh.local/workload/backend", []string{metadata.KumaWorkload}),
		Entry("other label", config_core.UniversalEnvironment, pointer.To(`/app/{{ label "app" }}`),
			"spiffe://default.zone-1.mesh.local/app/backend", []string{"app"}),
		Entry("label missing on the dataplane fails the render and is still reported", config_core.UniversalEnvironment, pointer.To(`/{{ label "missing" }}/{{ .Workload }}`),
			"", []string{"missing"}),
		Entry("only the branch taken for the dataplane", config_core.UniversalEnvironment, pointer.To(`/{{ if eq .Workload "backend" }}{{ label "app" }}{{ else }}{{ label "other" }}{{ end }}`),
			"spiffe://default.zone-1.mesh.local/backend", []string{"app", metadata.KumaWorkload}),
		Entry("branch on the real trust domain", config_core.UniversalEnvironment, pointer.To(`{{ if eq .TrustDomain "default.zone-1.mesh.local" }}/{{ label "app" }}{{ else }}/static{{ end }}`),
			"spiffe://default.zone-1.mesh.local/backend", []string{"app"}),
		Entry("printing the inputs exposes no label", config_core.UniversalEnvironment, pointer.To(`/{{ . }}`),
			"spiffe://default.zone-1.mesh.local/spiffeIDInputs", nil),
		Entry("print exposes no label", config_core.UniversalEnvironment, pointer.To(`/{{ print . }}`),
			"spiffe://default.zone-1.mesh.local/spiffeIDInputs", nil),
		Entry("printf %#v exposes no label", config_core.UniversalEnvironment, pointer.To(`/{{ printf "%#v" . }}`),
			"spiffe://default.zone-1.mesh.local/spiffeIDInputs", nil),
		Entry("js exposes no label", config_core.UniversalEnvironment, pointer.To(`/{{ js . }}`),
			"spiffe://default.zone-1.mesh.local/spiffeIDInputs", nil),
		Entry("static path", config_core.UniversalEnvironment, pointer.To("/static"),
			"spiffe://default.zone-1.mesh.local/static", nil),
	)
})
