package store

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"os"
	"regexp"
	"slices"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
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
			err = rs.Create(ctx, resource, store.CreateByKey(resource.GetMeta().GetName(), resource.GetMeta().GetMesh()), store.CreateWithLabels(resource.GetMeta().GetLabels()))
		} else {
			_ = curResource.SetSpec(resource.GetSpec())
			err = rs.Update(ctx, curResource, store.UpdateWithLabels(resource.GetMeta().GetLabels()))
		}
		if err != nil {
			return errors.Wrapf(err, "failed with entry %d meta: %s", i, resource.GetMeta())
		}
	}
	return nil
}

func ExtractResources(ctx context.Context, rs store.ResourceStore) (string, error) {
	outputs := &bytes.Buffer{}
	descriptors := registry.Global().ObjectDescriptors()
	slices.SortFunc(descriptors, func(a, b model.ResourceTypeDescriptor) int {
		return cmp.Compare(a.Name, b.Name)
	})
	for _, tDesc := range descriptors {
		resList := tDesc.NewList()
		err := rs.List(ctx, resList, store.ListOrdered())
		if err != nil {
			return "", err
		}

		items := slices.SortedFunc(
			slices.Values(resList.GetItems()),
			func(a, b model.Resource) int {
				return cmp.Compare(a.GetMeta().GetName(), b.GetMeta().GetName())
			},
		)

		if len(items) > 0 {
			_, _ = fmt.Fprintf(outputs, "# %s\n", tDesc.Name)
		}

		for i, resource := range items {
			if resource.Descriptor().Name == system.ZoneInsightType {
				zi := resource.(*system.ZoneInsightResource)
				zi.Spec.Subscriptions = nil
				zi.Spec.EnvoyAdminStreams = nil
				zi.Spec.KdsStreams = nil
			}
			entry := rest.From.Resource(resource)
			y, err := yaml.Marshal(entry)
			// Hardcore way to replace all times with 0001-01-01T00:00:00Z
			y = regexp.MustCompile(`[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9][^"]+`).ReplaceAll(y, []byte("0001-01-01T00:00:00Z"))
			// Hardcore way to replace uuids
			y = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}`).ReplaceAll(y, []byte("STRIPPED_UUID"))
			// Hardcore way to replace versions
			y = regexp.MustCompile(`(version|buildDate|gitCommit|gitTag): .*`).ReplaceAll(y, []byte("$1: STRIPPED_VERSION"))
			y = regexp.MustCompile(`(config): .*`).ReplaceAll(y, []byte("$1: STRIPPED_CONFIG"))
			if err != nil {
				return "", errors.Wrapf(err, "failed to marshal resource %d", i)
			}
			_, _ = fmt.Fprintf(outputs, "---\n%s\n", y)
		}
	}
	return outputs.String(), nil
}
