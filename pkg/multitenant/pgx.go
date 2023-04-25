package multitenant

import "github.com/jackc/pgx/v5/pgxpool"

type PgxConfigCustomization interface {
	Customize(pgxConfig *pgxpool.Config)
}

type DefaultPgxConfigCustomization struct{}

var _ PgxConfigCustomization = &DefaultPgxConfigCustomization{}

func (d DefaultPgxConfigCustomization) Customize(_ *pgxpool.Config) {}
