package dns_test

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	resources_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/dns"
	"github.com/kumahq/kuma/pkg/dns/resolver"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DNS sync", func() {

	var resManager resources_manager.ResourceManager
	var dnsResolver resolver.DNSResolver
	var dnsResolverFollower resolver.DNSResolver
	var stop chan struct{}

	BeforeEach(func() {
		stop = make(chan struct{})
		memory := memory_resources.NewStore()
		resManager = resources_manager.NewResourceManager(memory)
		cfgManager := config_manager.NewConfigManager(memory)
		dnsResolver = resolver.NewDNSResolver("mesh")

		vipAllocator, err := dns.NewVIPsAllocator(resManager, cfgManager, "240.0.0.0/24", dnsResolver)
		Expect(err).ToNot(HaveOccurred())
		go func() {
			Expect(vipAllocator.Start(stop)).ToNot(HaveOccurred())
		}()

		dnsResolverFollower = resolver.NewDNSResolver("mesh")
		vipsSynchronizer := dns.NewVIPsSynchronizer(dnsResolverFollower, resManager, cfgManager, neverLeaderInfo{})
		go func() {
			Expect(vipsSynchronizer.Start(stop)).ToNot(HaveOccurred())
		}()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		close(stop)
	})

	Describe("should allocate VIPs and synchronize to DNS Resolver", func() {
		BeforeEach(func() {
			// given a mesh and one service
			err := resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(model.DefaultMesh, model.NoMesh))
			Expect(err).ToNot(HaveOccurred())

			webDp := core_mesh.DataplaneResource{
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.1",
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port: 1234,
								Tags: map[string]string{
									"kuma.io/service": "web",
								},
							},
						},
					},
				},
			}
			err = resManager.Create(context.Background(), &webDp, core_store.CreateByKey("dp-1", "default"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should sync web to DNS resolver and to the follower", func() {
			// then service "web" is synchronized to DNS Resolver
			Eventually(func() error {
				_, err := dnsResolver.ForwardLookup("web")
				return err
			}, "5s").ShouldNot(HaveOccurred())
			ip, _ := dnsResolver.ForwardLookup("web")
			Expect(ip).Should(HavePrefix("240.0.0"))

			// and replicated to a follower
			Eventually(func() error {
				_, err := dnsResolverFollower.ForwardLookup("web")
				return err
			}, "5s").ShouldNot(HaveOccurred())
			ip2, _ := dnsResolverFollower.ForwardLookup("web")
			Expect(ip).To(Equal(ip2))
		})

		It("should sync another service", func() {
			// when "backend" service is up
			backendDp := core_mesh.DataplaneResource{
				Spec: &mesh_proto.Dataplane{
					Networking: &mesh_proto.Dataplane_Networking{
						Address: "192.168.0.1",
						Ingress: &mesh_proto.Dataplane_Networking_Ingress{
							AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{
								{
									Mesh: "default",
									Tags: map[string]string{
										"kuma.io/service": "backend",
									},
								},
							},
						},
						Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
							{
								Port: 1234,
								Tags: map[string]string{
									mesh_proto.ZoneTag: "zone-2",
								},
							},
						},
					},
				},
			}
			err := resManager.Create(context.Background(), &backendDp, core_store.CreateByKey("dp-2", "default"))
			Expect(err).ToNot(HaveOccurred())

			// then service "backend" is synchronized to DNS Resolver
			Eventually(func() error {
				_, err := dnsResolver.ForwardLookup("backend")
				return err
			}, "5s").ShouldNot(HaveOccurred())
			ip, _ := dnsResolver.ForwardLookup("backend")
			Expect(ip).Should(HavePrefix("240.0.0"))

			// and replicated to a follower
			Eventually(func() error {
				_, err := dnsResolverFollower.ForwardLookup("backend")
				return err
			}, "5s").ShouldNot(HaveOccurred())
			ip2, _ := dnsResolverFollower.ForwardLookup("backend")
			Expect(ip).To(Equal(ip2))
		})

		It("should remove web from DNS resolver when service is deleted", func() {
			// when service "web" is deleted
			err := resManager.Delete(context.Background(), core_mesh.NewDataplaneResource(), core_store.DeleteByKey("dp-1", "default"))
			Expect(err).ToNot(HaveOccurred())

			// then service "web" is removed from DNS Resolver
			Eventually(func() error {
				_, err := dnsResolver.ForwardLookup("web")
				return err
			}, "5s").Should(MatchError("service [web] not found in domain [mesh]."))

			// and replicated to a follower
			Eventually(func() error {
				_, err := dnsResolverFollower.ForwardLookup("web")
				return err
			}, "5s").Should(MatchError("service [web] not found in domain [mesh]."))
		})
	})
})

type neverLeaderInfo struct {
}

func (n neverLeaderInfo) IsLeader() bool {
	return false
}

var _ component.LeaderInfo = &neverLeaderInfo{}
