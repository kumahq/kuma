package resolver_test

import (
	"fmt"
	"strconv"

	config_manager "github.com/Kong/kuma/pkg/core/config/manager"
	config_store "github.com/Kong/kuma/pkg/core/config/store"
	resources_memory "github.com/Kong/kuma/pkg/plugins/resources/memory"

	"github.com/miekg/dns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/dns-server/resolver"
	"github.com/Kong/kuma/pkg/test"
)

var _ = Describe("DNS server", func() {

	var configm config_manager.ConfigManager

	BeforeEach(func() {
		store := config_store.NewConfigStore(resources_memory.NewStore())
		configm = config_manager.NewConfigManager(store)
	})

	Describe("Network Operation", func() {

		var ip, port string
		stop := make(chan struct{})
		done := make(chan struct{})
		BeforeEach(func() {
			// setup
			p, err := test.GetFreePort()
			Expect(err).ToNot(HaveOccurred())
			port = strconv.Itoa(p)

			resolver, err := NewSimpleDNSResolver("mesh", "127.0.0.1", port, "240.0.0.0/4", configm)
			Expect(err).ToNot(HaveOccurred())
			resolver.SetElectedLeader(true)

			// given
			_, err = resolver.AddService("service")
			Expect(err).ToNot(HaveOccurred())

			ip, err = resolver.ForwardLookupFQDN("service.mesh")
			Expect(err).ToNot(HaveOccurred())

			go func() {
				err := resolver.Start(stop)
				Expect(err).ToNot(HaveOccurred())
				done <- struct{}{}
			}()
		})

		AfterEach(func() {
			// stop the resolver
			stop <- struct{}{}
			<-done
		})

		It("Should resolve", func() {
			// when
			client := new(dns.Client)
			message := new(dns.Msg)
			_ = message.SetQuestion("service.mesh.", dns.TypeA)
			var response *dns.Msg
			var err error
			Eventually(func() error {
				response, _, err = client.Exchange(message, "127.0.0.1:"+port)
				return err
			}).ShouldNot(HaveOccurred())
			// then
			Expect(response.Answer[0].String()).To(Equal(fmt.Sprintf("service.mesh.\t60\tIN\tA\t%s", ip)))
		})

		It("Should resolve concurrent", func() {
			resolved := make(chan struct{})
			for i := 0; i < 100; i++ {
				go func() {
					// when
					client := new(dns.Client)
					message := new(dns.Msg)
					_ = message.SetQuestion("service.mesh.", dns.TypeA)
					var response *dns.Msg
					var err error
					Eventually(func() error {
						response, _, err = client.Exchange(message, "127.0.0.1:"+port)
						return err
					}).ShouldNot(HaveOccurred())
					// then
					Expect(response.Answer[0].String()).To(Equal(fmt.Sprintf("service.mesh.\t60\tIN\tA\t%s", ip)))
					resolved <- struct{}{}
				}()
			}

			for i := 0; i < 100; i++ {
				<-resolved
			}
		})

		It("Should not resolve", func() {
			// when
			client := new(dns.Client)
			message := new(dns.Msg)
			_ = message.SetQuestion("backend.mesh.", dns.TypeA)
			var response *dns.Msg
			var err error
			Eventually(func() error {
				response, _, err = client.Exchange(message, "127.0.0.1:"+port)
				return err
			}).ShouldNot(HaveOccurred())
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(len(response.Answer)).To(Equal(0))
		})
	})

	It("DNS Server basic functionality", func() {
		// setup
		resolver, err := NewSimpleDNSResolver("mesh", "127.0.0.1", "5653", "240.0.0.0/4", configm)
		Expect(err).ToNot(HaveOccurred())
		resolver.SetElectedLeader(true)

		// given
		_, err = resolver.AddService("service")
		Expect(err).ToNot(HaveOccurred())

		ipService, err := resolver.ForwardLookup("service")
		Expect(err).ToNot(HaveOccurred())

		ipFQDN, err := resolver.ForwardLookupFQDN("service.mesh")
		Expect(err).ToNot(HaveOccurred())

		// when
		service, err := resolver.ReverseLookup(ipFQDN)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(service).To(Equal("service.mesh"))
		// and
		Expect(ipService).To(Equal(ipFQDN))
	})

	It("DNS Server service operation", func() {
		// given
		resolver, err := NewSimpleDNSResolver("mesh", "127.0.0.1", "5653", "240.0.0.0/4", configm)
		Expect(err).ToNot(HaveOccurred())
		resolver.SetElectedLeader(true)

		// when
		_, err = resolver.AddService("service")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.AddService("backend")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = resolver.RemoveService("service")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = resolver.RemoveService("backend")
		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("DNS Server sync operation", func() {
		// setup
		resolver, err := NewSimpleDNSResolver("mesh", "127.0.0.1", "5653", "240.0.0.0/4", configm)
		Expect(err).ToNot(HaveOccurred())
		resolver.SetElectedLeader(true)

		services := map[string]bool{
			"example-one.kuma-test.svc:80":   true,
			"example-two.kuma-test.svc:80":   true,
			"example-three.kuma-test.svc:80": true,
			"example-four.kuma-test.svc:80":  true,
			"example-five.kuma-test.svc:80":  true,
		}

		// given
		err = resolver.SyncServices(services)
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookupFQDN("example-one.mesh")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookupFQDN("example-five.mesh")
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookupFQDN("example-five.other")
		// then
		Expect(err).To(HaveOccurred())

		// given
		delete(services, "example-five.kuma-test.svc:80")

		// when
		err = resolver.SyncServices(services)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookupFQDN("example-five.mesh")
		// then
		Expect(err).To(HaveOccurred())

		// when
		err = resolver.SyncServices(map[string]bool{})
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = resolver.ForwardLookupFQDN("example-five.mesh")
		// then
		Expect(err).To(HaveOccurred())
	})

	It("should sync leader and follower", func() {
		// setup
		leader, err := NewSimpleDNSResolver("mesh", "127.0.0.1", "5653", "240.0.0.0/4", configm)
		Expect(err).ToNot(HaveOccurred())
		leader.SetElectedLeader(true)

		follower, err := NewSimpleDNSResolver("mesh", "127.0.0.1", "15653", "240.0.0.0/4", configm)
		Expect(err).ToNot(HaveOccurred())

		services := map[string]bool{
			"example-one.kuma-test.svc:80":   true,
			"example-two.kuma-test.svc:80":   true,
			"example-three.kuma-test.svc:80": true,
			"example-four.kuma-test.svc:80":  true,
			"example-five.kuma-test.svc:80":  true,
		}

		// given
		err = leader.SyncServices(services)
		Expect(err).ToNot(HaveOccurred())

		err = follower.SyncServices(services)
		Expect(err).ToNot(HaveOccurred())

		// when
		ip1, err := leader.ForwardLookupFQDN("example-one.mesh")
		Expect(err).ToNot(HaveOccurred())

		// then
		ip2, err := follower.ForwardLookupFQDN("example-one.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(ip1).To(Equal(ip2))

		// given
		delete(services, "example-five.kuma-test.svc:80")

		// when
		err = leader.SyncServices(services)
		// then
		Expect(err).ToNot(HaveOccurred())

		err = follower.SyncServices(services)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		_, err = leader.ForwardLookupFQDN("example-five.mesh")
		// then
		Expect(err).To(HaveOccurred())

		// when
		_, err = follower.ForwardLookupFQDN("example-five.mesh")
		// then
		Expect(err).To(HaveOccurred())

	})
})
