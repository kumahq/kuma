package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kumahq/kuma/pkg/core/runtime"
)

const duplicateKeyErrorMsg = "duplicate key value violates unique constraint"

var noopPgxCustomizer runtime.PgxConfigCustomizer = func(pgxConfig *pgxpool.Config) {}
