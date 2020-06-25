package dns

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	config_manager "github.com/Kong/kuma/pkg/core/config/manager"
	config_model "github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

type DNSPersistence struct {
	manager config_manager.ConfigManager
}

func NewDNSPersistence(manager config_manager.ConfigManager) *DNSPersistence {
	return &DNSPersistence{
		manager: manager,
	}
}

const dnsConfigKey = "kuma-dns-vips"

func (p *DNSPersistence) Get() (VIPList, error) {
	viplist := VIPList{}
	resource := &config_model.ConfigResource{}
	err := p.manager.Get(context.Background(), resource, store.GetByKey(dnsConfigKey, ""))
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

func (p *DNSPersistence) Set(viplist VIPList) error {
	resource := &config_model.ConfigResource{}
	err := p.manager.Get(context.Background(), resource, store.GetByKey(dnsConfigKey, ""))
	if err != nil {
		if store.IsResourceNotFound(err) {
			if err := p.manager.Create(context.Background(), resource, store.CreateByKey(dnsConfigKey, "")); err != nil {
				return errors.Wrap(err, "could not create config")
			}
		} else {
			return err
		}
	}

	jsonBytes, err := json.Marshal(viplist)
	if err != nil {
		return errors.Wrap(err, "unable to marshall VIP list")
	}
	resource.Spec.Config = string(jsonBytes)

	err = p.manager.Update(context.Background(), resource)
	if err != nil {
		return errors.Wrap(err, "unable to update VIP list")
	}
	return nil
}
