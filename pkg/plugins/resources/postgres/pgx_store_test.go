package postgres_test

import (
	"context"

	"github.com/jackc/pgx/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_postgres "github.com/kumahq/kuma/v2/pkg/config/plugins/resources/postgres"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/v2/pkg/metrics"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/postgres"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/postgres/config"
)

// failingRowsTx is a pgx.Tx whose queries end with an error after the last row,
// like a query whose context is cancelled while rows are being read.
type failingRowsTx struct {
	pgx.Tx
}

func (failingRowsTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return failingRows{}, nil
}

type failingRows struct {
	pgx.Rows
}

func (failingRows) Next() bool { return false }
func (failingRows) Close()     {}
func (failingRows) Err() error { return context.Canceled }

var _ = Describe("PgxStore", func() {
	It("should return the error of the rows iteration from List", func() {
		dbCfg, err := c.Config()
		Expect(err).ToNot(HaveOccurred())
		dbCfg.DriverName = config_postgres.DriverNamePgx
		_, err = postgres.MigrateDb(dbCfg)
		Expect(err).ToNot(HaveOccurred())
		metrics, err := core_metrics.NewMetrics("Zone")
		Expect(err).ToNot(HaveOccurred())
		pStore, err := postgres.NewPgxStore(metrics, dbCfg, config.NoopPgxConfigCustomizationFn)
		Expect(err).ToNot(HaveOccurred())

		ctx := store.CtxWithTx(context.Background(), failingRowsTx{})
		list := &core_mesh.MeshResourceList{}
		err = pStore.List(ctx, list)

		Expect(err).To(MatchError(context.Canceled))
		Expect(list.Items).To(BeEmpty())
	})
})
