package kri_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	// register core resource types (Dataplane, *Overview, ...) in the global registry
	_ "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	rest_v1alpha1 "github.com/kumahq/kuma/v3/pkg/core/resources/model/rest/v1alpha1"
)

var _ = Describe("FromResourceMetaE", func() {
	meta := rest_v1alpha1.ResourceMeta{
		Mesh: "kuma-runner",
		Name: "grpc-service-84c8b47975-kg5xx_4455fx6cb5fd2494.kuma-system",
		Labels: map[string]string{
			mesh_proto.ZoneTag:          "east",
			mesh_proto.KubeNamespaceTag: "kuma-runner-ns",
			mesh_proto.DisplayName:      "grpc-service-84c8b47975-kg5xx",
		},
	}

	It("should resolve an overview to the underlying resource KRI", func() {
		// Overview shares the identity of the underlying resource: its KRI must
		// point at the actual Dataplane (kri_dp_...), not the overview type.
		id, err := kri.FromResourceMetaE(meta, "DataplaneOverview")
		Expect(err).ToNot(HaveOccurred())
		Expect(id.String()).To(Equal("kri_dp_kuma-runner_east_kuma-runner-ns_grpc-service-84c8b47975-kg5xx_"))
	})

	It("should produce the same KRI for the overview and its base resource", func() {
		overview, err := kri.FromResourceMetaE(meta, "DataplaneOverview")
		Expect(err).ToNot(HaveOccurred())
		base, err := kri.FromResourceMetaE(meta, "Dataplane")
		Expect(err).ToNot(HaveOccurred())
		Expect(overview.String()).To(Equal(base.String()))
	})
})
