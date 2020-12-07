package universal_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/resources/model"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/plugins/runtime/universal"
)

type countingManager struct {
	core_manager.ResourceManager
	updated int
}

func (c *countingManager) Update(ctx context.Context, resource model.Resource, optionsFunc ...store.UpdateOptionsFunc) error {
	c.updated++
	return c.ResourceManager.Update(ctx, resource, optionsFunc...)
}

var _ = Describe("UpdateOutbound", func() {
	var rm *countingManager
	vips := map[string]string{
		"service-1": "240.0.0.1",
		"service-2": "240.0.0.2",
	}

	BeforeEach(func() {
		rm = &countingManager{ResourceManager: core_manager.NewResourceManager(memory.NewStore())}

		err := rm.Create(context.Background(), &mesh.MeshResource{}, store.CreateByKey(model.DefaultMesh, model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	Context("transparentProxying enabled", func() {
		BeforeEach(func() {
			err := rm.Create(context.Background(), &mesh.DataplaneResource{
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port: 8080,
								Tags: map[string]string{
									"kuma.io/service": "service-1",
								},
							},
						},
						TransparentProxying: &mesh_proto.Dataplane_Networking_TransparentProxying{
							RedirectPortInbound:  15006,
							RedirectPortOutbound: 15001,
						},
					},
				},
			}, store.CreateByKey("dp-1", "default"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should update dataplane outbounds when new service is added", func() {
			// when
			err := rm.Create(context.Background(), &mesh.DataplaneResource{
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port: 8081,
								Tags: map[string]string{
									"kuma.io/service": "service-2",
								},
							},
						},
					},
				},
			}, store.CreateByKey("dp-2", "default"))
			Expect(err).ToNot(HaveOccurred())
			// and
			err = universal.UpdateOutbounds(context.Background(), rm, vips)
			Expect(err).ToNot(HaveOccurred())

			// then
			dp1 := &mesh.DataplaneResource{}
			err = rm.Get(context.Background(), dp1, store.GetByKey("dp-1", "default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(dp1.Spec.Networking.Outbound).To(HaveLen(1))
			Expect(dp1.Spec.Networking.Outbound[0].Tags["kuma.io/service"]).To(Equal("service-2"))
			Expect(dp1.Spec.Networking.Outbound[0].Address).To(Equal("240.0.0.2"))
		})

		It("should not update dataplane outbounds when new service is added to another mesh", func() {
			// when
			err := rm.Create(context.Background(), &mesh.MeshResource{}, store.CreateByKey("another-mesh", model.NoMesh))
			Expect(err).ToNot(HaveOccurred())
			// and
			err = rm.Create(context.Background(), &mesh.DataplaneResource{
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port: 8081,
								Tags: map[string]string{
									"kuma.io/service": "service-2",
								},
							},
						},
					},
				},
			}, store.CreateByKey("dp-2", "another-mesh"))
			Expect(err).ToNot(HaveOccurred())
			// and
			err = universal.UpdateOutbounds(context.Background(), rm, vips)
			Expect(err).ToNot(HaveOccurred())
			// then
			dp1 := &mesh.DataplaneResource{}
			err = rm.Get(context.Background(), dp1, store.GetByKey("dp-1", "default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(dp1.Spec.Networking.Outbound).To(HaveLen(0))
		})

		Context("outbounds already updated", func() {
			BeforeEach(func() {
				err := rm.Create(context.Background(), &mesh.DataplaneResource{
					Spec: &mesh_proto.Dataplane{
						Networking: &mesh_proto.Dataplane_Networking{
							Address: "127.0.0.1",
							Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
								{
									Port: 8081,
									Tags: map[string]string{
										"kuma.io/service": "service-2",
									},
								},
							},
						},
					},
				}, store.CreateByKey("dp-2", "default"))
				Expect(err).ToNot(HaveOccurred())

				err = universal.UpdateOutbounds(context.Background(), rm, vips)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not update outbounds if they are not changed", func() {
				Expect(rm.updated).To(Equal(1))

				err := universal.UpdateOutbounds(context.Background(), rm, vips)
				Expect(err).ToNot(HaveOccurred())

				Expect(rm.updated).To(Equal(1))
			})

			It("should delete outbounds when service is deleted", func() {
				// when
				err := rm.Delete(context.Background(), &mesh.DataplaneResource{}, store.DeleteByKey("dp-2", "default"))
				Expect(err).ToNot(HaveOccurred())
				// and
				err = universal.UpdateOutbounds(context.Background(), rm, vips)
				Expect(err).ToNot(HaveOccurred())

				// then
				dp1 := &mesh.DataplaneResource{}
				err = rm.Get(context.Background(), dp1, store.GetByKey("dp-1", "default"))
				Expect(err).ToNot(HaveOccurred())
				Expect(dp1.Spec.Networking.Outbound).To(HaveLen(0))
			})
		})
	})

	Context("transparentProxying disabled", func() {
		BeforeEach(func() {
			err := rm.Create(context.Background(), &mesh.DataplaneResource{
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port: 8080,
								Tags: map[string]string{
									"kuma.io/service": "service-1",
								},
							},
						},
						Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
							{
								Address: "127.0.0.1",
								Port:    81,
								Tags: map[string]string{
									"kuma.io/service": "service-2",
								},
							},
						},
					},
				},
			}, store.CreateByKey("dp-1", "default"))
			Expect(err).ToNot(HaveOccurred())

			err = rm.Create(context.Background(), &mesh.DataplaneResource{
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port: 8081,
								Tags: map[string]string{
									"kuma.io/service": "service-2",
								},
							},
						},
						Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
							{
								Address: "127.0.0.1",
								Port:    80,
								Tags: map[string]string{
									"kuma.io/service": "service-1",
								},
							},
						},
					},
				},
			}, store.CreateByKey("dp-2", "default"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not update outbounds", func() {
			// when
			err := universal.UpdateOutbounds(context.Background(), rm, vips)
			Expect(err).ToNot(HaveOccurred())

			// then
			dp1 := &mesh.DataplaneResource{}
			err = rm.Get(context.Background(), dp1, store.GetByKey("dp-1", "default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(dp1.Spec.Networking.Outbound).To(HaveLen(1))
			Expect(dp1.Spec.Networking.Outbound[0].Address).To(Equal("127.0.0.1"))
			Expect(dp1.Spec.Networking.Outbound[0].Port).To(Equal(uint32(81)))

			// and then
			dp2 := &mesh.DataplaneResource{}
			err = rm.Get(context.Background(), dp2, store.GetByKey("dp-2", "default"))
			Expect(err).ToNot(HaveOccurred())
			Expect(dp2.Spec.Networking.Outbound).To(HaveLen(1))
			Expect(dp2.Spec.Networking.Outbound[0].Address).To(Equal("127.0.0.1"))
			Expect(dp2.Spec.Networking.Outbound[0].Port).To(Equal(uint32(80)))
		})
	})
})
