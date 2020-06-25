package dns_test

import (
	"context"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	config_manager "github.com/Kong/kuma/pkg/core/config/manager"
	config_store "github.com/Kong/kuma/pkg/core/config/store"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	resources_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/dns"
	memory_resources "github.com/Kong/kuma/pkg/plugins/resources/memory"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("", func() {

	var resManager resources_manager.ResourceManager
	var dnsResolver dns.DNSResolver
	var stop chan struct{}

	BeforeEach(func() {
		stop = make(chan struct{})
		memory := memory_resources.NewStore()
		resManager = resources_manager.NewResourceManager(memory)
		cfgManager := config_manager.NewConfigManager(config_store.NewConfigStore(memory))
		persistence := dns.NewDNSPersistence(cfgManager)

		ipam, err := dns.NewSimpleIPAM("240.0.0.0/24")
		Expect(err).ToNot(HaveOccurred())
		vipAllocator, err := dns.NewVIPsAllocator(resManager, persistence, ipam)
		Expect(err).ToNot(HaveOccurred())
		go func() {
			Expect(vipAllocator.Start(stop)).ToNot(HaveOccurred())
		}()

		dnsResolver = dns.NewDNSResolver("mesh")
		vipsSynchronizer, err := dns.NewVIPsSynchronizer(resManager, dnsResolver, persistence)
		Expect(err).ToNot(HaveOccurred())
		go func() {
			Expect(vipsSynchronizer.Start(stop)).ToNot(HaveOccurred())
		}()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		close(stop)
	})

	It("", func() {
		// given
		mesh := core_mesh.MeshResource{}
		err := resManager.Create(context.Background(), &mesh, core_store.CreateByKey("default", "default"))
		Expect(err).ToNot(HaveOccurred())

		// when "web" service is up
		webDp := core_mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 1234,
							Tags: map[string]string{
								"service": "web",
							},
						},
					},
				},
			},
		}
		err = resManager.Create(context.Background(), &webDp, core_store.CreateByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then service "web" is synchronized to DNS Resolver
		Eventually(func() error {
			_, err := dnsResolver.ForwardLookup("web")
			return err
		}, "5s").ShouldNot(HaveOccurred())
		ip, _ := dnsResolver.ForwardLookup("web")
		Expect(ip).Should(HavePrefix("240.0.0"))

		// when "backend" service is up
		backendDp := core_mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port: 1234,
							Tags: map[string]string{
								"service": "backend",
							},
						},
					},
				},
			},
		}
		err = resManager.Create(context.Background(), &backendDp, core_store.CreateByKey("dp-2", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then service "backend" is synchronized to DNS Resolver
		Eventually(func() error {
			_, err := dnsResolver.ForwardLookup("backend")
			return err
		}, "5s").ShouldNot(HaveOccurred())
		ip, _ = dnsResolver.ForwardLookup("web")
		Expect(ip).Should(HavePrefix("240.0.0"))

		// when service "web" is deleted
		err = resManager.Delete(context.Background(), &core_mesh.DataplaneResource{}, core_store.DeleteByKey("dp-1", "default"))
		Expect(err).ToNot(HaveOccurred())

		// then service "web" is removed from DNS Resolver
		Eventually(func() error {
			_, err := dnsResolver.ForwardLookup("web")
			return err
		}, "5s").Should(MatchError("service [web] not found in domain [mesh]."))
	})

})
