package vips

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/pkg/errors"

	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type Persistence struct {
	configManager   config_manager.ConfigManager
	resourceManager manager.ReadOnlyResourceManager
	useTagFirstView bool
}

const template = "kuma-%s-dns-vips"

var re = regexp.MustCompile(fmt.Sprintf(template, `(.*)`))

func ConfigKey(mesh string) string {
	return fmt.Sprintf(template, mesh)
}

func MeshFromConfigKey(name string) (string, bool) {
	match := re.FindStringSubmatch(name)
	if len(match) < 2 {
		return "", false
	}
	mesh := match[1]
	return mesh, true
}

func NewPersistence(resourceManager manager.ReadOnlyResourceManager, configManager config_manager.ConfigManager, useTagFirstVirtualOutboundModel bool) *Persistence {
	return &Persistence{
		resourceManager: resourceManager,
		configManager:   configManager,
		useTagFirstView: useTagFirstVirtualOutboundModel,
	}
}

func (m *Persistence) GetByMesh(ctx context.Context, mesh string) (*VirtualOutboundMeshView, error) {
	name := fmt.Sprintf(template, mesh)
	resource := config_model.NewConfigResource()
	err := m.configManager.Get(ctx, resource, store.GetByKey(name, ""))
	if err != nil {
		if store.IsResourceNotFound(err) {
			return NewEmptyVirtualOutboundView(), nil
		}
		return nil, err
	}

	if resource.Spec.Config == "" {
		return NewEmptyVirtualOutboundView(), nil
	}

	var virtualOutboundView *VirtualOutboundMeshView
	// try reading new format
	res := NewEmptyTagFirstOutboundView()
	if err := json.Unmarshal([]byte(resource.Spec.GetConfig()), &res); err != nil {
		return nil, err
	}
	if len(res.PerService) != 0 {
		virtualOutboundView = res.ToVirtualOutboundView()
	} else {
		// fall back to default
		res := map[HostnameEntry]VirtualOutbound{}
		if err := json.Unmarshal([]byte(resource.Spec.GetConfig()), &res); err != nil {
			return nil, err
		}
		virtualOutboundView, err = NewVirtualOutboundView(res)
		if err != nil {
			return nil, err
		}
	}

	return virtualOutboundView, nil
}

func (m *Persistence) Set(ctx context.Context, mesh string, vips *VirtualOutboundMeshView) error {
	name := fmt.Sprintf(template, mesh)
	resource := config_model.NewConfigResource()
	create := false
	if err := m.configManager.Get(ctx, resource, store.GetByKey(name, model.NoMesh)); err != nil {
		if !store.IsResourceNotFound(err) {
			return err
		}
		create = true
	}

	var jsonBytes []byte
	var err error
	if m.useTagFirstView {
		view := NewTagFirstOutboundView(vips)
		jsonBytes, err = json.Marshal(view)
		if err != nil {
			return errors.Wrap(err, "unable to marshall VIP list")
		}
	} else {
		jsonBytes, err = json.Marshal(vips.byHostname)
		if err != nil {
			return errors.Wrap(err, "unable to marshall VIP list")
		}
	}

	resource.Spec.Config = string(jsonBytes)

	if create {
		meshRes := core_mesh.NewMeshResource()
		if err := m.resourceManager.Get(ctx, meshRes, store.GetByKey(mesh, model.NoMesh)); err != nil {
			return err
		}
		return m.configManager.Create(ctx, resource, store.CreateByKey(name, model.NoMesh), store.CreateWithOwner(meshRes))
	} else {
		return m.configManager.Update(ctx, resource)
	}
}
