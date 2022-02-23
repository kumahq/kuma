package cmd

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	test_postgres "github.com/kumahq/kuma/pkg/test/store/postgres"
)

var _ = Describe("Standalone Postgres test", func() {
	var c test_postgres.PostgresContainer
	var pgCfg *postgres.PostgresStoreConfig

	BeforeEach(func() {
		// setup migrate DB
		var err error
		c = test_postgres.PostgresContainer{}
		Expect(c.Start()).To(Succeed())
		pgCfg, err = c.Config()
		Expect(err).ToNot(HaveOccurred())
		cfg := kuma_cp.DefaultConfig()
		cfg.Store.Type = store.PostgresStore
		cfg.Store.Postgres = pgCfg
		err = config.Load("", &cfg)
		Expect(err).ToNot(HaveOccurred())
		err = migrate(cfg)
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		Expect(c.Stop()).To(Succeed())
	})

	RunSmokeTest(ConfigFactoryFunc(func() string {
		return fmt.Sprintf(`
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
  postgres:
    host: %s
    port: %d
diagnostics:
  serverPort: %%d
`, pgCfg.Host, pgCfg.Port)
	}), "./kuma-workdir")
})
