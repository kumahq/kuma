package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	hostnamegenerator_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
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

	It("changes when only the resourceVersion changes", func() {
		original := newMeshService()
		originalHash := original.Hash()

		changed := newMeshService()
		changed.Meta.(*test_model.ResourceMeta).Version = "2"

		Expect(changed.Hash()).NotTo(Equal(originalHash))
	})

	It("changes when DataplaneProxies stats change together with resourceVersion", func() {
		original := newMeshService()
		originalHash := original.Hash()

		changed := newMeshService()
		changed.Status.DataplaneProxies = api.DataplaneProxies{Connected: 3, Healthy: 2, Total: 3}
		changed.Meta.(*test_model.ResourceMeta).Version = "2"

		Expect(changed.Hash()).NotTo(Equal(originalHash))
	})

	It("treats nil spec and status the same as empty values", func() {
		meta := &test_model.ResourceMeta{
			Name: "backend",
			Mesh: "default",
		}

		withNilFields := &api.MeshServiceResource{Meta: meta}
		withEmptyFields := api.NewMeshServiceResource()
		withEmptyFields.SetMeta(meta)

		Expect(func() { _ = withNilFields.Hash() }).ToNot(Panic())
		Expect(func() { _ = withNilFields.XDSHash() }).ToNot(Panic())
		Expect(withNilFields.Hash()).To(Equal(withEmptyFields.Hash()))
		Expect(withNilFields.XDSHash()).To(Equal(withEmptyFields.XDSHash()))
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

var _ = Describe("MeshServiceResource.XDSHash()", func() {
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
		originalHash := original.XDSHash()

		changed := newMeshService()
		changed.Meta.(*test_model.ResourceMeta).Version = "2"

		Expect(changed.XDSHash()).To(Equal(originalHash))
	})

	It("is stable when only the DataplaneProxies stats change", func() {
		original := newMeshService()
		originalHash := original.XDSHash()

		changed := newMeshService()
		changed.Status.DataplaneProxies = api.DataplaneProxies{Connected: 3, Healthy: 2, Total: 3}
		changed.Meta.(*test_model.ResourceMeta).Version = "2"

		Expect(changed.XDSHash()).To(Equal(originalHash))
	})

	It("changes when an xDS-relevant status field changes", func() {
		original := newMeshService()
		originalHash := original.XDSHash()

		changed := newMeshService()
		changed.Status.Addresses = append(changed.Status.Addresses, hostnamegenerator_api.Address{Hostname: "backend.shadow.mesh"})

		Expect(changed.XDSHash()).NotTo(Equal(originalHash))
	})
})
