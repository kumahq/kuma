package dns_server_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	dns_server "github.com/kumahq/kuma/v2/pkg/config/dns-server"
)

var _ = Describe("Config", func() {
	type testCase struct {
		config dns_server.Config
		error  string
	}

	DescribeTable("should reject invalid config",
		func(given testCase) {
			err := given.config.Validate()
			Expect(err).To(MatchError(given.error))
		},
		Entry("domain starts with dot", testCase{
			config: dns_server.Config{
				Domain:         ".mesh",
				CIDR:           "240.0.0.0/4",
				ServiceVipPort: 80,
			},
			error: "domain must not start with a dot",
		}),
		Entry("invalid CIDR", testCase{
			config: dns_server.Config{
				Domain:         "mesh",
				CIDR:           "not-a-cidr",
				ServiceVipPort: 80,
			},
			error: "CIDR must be valid",
		}),
		Entry("port is zero", testCase{
			config: dns_server.Config{
				Domain:         "mesh",
				CIDR:           "240.0.0.0/4",
				ServiceVipPort: 0,
			},
			error: "port can't be 0",
		}),
	)

	It("should accept valid config", func() {
		cfg := dns_server.DefaultDNSServerConfig()
		Expect(cfg.Validate()).To(Succeed())
	})
})
