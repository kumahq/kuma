package dns_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns"
	"github.com/kumahq/kuma/pkg/dns/resolver"
	"github.com/kumahq/kuma/pkg/dns/vips"
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
		Expect(vipList.HostnameEntries()).To(HaveLen(2))

		vipList, err = persistence.GetByMesh("mesh-2")
		Expect(err).ToNot(HaveOccurred())

		for _, service := range []string{"backend.mesh", "frontend.mesh", "web.mesh"} {
			ip, err := r.ForwardLookupFQDN(service)
			Expect(err).ToNot(HaveOccurred())
			Expect(ip).To(HavePrefix("240.0.0"))
		}

		Expect(vipList.HostnameEntries()).To(HaveLen(1))
	})

	It("should respect already allocated VIPs in case of IPAM restarts", func() {
		// setup
		persistence := vips.NewPersistence(rm, cm)
		// we add VIPs directly to the 'persistence' object
		// that emulates situation when IPAM is fresh and doesn't aware of allocated VIPs
		err := persistence.Set("mesh-1", vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
			vips.NewServiceEntry("frontend"): {Address: "240.0.0.0"},
			vips.NewServiceEntry("backend"):  {Address: "240.0.0.1"},
		}))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), &mesh.DataplaneResource{Spec: dp("database")}, store.CreateByKey("dp-3", "mesh-1"))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = allocator.CreateOrUpdateVIPConfig("mesh-1")
		Expect(err).ToNot(HaveOccurred())

		vipList, err := persistence.GetByMesh("mesh-1")
		Expect(err).ToNot(HaveOccurred())
		// then
		expected := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
			vips.NewServiceEntry("backend"):  {Address: "240.0.0.1"},
			vips.NewServiceEntry("database"): {Address: "240.0.0.2"},
			vips.NewServiceEntry("frontend"): {Address: "240.0.0.0"},
		})
		Expect(vipList.HostnameEntries()).To(Equal(expected.HostnameEntries()))
		for _, k := range vipList.HostnameEntries() {
			Expect(vipList.Get(k).Address).To(Equal(expected.Get(k).Address))
		}
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

var _ = Describe("BuildVirtualOutboundMeshView", func() {
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
		serviceSet, err := dns.BuildVirtualOutboundMeshView(rm, "mesh-1")
		Expect(err).ToNot(HaveOccurred())

		// then
		expected := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{})
		Expect(expected.Add(vips.NewServiceEntry("backend"), vips.OutboundEntry{Port: 0, TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}, Origin: "service"})).ToNot(HaveOccurred())
		Expect(expected.Add(vips.NewServiceEntry("database"), vips.OutboundEntry{Port: 0, TagSet: map[string]string{mesh_proto.ServiceTag: "database"}, Origin: "service"})).ToNot(HaveOccurred())
		Expect(expected.Add(vips.NewServiceEntry("metrics"), vips.OutboundEntry{Port: 0, TagSet: map[string]string{mesh_proto.ServiceTag: "metrics"}, Origin: "service"})).ToNot(HaveOccurred())
		Expect(expected.Add(vips.NewServiceEntry("frontend"), vips.OutboundEntry{Port: 0, TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}, Origin: "service"})).ToNot(HaveOccurred())
		Expect(expected.Add(vips.NewServiceEntry("ingress-svc"), vips.OutboundEntry{Port: 0, TagSet: map[string]string{mesh_proto.ServiceTag: "ingress-svc"}, Origin: "service"})).ToNot(HaveOccurred())
		Expect(expected.Add(vips.NewServiceEntry("es-backend"), vips.OutboundEntry{Port: 8080, TagSet: map[string]string{mesh_proto.ServiceTag: "es-backend"}, Origin: "service"})).ToNot(HaveOccurred())
		Expect(expected.Add(vips.NewHostEntry("external.service.com"), vips.OutboundEntry{Port: 8080, TagSet: map[string]string{mesh_proto.ServiceTag: "es-backend"}, Origin: "host"})).ToNot(HaveOccurred())

		Expect(serviceSet.HostnameEntries()).To(Equal(expected.HostnameEntries()))
		for _, k := range serviceSet.HostnameEntries() {
			Expect(serviceSet.Get(k)).To(Equal(expected.Get(k)))
		}
	})
})

var _ = Describe("AllocateVIPs", func() {
	It("should allocate new VIPs", func() {
		// setup
		gv, err := vips.NewGlobalView("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		serviceSet := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{})
		Expect(serviceSet.Add(vips.NewServiceEntry("backend"), vips.OutboundEntry{TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}})).ToNot(HaveOccurred())
		Expect(serviceSet.Add(vips.NewServiceEntry("frontend"), vips.OutboundEntry{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}})).ToNot(HaveOccurred())
		// when
		err = dns.AllocateVIPs(gv, serviceSet)
		Expect(err).ToNot(HaveOccurred())
		// then
		Expect(serviceSet.HostnameEntries()).To(Equal([]vips.HostnameEntry{vips.NewServiceEntry("backend"), vips.NewServiceEntry("frontend")}))
		Expect(serviceSet.Get(vips.NewServiceEntry("backend")).Address).ToNot(BeEmpty())
		Expect(serviceSet.Get(vips.NewServiceEntry("frontend")).Address).ToNot(BeEmpty())
	})

	It("should generate the same VIP for services across meshes", func() {
		// setup
		gv, err := vips.NewGlobalView("240.0.0.0/4")
		Expect(err).ToNot(HaveOccurred())
		Expect(gv.Reserve(vips.NewServiceEntry("backend"), "240.0.0.0")).ToNot(HaveOccurred())
		Expect(gv.Reserve(vips.NewServiceEntry("frontend"), "240.0.0.1")).ToNot(HaveOccurred())
		Expect(gv.Reserve(vips.NewServiceEntry("database"), "240.0.0.10")).ToNot(HaveOccurred())
		Expect(err).ToNot(HaveOccurred())
		serviceSet := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{})
		Expect(serviceSet.Add(vips.NewServiceEntry("backend"), vips.OutboundEntry{Origin: "default", TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}})).ToNot(HaveOccurred())
		Expect(serviceSet.Add(vips.NewServiceEntry("frontend"), vips.OutboundEntry{Origin: "default", TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}})).ToNot(HaveOccurred())
		Expect(serviceSet.Add(vips.NewServiceEntry("database"), vips.OutboundEntry{Origin: "default", TagSet: map[string]string{mesh_proto.ServiceTag: "database"}})).ToNot(HaveOccurred())
		// when
		err = dns.AllocateVIPs(gv, serviceSet)
		Expect(err).ToNot(HaveOccurred())
		// then
		Expect(serviceSet.HostnameEntries()).To(Equal([]vips.HostnameEntry{vips.NewServiceEntry("backend"), vips.NewServiceEntry("database"), vips.NewServiceEntry("frontend")}))
		Expect(serviceSet.Get(vips.NewServiceEntry("backend")).Address).To(Equal("240.0.0.0"))
		Expect(serviceSet.Get(vips.NewServiceEntry("frontend")).Address).To(Equal("240.0.0.1"))
		Expect(serviceSet.Get(vips.NewServiceEntry("database")).Address).To(Equal("240.0.0.10"))
	})
})
