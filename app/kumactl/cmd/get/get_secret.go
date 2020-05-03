package get

import (
	"context"

	"github.com/Kong/kuma/pkg/core/resources/apis/system"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

func newGetSecretCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret NAME",
		Short: "Show a single Secret resource",
		Long:  `Show a single Secret resource.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentAdminResourceStore()
			if err != nil {
				return err
			}
			name := args[0]
			currentMesh := pctx.CurrentMesh()
			secret := &system.SecretResource{}
			if err := rs.Get(context.Background(), secret, store.GetByKey(name, currentMesh)); err != nil {
				if store.IsResourceNotFound(err) {
					return errors.Errorf("No resources found in %s mesh", currentMesh)
				}
				return errors.Wrapf(err, "failed to get a secret %s", currentMesh)
			}
			secrets := []*system.SecretResource{secret}
			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printSecrets(pctx.Now(), secrets, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(secret), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
