package validate

import (
	"net/netip"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core"
	kuma_log "github.com/kumahq/kuma/pkg/log"
)

var _ = Describe("Should Validate iptables rules", func() {
	Describe("generate default validator config", func() {
		It("ipv4", func() {
			// when
			validator := createValidator(false)

			// then
			Expect(validator.Config.ServerListenIP.String()).To(Equal("127.0.0.1"))
			Expect(validator.Config.ServerListenPort).To(Equal(uint16(15006)))
		})

		It("ipv6", func() {
			// when
			validator := createValidator(true)

			// then
			serverIP := validator.Config.ServerListenIP.String()
			Expect(serverIP).To(Equal("::1"))

			splitByCon := strings.Split(serverIP, ":")
			Expect(len(splitByCon)).To(BeNumerically(">", 2))
		})
	})

	It("should return pass when connect to correct address", func() {
		// when
		validator := createValidator(false)
		ipAddr := "127.0.0.1"
		addr, _ := netip.ParseAddr(ipAddr)
		validator.Config.ServerListenIP = addr
		validator.Config.ClientConnectIP = addr
		validator.Config.ClientConnectPort = ValidationServerPort

		err := validator.Run()

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return fail when no iptables rules setup", func() {
		// given
		validator := createValidator(false)
		validator.Config.ClientRetryInterval = 30 * time.Millisecond

		// when
		err := validator.Run()

		// then
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("i/o timeout"))
	})
})

func createValidator(ipv6Enabled bool) *Validator {
	return NewValidator(ipv6Enabled, ValidationServerPort, core.NewLoggerTo(os.Stdout, kuma_log.InfoLevel).WithName("validator"))
}
