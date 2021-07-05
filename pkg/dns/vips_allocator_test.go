package dns_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"

	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/dns/resolver"

	"github.com/kumahq/kuma/pkg/dns"
	"github.com/kumahq/kuma/pkg/dns/vips"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

func dp(services ...string) *mesh_proto.Dataplane {
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

type errConfigManager struct {
	config_manager.ConfigManager
}

func (e *errConfigManager) Update(ctx context.Context, r *config_model.ConfigResource, opts ...store.UpdateOptionsFunc) error {
	meshName, _ := vips.MeshFromConfigKey(r.GetMeta().GetName())
	return errors.Errorf("error during update, mesh = %s", meshName)
}

var _ = Describe("VIP Allocator", func() {
	var rm manager.ResourceManager
	var cm config_manager.ConfigManager
	var allocator *dns.VIPsAllocator
	var r resolver.DNSResolver

	BeforeEach(func() {
		s := memory.NewStore()
		rm = manager.NewResourceManager(s)
		cm = config_manager.NewConfigManager(s)
		r = resolver.NewDNSResolver("mesh")

		err := rm.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey("mesh-2", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("backend")}, store.CreateByKey("dp-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("frontend")}, store.CreateByKey("dp-2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("web")}, store.CreateByKey("dp-3", "mesh-2"))
		Expect(err).ToNot(HaveOccurred())

		allocator, err = dns.NewVIPsAllocator(rm, cm, "240.0.0.0/24", r)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create VIP config for each mesh", func() {
		// when
		err := allocator.CreateOrUpdateVIPConfigs()
		Expect(err).ToNot(HaveOccurred())

		persistence := vips.NewPersistence(rm, cm)

		// then
		vipList, err := persistence.GetByMesh("mesh-1")
		Expect(err).ToNot(HaveOccurred())
		Expect(vipList).To(HaveLen(2))

		vipList, err = persistence.GetByMesh("mesh-2")
		Expect(err).ToNot(HaveOccurred())

		for _, service := range []string{"backend.mesh", "frontend.mesh", "web.mesh"} {
			ip, err := r.ForwardLookupFQDN(service)
			Expect(err).ToNot(HaveOccurred())
			Expect(ip).To(HavePrefix("240.0.0"))
		}

		Expect(vipList).To(HaveLen(1))
	})

	It("should respect already allocated VIPs in case of IPAM restarts", func() {
		// setup
		persistence := vips.NewPersistence(rm, cm)
		// we add VIPs directly to the 'persistence' object
		// that emulates situation when IPAM is fresh and doesn't aware of allocated VIPs
		err := persistence.Set("mesh-1", vips.List{
			vips.NewServiceEntry("frontend"): "240.0.0.0",
			vips.NewServiceEntry("backend"):  "240.0.0.1",
		})
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("database")}, store.CreateByKey("dp-3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = allocator.CreateOrUpdateVIPConfig("mesh-1")
		Expect(err).ToNot(HaveOccurred())

		vipList, err := persistence.GetByMesh("mesh-1")
		Expect(err).ToNot(HaveOccurred())
		// then
		Expect(vipList).To(Equal(vips.List{
			vips.NewServiceEntry("frontend"): "240.0.0.0",
			vips.NewServiceEntry("backend"):  "240.0.0.1",
			vips.NewServiceEntry("database"): "240.0.0.2",
		}))
	})

	It("should return error if failed to update VIP config", func() {
		errConfigManager := &errConfigManager{ConfigManager: cm}
		errAllocator, err := dns.NewVIPsAllocator(rm, errConfigManager, "240.0.0.0/24", r)
		Expect(err).ToNot(HaveOccurred())

		err = errAllocator.CreateOrUpdateVIPConfig("mesh-1")
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("database")}, store.CreateByKey("dp-3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = errAllocator.CreateOrUpdateVIPConfig("mesh-1")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("error during update, mesh = mesh-1"))
	})

	It("should try to update all meshes and return combined error", func() {
		errConfigManager := &errConfigManager{ConfigManager: cm}
		errAllocator, err := dns.NewVIPsAllocator(rm, errConfigManager, "240.0.0.0/24", r)
		Expect(err).ToNot(HaveOccurred())

		err = errAllocator.CreateOrUpdateVIPConfigs()
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("database")}, store.CreateByKey("dp-3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("payment")}, store.CreateByKey("dp-4", "mesh-2"))
		Expect(err).ToNot(HaveOccurred())

		err = errAllocator.CreateOrUpdateVIPConfigs()
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError("error during update, mesh = mesh-1; error during update, mesh = mesh-2"))
	})
})

var _ = Describe("BuildServiceSet", func() {
	var rm manager.ResourceManager

	BeforeEach(func() {
		rm = manager.NewResourceManager(memory.NewStore())
	})

	It("should build service set for mesh", func() {
		// setup meshes
		err := rm.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey("mesh-1", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey("mesh-2", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), mesh.NewMeshResource(), store.CreateByKey("mesh-3", model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		// setup dataplanes
		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("backend")}, store.CreateByKey("backend-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("frontend")}, store.CreateByKey("frontend-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("frontend")}, store.CreateByKey("frontend-2", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("database", "metrics")}, store.CreateByKey("db-m-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("another-mesh-svc")}, store.CreateByKey("another-mesh-dp-1", "mesh-2"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("only-mesh-3-service")}, store.CreateByKey("dp-m-3", "mesh-3"))
		Expect(err).ToNot(HaveOccurred())

		// setup ingress
		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: &mesh_proto.Dataplane{
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

		err = rm.Create(context.Background(), &mesh.ExternalServiceResource{Spec: es("es-backend")}, store.CreateByKey("es-backend-1", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.ExternalServiceResource{Spec: es("another-mesh-es")}, store.CreateByKey("es-backend-1", "mesh-2"))
		Expect(err).ToNot(HaveOccurred())

		// when
		serviceSet, err := dns.BuildServiceSet(rm, "mesh-1")
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(serviceSet).To(Equal(vips.EntrySet{
			vips.NewServiceEntry("backend"):           true,
			vips.NewServiceEntry("frontend"):          true,
			vips.NewServiceEntry("database"):          true,
			vips.NewServiceEntry("metrics"):           true,
			vips.NewServiceEntry("ingress-svc"):       true,
			vips.NewServiceEntry("es-backend"):        true,
			vips.NewHostEntry("external.service.com"): true,
		}))
	})
})

var _ = Describe("UpdateMeshedVIPs", func() {
	It("should allocate new VIPs", func() {
		// setup
		vipsList := vips.List{}
		ipam, err := dns.NewSimpleIPAM("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		serviceSet := vips.EntrySet{
			vips.NewServiceEntry("backend"):  true,
			vips.NewServiceEntry("frontend"): true,
		}
		// when
		updated, err := dns.UpdateMeshedVIPs(vipsList, vipsList, ipam, serviceSet)
		Expect(err).ToNot(HaveOccurred())
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(updated).To(BeTrue())
		Expect(vipsList).To(Equal(vips.List{
			vips.NewServiceEntry("backend"):  "240.0.0.0",
			vips.NewServiceEntry("frontend"): "240.0.0.1",
		}))
	})

	It("should free IP for deleted service", func() {
		// setup
		vipsList := vips.List{
			vips.NewServiceEntry("backend"):  "240.0.0.0",
			vips.NewServiceEntry("frontend"): "240.0.0.1",
		}
		ipam, err := dns.NewSimpleIPAM("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		serviceSet := vips.EntrySet{
			vips.NewServiceEntry("backend"): true,
		}
		// when
		updated, err := dns.UpdateMeshedVIPs(vipsList, vipsList, ipam, serviceSet)
		Expect(err).ToNot(HaveOccurred())
		// then
		Expect(updated).To(BeTrue())
		Expect(vipsList).To(Equal(vips.List{
			vips.NewServiceEntry("backend"): "240.0.0.0",
		}))
	})

	It("should return updated=false if nothing changed", func() {
		// setup
		vipsList := vips.List{
			vips.NewServiceEntry("backend"):  "240.0.0.0",
			vips.NewServiceEntry("frontend"): "240.0.0.1",
		}
		ipam, err := dns.NewSimpleIPAM("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		serviceSet := vips.EntrySet{
			vips.NewServiceEntry("backend"):  true,
			vips.NewServiceEntry("frontend"): true,
		}
		// when
		updated, err := dns.UpdateMeshedVIPs(vipsList, vipsList, ipam, serviceSet)
		Expect(err).ToNot(HaveOccurred())
		// then
		Expect(updated).To(BeFalse())
		Expect(vipsList).To(Equal(vips.List{
			vips.NewServiceEntry("backend"):  "240.0.0.0",
			vips.NewServiceEntry("frontend"): "240.0.0.1",
		}))
	})

	It("should generate the same VIP for services across meshes", func() {
		// setup
		global := vips.List{
			vips.NewServiceEntry("backend"):  "240.0.0.0",
			vips.NewServiceEntry("frontend"): "240.0.0.1",
			vips.NewServiceEntry("database"): "240.0.0.10",
		}
		meshed := vips.List{
			vips.NewServiceEntry("backend"):  "240.0.0.0",
			vips.NewServiceEntry("frontend"): "240.0.0.1",
		}
		ipam, err := dns.NewSimpleIPAM("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		serviceSet := vips.EntrySet{
			vips.NewServiceEntry("backend"):  true,
			vips.NewServiceEntry("frontend"): true,
			vips.NewServiceEntry("database"): true,
		}
		// when
		updated, err := dns.UpdateMeshedVIPs(global, meshed, ipam, serviceSet)
		Expect(err).ToNot(HaveOccurred())
		// then
		Expect(updated).To(BeTrue())
		Expect(meshed).To(Equal(vips.List{
			vips.NewServiceEntry("backend"):  "240.0.0.0",
			vips.NewServiceEntry("frontend"): "240.0.0.1",
			vips.NewServiceEntry("database"): "240.0.0.10",
		}))
	})
})
