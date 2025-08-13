package naming_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/naming"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type fakeResource struct{}

func (f *fakeResource) GetMeta() core_model.ResourceMeta          { return nil }
func (f *fakeResource) SetMeta(core_model.ResourceMeta)           {}
func (f *fakeResource) GetSpec() core_model.ResourceSpec          { return nil }
func (f *fakeResource) SetSpec(core_model.ResourceSpec) error     { return nil }
func (f *fakeResource) GetStatus() core_model.ResourceStatus      { return nil }
func (f *fakeResource) SetStatus(core_model.ResourceStatus) error { return nil }
func (f *fakeResource) Descriptor() core_model.ResourceTypeDescriptor {
	return core_model.ResourceTypeDescriptor{Name: "UnknownType"}
}

var _ = Describe("ContextualInboundName", func() {
	It("should return error when resource is nil", func() {
		// when
		name, err := naming.ContextualInboundName(nil, "http")

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("resource is nil"))
		Expect(name).To(Equal(""))
	})

	It("should return error when resource type is not in registry", func() {
		// given
		r := &fakeResource{}

		// when
		name, err := naming.ContextualInboundName(r, "section-x")

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("cannot build contextual inbound name"))
		Expect(err.Error()).To(ContainSubstring("UnknownType"))
		Expect(err.Error()).To(ContainSubstring("section-x"))
		Expect(err.Error()).To(ContainSubstring("type not found in global registry"))
		Expect(name).To(Equal(""))
	})

	It("should build name for registered type with string section", func() {
		// given
		r := core_mesh.NewDataplaneResource()

		// when
		name, err := naming.ContextualInboundName(r, "http")

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(name).To(Equal("self_inbound_dp_http"))
	})

	It("should build name for registered type with numeric section", func() {
		// given
		r := core_mesh.NewDataplaneResource()

		// when
		name, err := naming.ContextualInboundName(r, uint32(8080))

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(name).To(Equal("self_inbound_dp_8080"))
	})

	It("MustContextualInboundName should return result on success", func() {
		// given
		r := core_mesh.NewDataplaneResource()

		// when/then
		Expect(naming.MustContextualInboundName(r, "http")).To(Equal("self_inbound_dp_http"))
	})

	It("MustContextualInboundName should panic on error", func() {
		// given
		r := &fakeResource{}

		// when/then
		Expect(func() { _ = naming.MustContextualInboundName(r, "x") }).To(Panic())
	})
})
