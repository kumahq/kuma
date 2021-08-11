package vips

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

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

func NewPersistence(resourceManager manager.ReadOnlyResourceManager, configManager config_manager.ConfigManager) *Persistence {
	return &Persistence{
		resourceManager: resourceManager,
		configManager:   configManager,
	}
}

func (m *Persistence) Get() (meshed map[string]*VirtualOutboundMeshView, errs error) {
	resourceList := &config_model.ConfigResourceList{}
	if err := m.configManager.List(context.Background(), resourceList); err != nil {
		return nil, err
	}

	meshed = map[string]*VirtualOutboundMeshView{}
	for _, resource := range resourceList.Items {
		mesh, ok := MeshFromConfigKey(resource.Meta.GetName())
		if !ok {
			continue
		}
		if resource.Spec.Config == "" {
			continue
		}
		v, err := m.unmarshal(resource.Spec.GetConfig())
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		meshed[mesh] = v
	}
	return
}

func (m *Persistence) unmarshal(config string) (*VirtualOutboundMeshView, error) {
	res := map[HostnameEntry]VirtualOutbound{}
	if err := json.Unmarshal([]byte(config), &res); err == nil {
		return NewVirtualOutboundView(res), nil
	}
	backCompat := map[HostnameEntry]string{}
	if err := json.Unmarshal([]byte(config), &backCompat); err != nil {
		// backwards compatibility
		backwardCompatible := map[string]string{}
		if err := json.Unmarshal([]byte(config), &backwardCompatible); err != nil {
			return nil, err
		}
		for service, vip := range backwardCompatible {
			backCompat[NewServiceEntry(service)] = vip
		}
	}
	for entry, vip := range backCompat {
		res[entry] = VirtualOutbound{
			Address: vip,
		}
	}
	return NewVirtualOutboundView(res), nil
}

func (m *Persistence) GetByMesh(mesh string) (*VirtualOutboundMeshView, error) {
	name := fmt.Sprintf(template, mesh)
	resource := config_model.NewConfigResource()
	err := m.configManager.Get(context.Background(), resource, store.GetByKey(name, ""))
	if err != nil {
		if store.IsResourceNotFound(err) {
			return NewVirtualOutboundView(map[HostnameEntry]VirtualOutbound{}), nil
		}
		return nil, err
	}

	if resource.Spec.Config == "" {
		return NewVirtualOutboundView(map[HostnameEntry]VirtualOutbound{}), nil
	}

	virtualOutboundView, err := m.unmarshal(resource.Spec.GetConfig())
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal")
	}

	return virtualOutboundView, nil
}

func (m *Persistence) Set(mesh string, vips *VirtualOutboundMeshView) error {
	ctx := context.Background()
	name := fmt.Sprintf(template, mesh)
	resource := config_model.NewConfigResource()
	create := false
	if err := m.configManager.Get(ctx, resource, store.GetByKey(name, model.NoMesh)); err != nil {
		if !store.IsResourceNotFound(err) {
			return err
		}
		create = true
	}

	jsonBytes, err := json.Marshal(vips.byHostname)
	if err != nil {
		return errors.Wrap(err, "unable to marshall VIP list")
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
