package store

import (
	"context"
	"os"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
)

func LoadResourcesFromFile(ctx context.Context, rs store.ResourceStore, fileName string) error {
	d, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	return LoadResources(ctx, rs, string(d))
}

func LoadResources(ctx context.Context, rs store.ResourceStore, inputs string) error {
	rawResources := util_yaml.SplitYAML(inputs)
	for i, rawResource := range rawResources {
		resource, err := rest.YAML.UnmarshalCore([]byte(rawResource))
		if err != nil {
			return errors.Wrapf(err, "failed to parse yaml %d", i)
		}
		curResource := resource.Descriptor().NewObject()
		create := false
		if err := rs.Get(ctx, curResource, store.GetByKey(resource.GetMeta().GetName(), resource.GetMeta().GetMesh())); err != nil {
			if !store.IsResourceNotFound(err) {
				return err
			}
			create = true
		}

		if create {
			err = rs.Create(ctx, resource, store.CreateByKey(resource.GetMeta().GetName(), resource.GetMeta().GetMesh()))
		} else {
			_ = curResource.SetSpec(resource.GetSpec())
			err = rs.Update(ctx, curResource)
		}
		if err != nil {
			return errors.Wrapf(err, "failed with entry %d meta: %s", i, resource.GetMeta())
		}
	}
	return nil
}
