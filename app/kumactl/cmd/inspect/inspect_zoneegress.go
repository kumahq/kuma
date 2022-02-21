package inspect

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

const inspectZoneEgressError = "Policies are not applied on ZoneEgress, please use '--config-dump' flag to get " +
	"envoy config dump of the ZoneEgress"

func newInspectZoneEgressCmd(pctx *cmd.RootContext) *cobra.Command {
	var configDump bool
	cmd := &cobra.Command{
		Use:   "zoneegress NAME",
		Short: "Inspect ZoneEgress",
		Long:  "Inspect ZoneEgress.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !configDump {
				_, err := fmt.Fprintln(cmd.OutOrStderr(), inspectZoneEgressError)
				return err
			}
			client, err := pctx.CurrentInspectEnvoyProxyClient(mesh.ZoneEgressResourceTypeDescriptor)
			if err != nil {
				return errors.Wrap(err, "failed to create a zoneegress inspect client")
			}
			name := args[0]
			bytes, err := client.ConfigDump(context.Background(), core_model.ResourceKey{Name: name, Mesh: core_model.NoMesh})
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), string(bytes))
			return err
		},
	}
	cmd.PersistentFlags().BoolVar(&configDump, "config-dump", false, "if set then the command returns envoy config dump for provided dataplane")
	return cmd
}
