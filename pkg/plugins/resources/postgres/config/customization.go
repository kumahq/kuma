package config

import "github.com/jackc/pgx/v5/pgxpool"

type PgxConfigCustomization interface {
	Customize(pgxConfig *pgxpool.Config)
}

type PgxConfigCustomizationFn func(pgxConfig *pgxpool.Config)

func (p PgxConfigCustomizationFn) Customize(pgxConfig *pgxpool.Config) {
	p(pgxConfig)
}

var NoopPgxConfigCustomizationFn = PgxConfigCustomizationFn(func(pgxConfig *pgxpool.Config) {})
