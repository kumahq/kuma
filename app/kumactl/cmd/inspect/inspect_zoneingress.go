package inspect

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

const inspectZoneIngressError = "Policies are not applied on ZoneIngress, please use '--config-dump' flag to get " +
	"envoy config dump of the ZoneIngress"

func newInspectZoneIngressCmd(pctx *cmd.RootContext) *cobra.Command {
	var configDump bool
	cmd := &cobra.Command{
		Use:   "zoneingress NAME",
		Short: "Inspect ZoneIngress",
		Long:  "Inspect ZoneIngress.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !configDump {
				_, err := fmt.Fprintln(cmd.OutOrStderr(), inspectZoneIngressError)
				return err
			}
			client, err := pctx.CurrentZoneIngressInspectClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a zoneingress inspect client")
			}
			name := args[0]
			bytes, err := client.InspectConfigDump(context.Background(), name)
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
