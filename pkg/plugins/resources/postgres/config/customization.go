package config

import "github.com/jackc/pgx/v5/pgxpool"

type PgxConfigCustomizationFn func(pgxConfig *pgxpool.Config)

func (p PgxConfigCustomizationFn) Customize(pgxConfig *pgxpool.Config) {
	p(pgxConfig)
}

var DefaultPgxConfigCustomizationFn = PgxConfigCustomizationFn(func(pgxConfig *pgxpool.Config) {})
