package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
)

var _ = Describe("MeshServiceResource.Hash()", func() {
	newMeshService := func() *api.MeshServiceResource {
		return builders.MeshService().
			WithName("backend").
			WithMesh("default").
			WithLabels(map[string]string{
				mesh_proto.ZoneTag: "zone-1",
				"team":             "infra",
			}).
			AddIntPort(80, 8080, core_meta.ProtocolHTTP).
			WithKumaVIP("10.0.0.1").
			WithTLSStatus(api.TLSReady).
			Build()
	}

	It("is stable when only the resourceVersion changes", func() {
		original := newMeshService()
		originalHash := original.Hash()

		changed := newMeshService()
		changed.Meta.(*test_model.ResourceMeta).Version = "2"

		Expect(changed.Hash()).To(Equal(originalHash))
	})

	It("is stable when only the DataplaneProxies stats change", func() {
		original := newMeshService()
		originalHash := original.Hash()

		changed := newMeshService()
		changed.Status.DataplaneProxies = api.DataplaneProxies{Connected: 3, Healthy: 2, Total: 3}

		Expect(changed.Hash()).To(Equal(originalHash))
	})

	It("changes when a label is added", func() {
		original := newMeshService()
		originalHash := original.Hash()

		changed := newMeshService()
		changed.Meta.(*test_model.ResourceMeta).Labels["new-label"] = "value"

		Expect(changed.Hash()).NotTo(Equal(originalHash))
	})

	It("changes when the spec changes", func() {
		original := newMeshService()
		originalHash := original.Hash()

		changed := newMeshService()
		changed.Spec.Ports[0].Port = 8081

		Expect(changed.Hash()).NotTo(Equal(originalHash))
	})

	It("changes when the VIP changes", func() {
		original := newMeshService()
		originalHash := original.Hash()

		changed := newMeshService()
		changed.Status.VIPs[0].IP = "10.0.0.2"

		Expect(changed.Hash()).NotTo(Equal(originalHash))
	})

	It("changes when the TLS status changes", func() {
		original := newMeshService()
		originalHash := original.Hash()

		changed := newMeshService()
		changed.Status.TLS.Status = api.TLSNotReady

		Expect(changed.Hash()).NotTo(Equal(originalHash))
	})
})
