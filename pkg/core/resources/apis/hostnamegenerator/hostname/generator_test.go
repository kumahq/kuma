package hostname_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/hostnamegenerator/hostname"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
)

var _ = Describe("EvaluateTemplate", func() {
	type testCase struct {
		localZone string
		template  string
		meta      core_model.ResourceMeta
		expected  string
		errMsg    string
	}

	DescribeTable("should evaluate template",
		func(given testCase) {
			result, err := hostname.EvaluateTemplate(given.localZone, given.template, given.meta)
			if given.errMsg != "" {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(given.errMsg))
			} else {
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(given.expected))
			}
		},
		Entry("basic name template", testCase{
			template: "{{ .Name }}.mesh",
			meta: &test_model.ResourceMeta{
				Name: "backend",
				Mesh: "default",
			},
			expected: "backend.mesh",
		}),
		Entry("name with namespace", testCase{
			template: "{{ .Name }}.{{ .Namespace }}.mesh",
			meta: &test_model.ResourceMeta{
				Name: "backend",
				Mesh: "default",
				Labels: map[string]string{
					mesh_proto.KubeNamespaceTag: "my-ns",
				},
			},
			expected: "backend.my-ns.mesh",
		}),
		Entry("uses K8s name component over resource name", testCase{
			template: "{{ .Name }}.mesh",
			meta: &test_model.ResourceMeta{
				Name:           "backend.my-ns",
				Mesh:           "default",
				NameExtensions: core_model.ResourceNameExtensions{core_model.K8sNameComponent: "backend"},
			},
			expected: "backend.mesh",
		}),
		Entry("zone from label", testCase{
			localZone: "local",
			template:  "{{ .Name }}.{{ .Zone }}.mesh",
			meta: &test_model.ResourceMeta{
				Name: "backend",
				Mesh: "default",
				Labels: map[string]string{
					mesh_proto.ZoneTag: "zone-1",
				},
			},
			expected: "backend.zone-1.mesh",
		}),
		Entry("zone falls back to localZone", testCase{
			localZone: "local",
			template:  "{{ .Name }}.{{ .Zone }}.mesh",
			meta: &test_model.ResourceMeta{
				Name: "backend",
				Mesh: "default",
			},
			expected: "backend.local.mesh",
		}),
		Entry("display name from label", testCase{
			template: "{{ .DisplayName }}.mesh",
			meta: &test_model.ResourceMeta{
				Name: "backend.my-ns",
				Mesh: "default",
				Labels: map[string]string{
					mesh_proto.DisplayName: "backend",
				},
			},
			expected: "backend.mesh",
		}),
		Entry("label function", testCase{
			template: `{{ label "app" }}.mesh`,
			meta: &test_model.ResourceMeta{
				Name: "backend",
				Mesh: "default",
				Labels: map[string]string{
					"app": "my-app",
				},
			},
			expected: "my-app.mesh",
		}),
		Entry("label function missing key", testCase{
			template: `{{ label "missing" }}.mesh`,
			meta: &test_model.ResourceMeta{
				Name: "backend",
				Mesh: "default",
			},
			errMsg: "pre evaluation of template with parameters failed",
		}),
		Entry("invalid template syntax", testCase{
			template: "{{ .Name[0 }}.mesh",
			meta: &test_model.ResourceMeta{
				Name: "backend",
				Mesh: "default",
			},
			errMsg: "failed compiling gotemplate error",
		}),
		Entry("generated hostname starts with dot", testCase{
			template: ".{{ .Name }}.mesh",
			meta: &test_model.ResourceMeta{
				Name: "backend",
				Mesh: "default",
			},
			errMsg: "is not a valid DNS name",
		}),
		Entry("generated hostname has uppercase", testCase{
			template: "{{ .Name }}.Mesh",
			meta: &test_model.ResourceMeta{
				Name: "Backend",
				Mesh: "default",
			},
			errMsg: "is not a valid DNS name",
		}),
		Entry("generated hostname has consecutive dots", testCase{
			template: "{{ .Name }}..mesh",
			meta: &test_model.ResourceMeta{
				Name: "backend",
				Mesh: "default",
			},
			errMsg: "is not a valid DNS name",
		}),
	)
})
