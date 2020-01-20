package install

import (
	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
	postgres_schema "github.com/Kong/kuma/app/kumactl/pkg/install/universal/control-plane/postgres"
	kuma_cmd "github.com/Kong/kuma/pkg/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newInstallDatabaseSchemaCmd() *cobra.Command {
	args := struct {
		target string
	}{}
	cmd := &cobra.Command{
		Use:   "database-schema",
		Short: "Install Kuma schema on DB",
		Long:  `Install Kuma schema on DB.`,
		Example: `1. kumactl install database-schema --target=postgres | PGPASSWORD=mysecretpassword psql -h localhost -U postgres
2. sql_file=$(mktemp) ; \ 
kumactl install database-schema --target=postgres >$sql_file ; \
psql --host=localhost --username=postgres --password --file=$sql_file ; \
rm $sql_file`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			switch args.target {
			case "postgres":
				file, err := data.ReadFile(postgres_schema.Schema, "resource.sql")
				if err != nil {
					return errors.Wrap(err, "could not read schema file")
				}
				_, err = cmd.OutOrStdout().Write(file.Data)
				return err
			default:
				return errors.Errorf("unknown target type: %s", args.target)
			}
		},
	}
	cmd.Flags().StringVar(&args.target, "target", "postgres", kuma_cmd.UsageOptions("Database type", "postgres"))
	return cmd
}
