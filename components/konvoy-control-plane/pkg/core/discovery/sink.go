package discovery

import (
	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

var _ DiscoverySource = &DiscoverySink{}
var _ DiscoveryConsumer = &DiscoverySink{}

// DiscoverySink is both a source and a consumer of discovery information.
type DiscoverySink struct {
	ServiceConsumer   ServiceDiscoveryConsumer
	WorkloadConsumer  WorkloadDiscoveryConsumer
	DataplaneConsumer DataplaneDiscoveryConsumer
}

func (s *DiscoverySink) AddConsumer(consumer DiscoveryConsumer) {
	s.ServiceConsumer = consumer
	s.WorkloadConsumer = consumer
	s.DataplaneConsumer = consumer
}

func (s *DiscoverySink) OnServiceUpdate(svc *ServiceInfo) error {
	if s.ServiceConsumer != nil {
		return s.ServiceConsumer.OnServiceUpdate(svc)
	}
	return nil
}
func (s *DiscoverySink) OnServiceDelete(key core_model.ResourceKey) error {
	if s.ServiceConsumer != nil {
		return s.ServiceConsumer.OnServiceDelete(key)
	}
	return nil
}
func (s *DiscoverySink) OnWorkloadUpdate(wrk *WorkloadInfo) error {
	if s.WorkloadConsumer != nil {
		return s.WorkloadConsumer.OnWorkloadUpdate(wrk)
	}
	return nil
}
func (s *DiscoverySink) OnWorkloadDelete(key core_model.ResourceKey) error {
	if s.WorkloadConsumer != nil {
		return s.WorkloadConsumer.OnWorkloadDelete(key)
	}
	return nil
}
func (s *DiscoverySink) OnDataplaneUpdate(dataplane *mesh_core.DataplaneResource) error {
	if s.DataplaneConsumer != nil {
		return s.DataplaneConsumer.OnDataplaneUpdate(dataplane)
	}
	return nil
}
func (s *DiscoverySink) OnDataplaneDelete(key core_model.ResourceKey) error {
	if s.DataplaneConsumer != nil {
		return s.DataplaneConsumer.OnDataplaneDelete(key)
	}
	return nil
}
