package catalog

import (
	"context"
	"encoding/json"
	"sort"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type ConfigInstances struct {
	Instances []Instance `json:"instances"`
}

var CatalogKey = model.ResourceKey{
	Name: "cp-catalog",
}

type ConfigCatalog struct {
	resManager manager.ResourceManager
	ConfigCatalogReader
}

var _ Catalog = &ConfigCatalog{}

func NewConfigCatalog(resManager manager.ResourceManager) Catalog {
	return &ConfigCatalog{
		resManager: resManager,
		ConfigCatalogReader: ConfigCatalogReader{
			resManager: resManager,
		},
	}
}

func (c *ConfigCatalog) Replace(ctx context.Context, instances []Instance) (bool, error) {
	sort.Stable(InstancesByID(instances))
	bytes, err := json.Marshal(ConfigInstances{
		Instances: instances,
	})
	if err != nil {
		return false, nil
	}
	newConfig := string(bytes)
	var updated bool
	err = manager.Upsert(ctx, c.resManager, CatalogKey, system.NewConfigResource(), func(resource model.Resource) error {
		if resource.(*system.ConfigResource).Spec.GetConfig() != newConfig {
			resource.(*system.ConfigResource).Spec = &system_proto.Config{
				Config: newConfig,
			}
			updated = true
		}
		return nil
	})
	return updated, err
}

func (c *ConfigCatalog) ReplaceLeader(ctx context.Context, leader Instance) error {
	return manager.Upsert(ctx, c.resManager, CatalogKey, system.NewConfigResource(), func(resource model.Resource) error {
		instances := &ConfigInstances{}
		if cfg := resource.(*system.ConfigResource).Spec.GetConfig(); cfg != "" {
			if err := json.Unmarshal([]byte(cfg), instances); err != nil {
				return err
			}
		}
		leaderFound := false
		for i, instance := range instances.Instances {
			instance.Leader = false
			if instance.Id == leader.Id {
				instance.Leader = true
				leaderFound = true
			}
			instances.Instances[i] = instance
		}
		if !leaderFound {
			instances.Instances = append(instances.Instances, leader)
			sort.Stable(InstancesByID(instances.Instances))
		}
		bytes, err := json.Marshal(instances)
		if err != nil {
			return err
		}
		resource.(*system.ConfigResource).Spec = &system_proto.Config{
			Config: string(bytes),
		}
		return nil
	})
}

func (c *ConfigCatalog) DropLeader(ctx context.Context, leader Instance) error {
	return manager.Upsert(ctx, c.resManager, CatalogKey, system.NewConfigResource(), func(resource model.Resource) error {
		instances := &ConfigInstances{}
		if cfg := resource.(*system.ConfigResource).Spec.GetConfig(); cfg != "" {
			if err := json.Unmarshal([]byte(cfg), instances); err != nil {
				return err
			}
		}
		for i, instance := range instances.Instances {
			if instance.Id == leader.Id {
				instance.Leader = false
			}
			instances.Instances[i] = instance
		}
		bytes, err := json.Marshal(instances)
		if err != nil {
			return err
		}
		resource.(*system.ConfigResource).Spec = &system_proto.Config{
			Config: string(bytes),
		}
		return nil
	})
}

type ConfigCatalogReader struct {
	resManager manager.ReadOnlyResourceManager
}

var _ Reader = &ConfigCatalogReader{}

func NewConfigCatalogReader(resManager manager.ReadOnlyResourceManager) Reader {
	return &ConfigCatalogReader{
		resManager: resManager,
	}
}

func (c *ConfigCatalogReader) Instances(ctx context.Context) ([]Instance, error) {
	cfg := system.NewConfigResource()
	if err := c.resManager.Get(ctx, cfg, store.GetBy(CatalogKey)); err != nil {
		if store.IsNotFound(err) {
			return []Instance{}, nil
		}
		return nil, err
	}
	var instances ConfigInstances
	if err := json.Unmarshal([]byte(cfg.Spec.Config), &instances); err != nil {
		return nil, err
	}
	return instances.Instances, nil
}
