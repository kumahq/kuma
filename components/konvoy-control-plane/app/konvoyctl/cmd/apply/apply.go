package apply

import (
	"context"
	konvoyctl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/registry"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	"github.com/ghodss/yaml"
	"github.com/hoisie/mustache"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
)

type applyContext struct {
	*konvoyctl_cmd.RootContext

	args struct {
		file string
		arg  map[string]string
	}
}

func NewApplyCmd(pctx *konvoyctl_cmd.RootContext) *cobra.Command {
	ctx := &applyContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Create or modify Konvoy resources",
		Long:  `Create or modify Konvoy resources.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var b []byte
			var err error

			if ctx.args.file == "" || ctx.args.file == "-" {
				b, err = ioutil.ReadAll(cmd.InOrStdin())
			} else {
				b, err = ioutil.ReadFile(ctx.args.file)
			}
			if err != nil {
				return errors.Wrap(err, "error while reading provided file")
			}

			configBytes, err := processConfigTemplate(string(b), ctx.args.arg)
			if err != nil {
				return errors.Wrap(err, "error compiling config from template")
			}

			res, err := parseResource(configBytes)
			if err != nil {
				return errors.Wrap(err, "yaml contains invalid resource")
			}
			cp, err := pctx.CurrentControlPlane()
			if err != nil {
				return err
			}
			rs, err := pctx.NewResourceStore(cp)
			if err != nil {
				return err
			}

			if err := upsert(rs, res); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&ctx.args.file, "file", "f", "", "Path to file to apply")
	cmd.PersistentFlags().StringToStringVarP(&ctx.args.arg, "arg", "a", map[string]string{}, "Variable to replace in configuration")
	return cmd
}

func processConfigTemplate(config string, values map[string]string) ([]byte, error) {
	// TODO error checking -- match number of placeholders with number of
	// passed values
	data := mustache.Render(config, values)
	return []byte(data), nil
}

func upsert(rs store.ResourceStore, res model.Resource) error {
	newRes, err := registry.Global().NewObject(res.GetType())
	if err != nil {
		return err
	}
	meta := res.GetMeta()
	if err := rs.Get(context.Background(), newRes, store.GetByKey(meta.GetNamespace(), meta.GetName(), meta.GetMesh())); err != nil {
		if store.IsResourceNotFound(err) {
			return rs.Create(context.Background(), res, store.CreateByKey(meta.GetNamespace(), meta.GetName(), meta.GetMesh()))
		} else {
			return err
		}
	}
	if err := newRes.SetSpec(res.GetSpec()); err != nil {
		return err
	}
	return rs.Update(context.Background(), newRes)
}

func parseResource(bytes []byte) (model.Resource, error) {
	resMeta := rest.ResourceMeta{}
	if err := yaml.Unmarshal(bytes, &resMeta); err != nil {
		return nil, err
	}
	resource, err := registry.Global().NewObject(model.ResourceType(resMeta.Type))
	if err != nil {
		return nil, err
	}
	if err := proto.FromYAML(bytes, resource.GetSpec()); err != nil {
		return nil, err
	}
	resource.SetMeta(meta{
		Name: resMeta.Name,
		Mesh: resMeta.Mesh,
	})
	return resource, nil
}

var _ model.ResourceMeta = &meta{}

type meta struct {
	Name string
	Mesh string
}

func (m meta) GetName() string {
	return m.Name
}

func (m meta) GetNamespace() string {
	return "default"
}

func (m meta) GetVersion() string {
	return ""
}

func (m meta) GetMesh() string {
	return m.Mesh
}
