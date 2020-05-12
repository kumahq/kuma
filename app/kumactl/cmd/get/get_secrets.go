package get

import (
	"context"
	"io"
	"time"

	"github.com/Kong/kuma/app/kumactl/pkg/output/table"

	"github.com/Kong/kuma/pkg/core/resources/apis/system"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

func newGetSecretsCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Show Secrets",
		Long:  `Show Secrets.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentAdminResourceStore()
			if err != nil {
				return err
			}

			secrets := &system.SecretResourceList{}
			if err := rs.List(context.Background(), secrets, core_store.ListByMesh(pctx.CurrentMesh())); err != nil {
				return errors.Wrapf(err, "failed to list Secrets")
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printSecrets(pctx.Now(), secrets.Items, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(secrets), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printSecrets(rootTime time.Time, secrets []*system.SecretResource, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(secrets) <= i {
					return nil
				}
				secret := secrets[i]

				return []string{
					secret.Meta.GetMesh(), // MESH
					secret.Meta.GetName(), // NAME
					table.TimeSince(secret.Meta.GetModificationTime(), rootTime), //AGE
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
