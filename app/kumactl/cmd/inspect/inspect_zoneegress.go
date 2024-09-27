package inspect

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

const inspectZoneEgressError = "Policies are not applied on ZoneEgress, please use '--config-dump' flag to get " +
	"envoy config dump of the ZoneEgress"

func newInspectZoneEgressCmd(pctx *cmd.RootContext) *cobra.Command {
	var configDump bool
	var inspectionType string
	cmd := &cobra.Command{
		Use:   "zoneegress NAME",
		Short: "Inspect ZoneEgress",
		Long:  "Inspect ZoneEgress.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if configDump {
				inspectionType = InspectionTypeConfigDump
			}

			client, err := pctx.CurrentInspectEnvoyProxyClient(mesh.ZoneEgressResourceTypeDescriptor)
			if err != nil {
				return errors.Wrap(err, "failed to create a zoneegress inspect client")
			}
			name := args[0]
			resourceKey := core_model.ResourceKey{Name: name, Mesh: core_model.NoMesh}

			switch inspectionType {
			case InspectionTypeConfigDump:
				bytes, err := client.ConfigDump(context.Background(), resourceKey, true)
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), string(bytes))
				return err
			case InspectionTypeStats:
				bytes, err := client.Stats(context.Background(), resourceKey)
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), string(bytes))
				return err
			case InspectionTypeClusters:
				bytes, err := client.Clusters(context.Background(), resourceKey)
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), string(bytes))
				return err
			case InspectionTypePolicies:
				return errors.New(inspectZoneEgressError)
			default:
				return errors.New("invalid inspection type")
			}
		},
	}
	cmd.PersistentFlags().BoolVar(&configDump, "config-dump", false, "if set then the command returns envoy config dump for provided dataplane")
	_ = cmd.PersistentFlags().MarkDeprecated("config-dump", "use --type=config-dump")
	cmd.PersistentFlags().StringVar(&inspectionType, "type", InspectionTypeConfigDump, kuma_cmd.UsageOptions("inspection type", InspectionTypeConfigDump, InspectionTypeStats, InspectionTypeClusters))
	return cmd
}
