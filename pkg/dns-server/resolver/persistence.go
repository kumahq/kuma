package resolver

import (
	"context"
	"encoding/json"

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

func (p *DNSPersistence) Get() VIPList {
	viplist := VIPList{}
	resource := &config_model.ConfigResource{}
	err := p.manager.Get(context.Background(), resource, store.GetByKey("kuma-internal-config", ""))
	if err != nil {
		//simpleDNSLog.Error(err, "unable to get stored VIP list")
		return VIPList{}
	}

	err = json.Unmarshal([]byte(resource.Spec.Config), &viplist)
	if err != nil {
		simpleDNSLog.Error(err, "unable to unmarshall stored VIP list")
		return VIPList{}
	}

	return viplist
}

func (p *DNSPersistence) Set(viplist VIPList) {
	resource := &config_model.ConfigResource{}
	err := p.manager.Get(context.Background(), resource, store.GetByKey("kuma-internal-config", ""))
	if err != nil {
		err = p.manager.Create(context.Background(), resource, store.CreateByKey("kuma-internal-config", ""))
		if err != nil {
			simpleDNSLog.Error(err, "unable to create resource for VIP list")
		}
	}

	jsonBytes, err := json.Marshal(viplist)
	if err != nil {
		simpleDNSLog.Error(err, "unable to marshall VIP list")
		return
	}
	resource.Spec.Config = string(jsonBytes)

	err = p.manager.Update(context.Background(), resource)
	if err != nil {
		simpleDNSLog.Error(err, "unable to update VIP list")
		return
	}
}
