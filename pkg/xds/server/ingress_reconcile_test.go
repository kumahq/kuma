package server_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
	"github.com/Kong/kuma/pkg/xds/server"
)

var _ = Describe("Ingress Reconciler", func() {

	dataplane1 := mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: "1.1.1.2",
			Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
				{
					Port: 8080,
					Tags: map[string]string{
						"service": "payment",
					},
				},
			},
		},
	}
	dataplane2 := mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: "1.1.1.2",
			Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
				{
					Port: 8080,
					Tags: map[string]string{
						"service": "payment",
					},
				},
			},
		},
	}
	dataplane3 := mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: "1.1.1.4",
			Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
				{
					Port: 8080,
					Tags: map[string]string{
						"service": "payment",
						"version": "0.1",
					},
				},
			},
		},
	}
	dataplane4 := mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: "1.1.1.1",
			Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
				{
					Port: 8080,
					Tags: map[string]string{
						"service": "web",
						"region":  "eu",
						"version": "0.1",
					},
				},
				{
					Port: 8081,
					Tags: map[string]string{
						"service": "web-api",
						"region":  "eu",
						"version": "0.8",
					},
				},
			},
		},
	}
	ingress := mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Ingress: &mesh_proto.Dataplane_Networking_Ingress{},
			Inbound: nil,
		},
	}

	Context("initially", func() {

		manager := core_manager.NewResourceManager(memory.NewStore())
		reconciler := server.NewIngressReconciler(manager)

		BeforeEach(func() {
			err := manager.Create(context.Background(), &core_mesh.MeshResource{}, store.CreateByKey("demo", "demo"))
			Expect(err).ToNot(HaveOccurred())

			err = manager.Create(context.Background(), &core_mesh.DataplaneResource{Spec: ingress}, store.CreateByKey("dp-ingress", "demo"))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when new dataplanes added", func() {

			BeforeEach(func() {
				err := manager.Create(context.Background(), &core_mesh.DataplaneResource{Spec: dataplane1}, store.CreateByKey("dp1", "demo"))
				Expect(err).ToNot(HaveOccurred())

				err = manager.Create(context.Background(), &core_mesh.DataplaneResource{Spec: dataplane2}, store.CreateByKey("dp2", "demo"))
				Expect(err).ToNot(HaveOccurred())

				err = manager.Create(context.Background(), &core_mesh.DataplaneResource{Spec: dataplane3}, store.CreateByKey("dp3", "demo"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("should create inbound in ingress for every unique tag set", func() {

				err := reconciler.Reconcile(model.ResourceKey{Name: "dp-ingress", Mesh: "demo"})
				Expect(err).ToNot(HaveOccurred())

				expected := `
                    networking:
                      inbound:
                        - port: 10001
                          tags:
                            service: payment
                        - port: 10002
                          tags:
                            service: payment
                            version: "0.1"
                      ingress: {}
`

				var actual core_mesh.DataplaneResource
				err = manager.Get(context.Background(), &actual, store.GetByKey("dp-ingress", "demo"))
				Expect(err).ToNot(HaveOccurred())

				actualYAML, err := util_proto.ToYAML(&actual.Spec)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualYAML).To(MatchYAML(expected))
			})
		})
	})

	Context("has existing dataplanes", func() {

		manager := core_manager.NewResourceManager(memory.NewStore())
		reconciler := server.NewIngressReconciler(manager)

		BeforeEach(func() {
			err := manager.Create(context.Background(), &core_mesh.MeshResource{}, store.CreateByKey("demo", "demo"))
			Expect(err).ToNot(HaveOccurred())

			err = manager.Create(context.Background(), &core_mesh.DataplaneResource{Spec: ingress}, store.CreateByKey("dp-ingress", "demo"))
			Expect(err).ToNot(HaveOccurred())

			err = manager.Create(context.Background(), &core_mesh.DataplaneResource{Spec: dataplane1}, store.CreateByKey("dp1", "demo"))
			Expect(err).ToNot(HaveOccurred())

			err = manager.Create(context.Background(), &core_mesh.DataplaneResource{Spec: dataplane2}, store.CreateByKey("dp2", "demo"))
			Expect(err).ToNot(HaveOccurred())

			err = manager.Create(context.Background(), &core_mesh.DataplaneResource{Spec: dataplane3}, store.CreateByKey("dp3", "demo"))
			Expect(err).ToNot(HaveOccurred())

			err = reconciler.Reconcile(model.ResourceKey{Name: "dp-ingress", Mesh: "demo"})
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when new dataplanes added", func() {

			BeforeEach(func() {
				err := manager.Create(context.Background(), &core_mesh.DataplaneResource{Spec: dataplane4}, store.CreateByKey("dp4", "demo"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("should leave the same ports for existing dataplanes", func() {

				err := reconciler.Reconcile(model.ResourceKey{Name: "dp-ingress", Mesh: "demo"})
				Expect(err).ToNot(HaveOccurred())

				expected := `
                    networking:
                      inbound:
                      - port: 10001
                        tags:
                          service: payment
                      - port: 10002
                        tags:
                          service: payment
                          version: "0.1"
                      - port: 10003
                        tags:
                          region: eu
                          service: web
                          version: "0.1"
                      - port: 10004
                        tags:
                          region: eu
                          service: web-api
                          version: "0.8"
                      ingress: {}
`
				var actual core_mesh.DataplaneResource
				err = manager.Get(context.Background(), &actual, store.GetByKey("dp-ingress", "demo"))
				Expect(err).ToNot(HaveOccurred())

				actualYAML, err := util_proto.ToYAML(&actual.Spec)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualYAML).To(MatchYAML(expected))
			})
		})
	})
})
