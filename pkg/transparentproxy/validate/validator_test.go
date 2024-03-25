package validate

import (
	"fmt"
	"github.com/kumahq/kuma/pkg/core"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/netip"
	"os"
	"strings"
	"time"
)

var _ = Describe("Should Validate iptables rules", func() {
	Describe("generate default validator config", func() {
		It("ipv4", func() {
			// when
			validator := createValidator(false)

			// then
			Expect(validator.Config.ServerListenAddress).To(Equal("127.0.0.1:15010"))

			lastIdxOfColon := strings.LastIndex(validator.Config.ServerListenAddress, ":")
			serverIP := validator.Config.ServerListenAddress[0:lastIdxOfColon]
			Expect(serverIP).ToNot(Equal(validator.Config.ClientConnectIP.String()))
		})

		It("ipv6", func() {
			// when
			validator := createValidator(true)

			// then
			Expect(validator.Config.ServerListenAddress).To(Equal("[::1]:15010"))

			splitByCon := strings.Split(validator.Config.ServerListenAddress, ":")
			Expect(len(splitByCon) > 2).To(BeTrue())
		})
	})

	It("should return pass when connect to correct address", func() {
		// when
		validator := createValidator(false)
		ipAddr := "127.0.0.1"
		addr, _ := netip.ParseAddr(ipAddr)
		validator.Config.ServerListenAddress = fmt.Sprintf("%s:%d", ipAddr, ValidationServerPort)
		validator.Config.ClientConnectIP = addr

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
	return NewValidator(ipv6Enabled, core.NewLoggerTo(os.Stdout, kuma_log.InfoLevel).WithName("validator"))
}
