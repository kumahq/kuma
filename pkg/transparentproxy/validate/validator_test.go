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
			validator := createValidator(false, ValidationServerPort)

			// then
			Expect(validator.Config.ServerListenIP.String()).To(Equal("127.0.0.1"))
		})

		It("ipv6", func() {
			// when
			validator := createValidator(true, ValidationServerPort)

			// then
			serverIP := validator.Config.ServerListenIP.String()
			Expect(serverIP).To(Equal("::1"))

			splitByCon := strings.Split(serverIP, ":")
			Expect(len(splitByCon)).To(BeNumerically(">", 2))
		})
	})

	It("should return pass when connect to correct address", func() {
		// given
		validator := createValidator(false, uint16(0))
		ipAddr := "127.0.0.1"
		addr, _ := netip.ParseAddr(ipAddr)
		validator.Config.ServerListenIP = addr
		validator.Config.ClientConnectIP = addr

		// when
		err := validator.Run()

		// then
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return fail when no iptables rules setup", func() {
		// given
		validator := createValidator(false, uint16(0))
		validator.Config.ClientConnectPort = ValidationServerPort
		validator.Config.ClientRetryInterval = 30 * time.Millisecond // just to make test faster and there should be no flakiness here because the connection will never establish successfully without the redirection

		// when
		err := validator.Run()

		// then
		Expect(err).To(HaveOccurred())
		errMsg := err.Error()
		containsTimeout := strings.Contains(errMsg, "i/o timeout")
		containsRefused := strings.Contains(errMsg, "refused")
		Expect(containsTimeout || containsRefused).To(BeTrue())
	})
})

func createValidator(ipv6Enabled bool, validationServerPort uint16) *Validator {
	return NewValidator(ipv6Enabled, validationServerPort, core.NewLoggerTo(os.Stdout, kuma_log.InfoLevel).WithName("validator"))
}
