package export

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_system "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type exportContext struct {
	*kumactl_cmd.RootContext

	args struct {
		profile string
		format  string
	}
}

const (
	profileFederation = "federation"

	formatUniversal = "universal"
)

var profiles = map[string][]model.ResourceType{
	profileFederation: {
		core_mesh.MeshType,
		core_system.GlobalSecretType,
		core_system.SecretType,
	},
}

func NewExportCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &exportContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export Kuma resources",
		Long:  `Export Kuma resources.`,
		Example: `
Export Kuma resources
$ kumactl export --profile federation --format universal > policies.yaml
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := pctx.CheckServerVersionCompatibility(); err != nil {
				cmd.PrintErrln(err)
			}

			resTypes, ok := profiles[ctx.args.profile]
			if !ok {
				return errors.New("invalid profile")
			}

			if ctx.args.format != formatUniversal {
				return errors.New("invalid format")
			}

			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			meshes := &core_mesh.MeshResourceList{}
			if err := rs.List(cmd.Context(), meshes); err != nil {
				return errors.Wrap(err, "could not list meshes")
			}

			var resources []model.Resource
			for _, resType := range resTypes {
				resDesc, err := pctx.Runtime.Registry.DescriptorFor(resType)
				if err != nil {
					return err
				}
				if resDesc.Scope == model.ScopeGlobal {
					list := resDesc.NewList()
					if err := rs.List(cmd.Context(), list); err != nil {
						return errors.Wrapf(err, "could not list %q", resType)
					}
					resources = append(resources, list.GetItems()...)
				} else {
					for _, mesh := range meshes.Items {
						list := resDesc.NewList()
						if err := rs.List(cmd.Context(), list, store.ListByMesh(mesh.GetMeta().GetName())); err != nil {
							return errors.Wrapf(err, "could not list %q", resType)
						}
						resources = append(resources, list.GetItems()...)
					}
				}
			}

			for _, res := range resources {
				if _, err := cmd.OutOrStdout().Write([]byte("---\n")); err != nil {
					return err
				}
				if err := printers.GenericPrint(output.YAMLFormat, res, table.Table{}, cmd.OutOrStdout()); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&ctx.args.profile, "profile", "p", profileFederation, fmt.Sprintf(`Profile. Available values: "%s"`, profileFederation))
	cmd.Flags().StringVarP(&ctx.args.format, "format", "f", formatUniversal, fmt.Sprintf(`Policy format output. Available values: "%q"`, formatUniversal))
	return cmd
}
