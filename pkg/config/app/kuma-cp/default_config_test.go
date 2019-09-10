package kuma_cp

import (
	"github.com/Kong/kuma/pkg/config"
	"github.com/Kong/kuma/pkg/config/core/discovery"
	"github.com/Kong/kuma/pkg/config/plugins/discovery/universal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"strings"
	"time"
)

var _ = Describe("Default config", func() {

	var backupEnvVars []string

	BeforeEach(func() {
		backupEnvVars = os.Environ()
		os.Clearenv()
	})
	AfterEach(func() {
		for _, envVar := range backupEnvVars {
			parts := strings.SplitN(envVar, "=", 2)
			Expect(os.Setenv(parts[0], parts[1])).To(Succeed())
		}
	})

	It("should be check agains the kuma-cp.defaults.yaml file", func() {
		// given
		cfg := Config{
			Discovery: &discovery.DiscoveryConfig{
				Universal: &universal.UniversalDiscoveryConfig{
					// todo(jakubdyszkiewicz) this will be removed in the next versions of Kuma
					PollingInterval: time.Second,
				},
			},
		}

		// when
		err := config.Load("kuma-cp.defaults.yaml", &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(DefaultConfig()).To(Equal(cfg))
	})
})
