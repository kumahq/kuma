package install

import (
	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
	postgres_schema "github.com/Kong/kuma/app/kumactl/pkg/install/universal/control-plane/postgres"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newInstallPostgresSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "postgres-schema",
		Short:   "Install Kuma on Postgres DB.",
		Long:    `Install Kuma on Postgres DB.`,
		Example: "kumactl install postgres-schema | PGPASSWORD=mysecretpassword psql -h localhost -U postgres",
		RunE: func(cmd *cobra.Command, _ []string) error {
			file, err := data.ReadFile(postgres_schema.Schema, "resource.sql")
			if err != nil {
				return errors.Wrap(err, "could not read schema file")
			}
			_, err = cmd.OutOrStdout().Write(file)
			return err
		},
	}
}
