// +build integration

package cmd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
)

var _ = Describe("Standalone Postgres test", func() {

	BeforeEach(func() {
		// setup migrate DB
		cfg := kuma_cp.DefaultConfig()
		err := config.Load("", &cfg)
		cfg.Store.Type = store.PostgresStore
		Expect(err).ToNot(HaveOccurred())
		err = migrate(cfg)
		Expect(err).ToNot(HaveOccurred())
	})

	RunSmokeTest(StaticConfig(`
general:
  workDir: ./kuma-workdir
apiServer:
  http:
    port: 0
  https:
    port: 0
dnsServer:
  port: 0
environment: universal
store:
  type: postgres
diagnostics:
  serverPort: %d
`), "./kuma-workdir")
})
