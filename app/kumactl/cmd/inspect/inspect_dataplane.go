package inspect

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	api_common "github.com/kumahq/kuma/v3/api/openapi/types/common"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output/printers"
	kuma_cmd "github.com/kumahq/kuma/v3/pkg/cmd"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

const (
	InspectionTypePolicies   = "policies"
	InspectionTypeConfigDump = "config-dump"
	InspectionTypeStats      = "stats"
	InspectionTypeClusters   = "clusters"
	InspectionConfig         = "config"
)

func newInspectDataplaneCmd(pctx *cmd.RootContext) *cobra.Command {
	var includeEDS bool
	var inspectionType string
	var shadow bool
	var include []string
	cmd := &cobra.Command{
		Use:   "dataplane NAME",
		Short: "Inspect Dataplane",
		Long:  "Inspect Dataplane.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if shadow && inspectionType != InspectionConfig {
				return errors.New("flag '--shadow' can be used only when '--type=config'")
			}
			if len(include) > 0 && inspectionType != InspectionConfig {
				return errors.New("flag '--include' can be used only when '--type=config'")
			}
			if includeEDS && inspectionType != InspectionTypeConfigDump {
				return errors.New(fmt.Sprintf("flag '--include-eds' can be used only when '--type=%s'", InspectionTypeConfigDump))
			}

			client, err := pctx.CurrentInspectEnvoyProxyClient(mesh.DataplaneResourceTypeDescriptor)
			if err != nil {
				return errors.Wrap(err, "failed to create a dataplane inspect client")
			}

			resourceKey := core_model.ResourceKey{Name: name, Mesh: pctx.CurrentMesh()}
			switch inspectionType {
			case InspectionTypePolicies:
				client, err := pctx.CurrentDataplaneInspectClient()
				if err != nil {
					return errors.Wrap(err, "failed to create a dataplane inspect client")
				}

				policies, err := client.InspectPolicies(context.Background(), pctx.CurrentMesh(), name)
				if err != nil {
					return err
				}
				format := output.Format(pctx.InspectContext.Args.OutputFormat)
				return printers.GenericPrint(format, policies, printers.Table{
					Headers: []string{"Kind", "Origins"},
					RowForItem: func(i int, container any) ([]string, error) {
						list, ok := container.(api_common.PoliciesList)
						if !ok {
							return nil, errors.Errorf("unexpected container type %T, expected %T", container, api_common.PoliciesList{})
						}
						items := list.Policies
						if i >= len(items) {
							return nil, nil
						}
						itm := items[i]
						origins := make([]string, len(itm.Origins))
						for j, origin := range itm.Origins {
							origins[j] = origin.Kri
						}
						return []string{itm.Kind, strings.Join(origins, ",")}, nil
					},
				}, cmd.OutOrStdout())
			case InspectionTypeConfigDump:
				bytes, err := client.ConfigDump(context.Background(), resourceKey, includeEDS)
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
			case InspectionConfig:
				bytes, err := client.Config(context.Background(), resourceKey, shadow, include)
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), string(bytes))
				return err
			default:
				return errors.New("invalid inspection type")
			}
		},
	}
	cmd.PersistentFlags().StringVar(&inspectionType, "type", InspectionTypePolicies, kuma_cmd.UsageOptions("inspection type", InspectionTypePolicies, InspectionTypeConfigDump, InspectionTypeStats, InspectionTypeClusters, InspectionConfig))
	cmd.PersistentFlags().BoolVar(&includeEDS, "include-eds", false, "include EDS when dumping envoy config for dataplane")
	cmd.PersistentFlags().StringVarP(&pctx.Args.Mesh, "mesh", "m", "default", "mesh to use")
	cmd.PersistentFlags().BoolVar(&shadow, "shadow", false, "when computing XDS config the CP takes into account policies with 'kuma.io/effect: shadow' label")
	cmd.PersistentFlags().StringSliceVar(&include, "include", []string{}, "an array of extra fields to include in the response. When `include=diff` the server computes a diff in JSONPatch format between the XDS config returned in 'xds' and the current proxy XDS config.")
	return cmd
}
