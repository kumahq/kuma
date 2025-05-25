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

func (m *Persistence) Get(ctx context.Context, meshes []string) (map[string]*VirtualOutboundMeshView, error) {
	resourcesByMesh, err := m.getGroupedByMesh(ctx)
	if err != nil {
		return nil, err
	}

	virtualOutboundByMesh := map[string]*VirtualOutboundMeshView{}
	for _, mesh := range meshes {
		resource, ok := resourcesByMesh[mesh]
		if !ok {
			virtualOutboundByMesh[mesh] = NewEmptyVirtualOutboundView()
			continue
		}

		virtualOutboundView, err := m.configToVirtualOutboundMeshView(resource)
		if err != nil {
			return nil, err
		}

		virtualOutboundByMesh[mesh] = virtualOutboundView
	}

	return virtualOutboundByMesh, nil
}

func (m *Persistence) GetByMesh(ctx context.Context, mesh string) (*VirtualOutboundMeshView, error) {
	name := fmt.Sprintf(template, mesh)
	resource := config_model.NewConfigResource()
	err := m.configManager.Get(ctx, resource, store.GetByKey(name, ""))
	if err != nil {
		if store.IsNotFound(err) {
			return NewEmptyVirtualOutboundView(), nil
		}
		return nil, err
	}

	return m.configToVirtualOutboundMeshView(resource)
}

func (m *Persistence) Set(ctx context.Context, mesh string, vips *VirtualOutboundMeshView) error {
	name := fmt.Sprintf(template, mesh)
	resource := config_model.NewConfigResource()
	create := false
	if err := m.configManager.Get(ctx, resource, store.GetByKey(name, model.NoMesh)); err != nil {
		if !store.IsNotFound(err) {
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

func (m *Persistence) getGroupedByMesh(ctx context.Context) (map[string]*config_model.ConfigResource, error) {
	resources := config_model.ConfigResourceList{}
	err := m.configManager.List(ctx, &resources, store.ListByFilterFunc(func(rs model.Resource) bool {
		return re.FindString(rs.GetMeta().GetName()) != ""
	}))
	if err != nil {
		return nil, err
	}

	resourceByMesh := map[string]*config_model.ConfigResource{}
	for _, resource := range resources.Items {
		mesh, _ := MeshFromConfigKey(resource.GetMeta().GetName())
		resourceByMesh[mesh] = resource
	}

	return resourceByMesh, nil
}

func (m *Persistence) configToVirtualOutboundMeshView(resource *config_model.ConfigResource) (*VirtualOutboundMeshView, error) {
	if resource.Spec.Config == "" {
		return NewEmptyVirtualOutboundView(), nil
	}

	var virtualOutboundView *VirtualOutboundMeshView

	// try reading new format
	emptyTagFirstOutboundView := NewEmptyTagFirstOutboundView()
	if err := json.Unmarshal([]byte(resource.Spec.GetConfig()), &emptyTagFirstOutboundView); err != nil {
		return nil, err
	}
	if len(emptyTagFirstOutboundView.PerService) != 0 {
		return emptyTagFirstOutboundView.ToVirtualOutboundView(), nil
	}

	// fall back to default
	res := map[HostnameEntry]VirtualOutbound{}
	if err := json.Unmarshal([]byte(resource.Spec.GetConfig()), &res); err != nil {
		return nil, err
	}
	virtualOutboundView, err := NewVirtualOutboundView(res)
	if err != nil {
		return nil, err
	}

	return virtualOutboundView, nil
}
