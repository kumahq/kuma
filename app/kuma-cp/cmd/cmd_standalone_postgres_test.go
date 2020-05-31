// +build integration

package cmd

import (
	"github.com/Kong/kuma/pkg/config"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/config/core/resources/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
xdsServer:
  grpcPort: 0
  diagnosticsPort: %d
bootstrapServer:
  port: 0
apiServer:
  port: 0
sdsServer:
  grpcPort: 0
dataplaneTokenServer:
  local:
    port: 0
guiServer:
  port: 0
environment: universal
store:
  type: postgres
`))
})
