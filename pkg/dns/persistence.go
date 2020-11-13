package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type VIPList map[string]string

func (vips VIPList) Append(other VIPList) {
	for k, v := range other {
		vips[k] = v
	}
}

type MeshedPersistence struct {
	manager config_manager.ConfigManager
}

const template = "kuma-%s-dns-vips"

var re = regexp.MustCompile(fmt.Sprintf(template, `(.*)`))

func MeshedConfigKey(name string) (string, bool) {
	match := re.FindStringSubmatch(name)
	if len(match) < 2 {
		return "", false
	}
	mesh := match[1]
	return mesh, true
}

func NewMeshedPersistence(manager config_manager.ConfigManager) *MeshedPersistence {
	return &MeshedPersistence{
		manager: manager,
	}
}

func (m *MeshedPersistence) Get() (VIPList, error) {
	resourceList := &config_model.ConfigResourceList{}
	if err := m.manager.List(context.Background(), resourceList); err != nil {
		return nil, err
	}

	var errs error
	vips := VIPList{}
	for _, resource := range resourceList.Items {
		if _, ok := MeshedConfigKey(resource.Meta.GetName()); !ok {
			continue
		}
		if resource.Spec.Config == "" {
			continue
		}
		v := VIPList{}
		if err := json.Unmarshal([]byte(resource.Spec.Config), &v); err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		vips.Append(v)
	}
	return vips, nil
}

func (m *MeshedPersistence) GetByMesh(mesh string) (VIPList, error) {
	name := fmt.Sprintf(template, mesh)
	vips := VIPList{}
	resource := &config_model.ConfigResource{}
	err := m.manager.Get(context.Background(), resource, store.GetByKey(name, ""))
	if err != nil {
		if store.IsResourceNotFound(err) {
			return vips, nil
		}
		return nil, err
	}

	if resource.Spec.Config == "" {
		return vips, nil
	}

	err = json.Unmarshal([]byte(resource.Spec.Config), &vips)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal")
	}

	return vips, nil
}

func (m *MeshedPersistence) Set(mesh string, vips VIPList) error {
	name := fmt.Sprintf(template, mesh)
	resource := &config_model.ConfigResource{}

	jsonBytes, err := json.Marshal(vips)
	if err != nil {
		return errors.Wrap(err, "unable to marshall VIP list")
	}

	if err := config_manager.Upsert(m.manager, model.ResourceKey{Name: name, Mesh: model.NoMesh}, resource, func(resource *config_model.ConfigResource) {
		resource.Spec.Config = string(jsonBytes)
	}); err != nil {
		return err
	}
	return nil
}
