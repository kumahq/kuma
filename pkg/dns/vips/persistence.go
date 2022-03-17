package vips

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
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
		v, err := m.unmarshal(resource.Spec.GetConfig(), mesh)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		meshed[mesh] = v
	}
	return
}

func (m *Persistence) unmarshal(config string, mesh string) (*VirtualOutboundMeshView, error) {
	res := map[HostnameEntry]VirtualOutbound{}
	if err := json.Unmarshal([]byte(config), &res); err == nil {
		return NewVirtualOutboundView(res)
	}
	backCompat := map[HostnameEntry]string{}
	if err := json.Unmarshal([]byte(config), &backCompat); err != nil {
		// backwards compatibility
		// https://github.com/kumahq/kuma/issues/4003
		backwardCompatible := map[string]string{}
		if err := json.Unmarshal([]byte(config), &backwardCompatible); err != nil {
			return nil, err
		}
		for service, vip := range backwardCompatible {
			backCompat[NewServiceEntry(service)] = vip
		}
	}
	// When doing backward compatibility we need to find which host belongs to which externalService.
	externalServices := core_mesh.ExternalServiceResourceList{}
	if err := m.resourceManager.List(context.Background(), &externalServices, store.ListByMesh(mesh)); err != nil {
		return nil, err
	}
	srvByHost := map[string][]*core_mesh.ExternalServiceResource{}
	for _, s := range externalServices.Items {
		srvByHost[s.Spec.GetHost()] = append(srvByHost[s.Spec.GetHost()], s)
	}
	res = map[HostnameEntry]VirtualOutbound{}
	for entry, vip := range backCompat {
		switch entry.Type {
		case Host:
			allPorts := map[uint32]bool{}
			for _, es := range srvByHost[entry.Name] {
				port := es.Spec.GetPortUInt32()
				if exists := allPorts[port]; !exists {
					res[entry] = VirtualOutbound{
						Address:   vip,
						Outbounds: append(res[entry].Outbounds, OutboundEntry{Port: port, TagSet: map[string]string{mesh_proto.ServiceTag: es.Spec.GetService()}, Origin: "legacy-host-upgrade"}),
					}
					allPorts[port] = true
				}
			}
		case Service:
			res[entry] = VirtualOutbound{
				Address: vip,
				Outbounds: []OutboundEntry{
					{TagSet: map[string]string{mesh_proto.ServiceTag: entry.Name}, Origin: "legacy-service-upgrade"},
				},
			}
		}
	}
	return NewVirtualOutboundView(res)
}

func (m *Persistence) GetByMesh(mesh string) (*VirtualOutboundMeshView, error) {
	name := fmt.Sprintf(template, mesh)
	resource := config_model.NewConfigResource()
	err := m.configManager.Get(context.Background(), resource, store.GetByKey(name, ""))
	if err != nil {
		if store.IsResourceNotFound(err) {
			return NewEmptyVirtualOutboundView(), nil
		}
		return nil, err
	}

	if resource.Spec.Config == "" {
		return NewEmptyVirtualOutboundView(), nil
	}

	virtualOutboundView, err := m.unmarshal(resource.Spec.GetConfig(), mesh)
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
