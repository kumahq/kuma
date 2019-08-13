package discovery

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

var _ DiscoverySource = &DiscoverySink{}
var _ DiscoveryConsumer = &DiscoverySink{}

// DiscoverySink is both a source and a consumer of discovery information.
type DiscoverySink struct {
	Consumer DiscoveryConsumer
}

func (s *DiscoverySink) AddConsumer(c DiscoveryConsumer) {
	s.Consumer = c
}

func (s *DiscoverySink) OnServiceUpdate(svc *ServiceInfo) error {
	if s.Consumer != nil {
		return s.Consumer.OnServiceUpdate(svc)
	}
	return nil
}
func (s *DiscoverySink) OnServiceDelete(name core.NamespacedName) error {
	if s.Consumer != nil {
		return s.Consumer.OnServiceDelete(name)
	}
	return nil
}
func (s *DiscoverySink) OnWorkloadUpdate(wrk *WorkloadInfo) error {
	if s.Consumer != nil {
		return s.Consumer.OnWorkloadUpdate(wrk)
	}
	return nil
}
func (s *DiscoverySink) OnWorkloadDelete(name core.NamespacedName) error {
	if s.Consumer != nil {
		return s.Consumer.OnWorkloadDelete(name)
	}
	return nil
}
func (s *DiscoverySink) OnDataplaneUpdate(dataplane *mesh_core.DataplaneResource) error {
	if s.Consumer != nil {
		return s.Consumer.OnDataplaneUpdate(dataplane)
	}
	return nil
}
func (s *DiscoverySink) OnDataplaneDelete(key model.ResourceKey) error {
	if s.Consumer != nil {
		return s.Consumer.OnDataplaneDelete(key)
	}
	return nil
}
