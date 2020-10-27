package persistence

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

const ConfigKey = "kuma-dns-vips"

type global struct {
	manager config_manager.ConfigManager
}

func NewGlobalPersistence(manager config_manager.ConfigManager) GlobalWriter {
	return &global{
		manager: manager,
	}
}

func (g *global) Get() (VIPList, error) {
	viplist := VIPList{}
	resource := &config_model.ConfigResource{}
	err := g.manager.Get(context.Background(), resource, store.GetByKey(ConfigKey, ""))
	if err != nil {
		if store.IsResourceNotFound(err) {
			return viplist, nil
		}
		return nil, err
	}

	err = json.Unmarshal([]byte(resource.Spec.Config), &viplist)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal")
	}

	return viplist, nil
}

func (g *global) Set(vips VIPList) error {
	resource := &config_model.ConfigResource{}
	err := g.manager.Get(context.Background(), resource, store.GetByKey(ConfigKey, ""))
	if err != nil {
		if store.IsResourceNotFound(err) {
			if err := g.manager.Create(context.Background(), resource, store.CreateByKey(ConfigKey, "")); err != nil {
				return errors.Wrap(err, "could not create config")
			}
		} else {
			return err
		}
	}

	jsonBytes, err := json.Marshal(vips)
	if err != nil {
		return errors.Wrap(err, "unable to marshall VIP list")
	}
	resource.Spec.Config = string(jsonBytes)

	err = g.manager.Update(context.Background(), resource)
	if err != nil {
		return errors.Wrap(err, "unable to update VIP list")
	}
	return nil
}
