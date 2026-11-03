package auth_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/xds/auth"
)

var _ = Describe("Binding", func() {
	labels := map[string]string{
		metadata.KumaWorkload: "backend",
		"app":                 "backend",
	}

	DescribeTable("ValidateIdentityLabels",
		func(binding auth.Binding, read []string, expectedErr string) {
			// when
			err := auth.ValidateIdentityLabels(binding, labels, read)

			// then
			if expectedErr == "" {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(MatchError(ContainSubstring(expectedErr)))
			}
		},
		Entry("workload bound by --workload",
			auth.Binding{Workload: "backend"}, []string{metadata.KumaWorkload}, ""),
		Entry("workload bound by a tag",
			auth.Binding{Tags: mesh_proto.MultiValueTagSet{metadata.KumaWorkload: {"backend": true, "frontend": true}}}, []string{metadata.KumaWorkload}, ""),
		Entry("workload unbound",
			auth.Binding{Tags: mesh_proto.MultiValueTagSet{"app": {"backend": true}}}, []string{metadata.KumaWorkload}, "requires a workload-bound token"),
		Entry("workload bound to another value",
			auth.Binding{Workload: "frontend"}, []string{metadata.KumaWorkload}, "requires a workload-bound token"),
		Entry("--workload and a tag that disagree",
			auth.Binding{Workload: "backend", Tags: mesh_proto.MultiValueTagSet{metadata.KumaWorkload: {"frontend": true}}}, []string{metadata.KumaWorkload}, "requires a workload-bound token"),
		Entry("other label bound by a tag",
			auth.Binding{Tags: mesh_proto.MultiValueTagSet{"app": {"backend": true}}}, []string{"app"}, ""),
		Entry("other label unbound",
			auth.Binding{Workload: "backend"}, []string{"app", metadata.KumaWorkload}, `"app" label "backend", which the token does not bind, so it requires a token with that tag`),
		Entry("label the dataplane doesn't declare",
			auth.Binding{}, []string{"missing"}, ""),
		Entry("all labels vouched for",
			auth.Binding{AllLabels: true}, []string{"app", metadata.KumaWorkload}, ""),
		Entry("nothing read",
			auth.Binding{}, nil, ""),
	)
})
