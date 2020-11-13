package dns_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("VIP Allocator", func() {
	Context("BuildServiceSet", func() {
		var rm manager.ResourceManager

		BeforeEach(func() {
			rm = manager.NewResourceManager(memory.NewStore())
		})

		It("should build service set for mesh", func() {
			// setup meshes
			err := rm.Create(context.Background(), &mesh.MeshResource{}, store.CreateByKey("mesh-1", model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			err = rm.Create(context.Background(), &mesh.MeshResource{}, store.CreateByKey("mesh-2", model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			dp := func(services ...string) *mesh_proto.Dataplane {
				inbound := []*mesh_proto.Dataplane_Networking_Inbound{}
				for _, service := range services {
					inbound = append(inbound, &mesh_proto.Dataplane_Networking_Inbound{
						Port: 8080,
						Tags: map[string]string{
							mesh_proto.ServiceTag: service,
						},
					})
				}
				return &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "127.0.0.1",
						Inbound: inbound,
					},
				}
			}

			// setup dataplanes
			err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: *dp("backend")}, store.CreateByKey("backend-1", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())

			err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: *dp("frontend")}, store.CreateByKey("frontend-1", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())

			err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: *dp("frontend")}, store.CreateByKey("frontend-2", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())

			err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: *dp("database", "metrics")}, store.CreateByKey("db-m-1", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())

			err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: *dp("another-mesh-svc")}, store.CreateByKey("another-mesh-dp-1", "mesh-2"))
			Expect(err).ToNot(HaveOccurred())

			// setup ingress
			err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 10001,
						},
					},
					Ingress: &mesh_proto.Dataplane_Networking_Ingress{
						AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
							{
								Mesh:      "mesh-1",
								Instances: 2,
								Tags: map[string]string{
									mesh_proto.ServiceTag: "ingress-svc",
								},
							},
							{
								Mesh:      "mesh-2",
								Instances: 3,
								Tags: map[string]string{
									mesh_proto.ServiceTag: "another-mesh-ingress-svc",
								},
							},
						},
					},
				},
			}}, store.CreateByKey("ingress-1", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())

			// setup external services
			es := func(service string) *mesh_proto.ExternalService {
				return &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "external.service.com:8080",
					},
					Tags: map[string]string{
						mesh_proto.ServiceTag: service,
					},
				}
			}

			err = rm.Create(context.Background(), &mesh.ExternalServiceResource{Spec: *es("es-backend")}, store.CreateByKey("es-backend-1", "mesh-1"))
			Expect(err).ToNot(HaveOccurred())

			err = rm.Create(context.Background(), &mesh.ExternalServiceResource{Spec: *es("another-mesh-es")}, store.CreateByKey("es-backend-1", "mesh-2"))
			Expect(err).ToNot(HaveOccurred())

			// when
			serviceSet, err := dns.BuildServiceSet(rm, "mesh-1")
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(serviceSet).To(Equal(dns.ServiceSet{
				"backend":     true,
				"frontend":    true,
				"database":    true,
				"metrics":     true,
				"ingress-svc": true,
				"es-backend":  true,
			}))
		})
	})

	Context("UpdateMeshedVIPs", func() {
		It("should allocate new VIPs", func() {
			// setup
			vips := dns.VIPList{}
			ipam, err := dns.NewSimpleIPAM("240.0.0.0/4")
			Expect(err).ToNot(HaveOccurred())
			serviceSet := dns.ServiceSet{
				"backend":  true,
				"frontend": true,
			}
			// when
			updated, err := dns.UpdateMeshedVIPs(vips, vips, ipam, serviceSet)
			Expect(err).ToNot(HaveOccurred())
			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(updated).To(BeTrue())
			Expect(vips).To(Equal(dns.VIPList{
				"backend":  "240.0.0.0",
				"frontend": "240.0.0.1",
			}))
		})

		It("should free IP for deleted service", func() {
			// setup
			vips := dns.VIPList{
				"backend":  "240.0.0.0",
				"frontend": "240.0.0.1",
			}
			ipam, err := dns.NewSimpleIPAM("240.0.0.0/4")
			Expect(err).ToNot(HaveOccurred())
			serviceSet := dns.ServiceSet{
				"backend": true,
			}
			// when
			updated, err := dns.UpdateMeshedVIPs(vips, vips, ipam, serviceSet)
			Expect(err).ToNot(HaveOccurred())
			// then
			Expect(updated).To(BeTrue())
			Expect(vips).To(Equal(dns.VIPList{
				"backend": "240.0.0.0",
			}))
		})

		It("should return updated=false if nothing changed", func() {
			// setup
			vips := dns.VIPList{
				"backend":  "240.0.0.0",
				"frontend": "240.0.0.1",
			}
			ipam, err := dns.NewSimpleIPAM("240.0.0.0/4")
			Expect(err).ToNot(HaveOccurred())
			serviceSet := dns.ServiceSet{
				"backend":  true,
				"frontend": true,
			}
			// when
			updated, err := dns.UpdateMeshedVIPs(vips, vips, ipam, serviceSet)
			Expect(err).ToNot(HaveOccurred())
			// then
			Expect(updated).To(BeFalse())
			Expect(vips).To(Equal(dns.VIPList{
				"backend":  "240.0.0.0",
				"frontend": "240.0.0.1",
			}))
		})

		It("should generate the same VIP for services across meshes", func() {
			// setup
			global := dns.VIPList{
				"backend":  "240.0.0.0",
				"frontend": "240.0.0.1",
				"database": "240.0.0.10",
			}
			meshed := dns.VIPList{
				"backend":  "240.0.0.0",
				"frontend": "240.0.0.1",
			}
			ipam, err := dns.NewSimpleIPAM("240.0.0.0/4")
			Expect(err).ToNot(HaveOccurred())
			serviceSet := dns.ServiceSet{
				"backend":  true,
				"frontend": true,
				"database": true,
			}
			// when
			updated, err := dns.UpdateMeshedVIPs(global, meshed, ipam, serviceSet)
			Expect(err).ToNot(HaveOccurred())
			// then
			Expect(updated).To(BeTrue())
			Expect(meshed).To(Equal(dns.VIPList{
				"backend":  "240.0.0.0",
				"frontend": "240.0.0.1",
				"database": "240.0.0.10",
			}))
		})
	})
})
