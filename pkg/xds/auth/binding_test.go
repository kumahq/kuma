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
		metadata.KumaWorkload:               "backend",
		"app":                               "backend",
		mesh_proto.ZoneTag:                  "zone-1",
		metadata.KumaMeshLabel:              "default",
		mesh_proto.DisplayName:              "backend-1",
		mesh_proto.ListenerZoneIngressLabel: "enabled",
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
		Entry("workload proven",
			auth.Binding{Labels: mesh_proto.MultiValueTagSet{metadata.KumaWorkload: {"backend": true, "frontend": true}}}, []string{metadata.KumaWorkload}, ""),
		Entry("workload not proven",
			auth.Binding{Labels: mesh_proto.MultiValueTagSet{"app": {"backend": true}}}, []string{metadata.KumaWorkload}, "requires a workload-bound token"),
		Entry("workload proven with another value",
			auth.Binding{Labels: mesh_proto.MultiValueTagSet{metadata.KumaWorkload: {"frontend": true}}}, []string{metadata.KumaWorkload}, "requires a workload-bound token"),
		Entry("other label proven",
			auth.Binding{Labels: mesh_proto.MultiValueTagSet{"app": {"backend": true}}}, []string{"app"}, ""),
		Entry("other label not proven",
			auth.Binding{Labels: mesh_proto.MultiValueTagSet{metadata.KumaWorkload: {"backend": true}}}, []string{"app", metadata.KumaWorkload}, `"app" label "backend", which the token does not bind, so it requires a token with that tag`),
		Entry("label the control plane chooses",
			auth.Binding{}, []string{mesh_proto.ZoneTag, metadata.KumaMeshLabel}, ""),
		Entry("control-plane-owned label computed from the proxy's name",
			auth.Binding{}, []string{mesh_proto.DisplayName}, "requires a name-bound token"),
		Entry("control-plane-owned label computed from the proxy's listeners",
			auth.Binding{}, []string{mesh_proto.ListenerZoneIngressLabel}, "requires a token with that tag"),
		Entry("display name proven",
			auth.Binding{Labels: mesh_proto.MultiValueTagSet{mesh_proto.DisplayName: {"backend-1": true}}}, []string{mesh_proto.DisplayName}, ""),
		Entry("label the dataplane doesn't declare",
			auth.Binding{}, []string{"missing"}, ""),
		Entry("namespace scope trusts labels it doesn't prove",
			auth.Binding{Scope: auth.ScopeNamespace}, []string{"app", metadata.KumaWorkload}, ""),
		Entry("namespace scope still checks labels it proves",
			auth.Binding{Labels: mesh_proto.MultiValueTagSet{"app": {"frontend": true}}, Scope: auth.ScopeNamespace}, []string{"app"}, "requires a token with that tag"),
		Entry("unauthenticated",
			auth.Binding{Unauthenticated: true}, []string{"app", metadata.KumaWorkload}, ""),
		Entry("nothing read",
			auth.Binding{}, nil, ""),
	)
})
