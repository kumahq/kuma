package apply

import (
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/registry"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
)

type applyContext struct {
	*cmd.RootContext

	args struct {
		file string
	}
}

func NewApplyCmd(pctx *cmd.RootContext) *cobra.Command {
	ctx := &applyContext{RootContext: pctx}
	command := &cobra.Command{
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

			res, err := parseResource(b)
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
	command.PersistentFlags().StringVarP(&ctx.args.file, "file", "f", "", "Path to file to apply")
	return command
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
