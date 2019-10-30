package universal

import (
	"context"
	"github.com/Kong/kuma/pkg/core"
	core_discovery "github.com/Kong/kuma/pkg/core/discovery"
	core_resource "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"time"
)

var (
	log = core.Log.WithName("discovery").WithName("universal")
)

type dataplanesByKey = map[model.ResourceKey]*core_resource.DataplaneResource

var _ runtime.Component = &storePollingSource{}

// It periodically polls the given store with given interval for the dataplanes.
// When a new dataplane is added or old one is changed the DiscoverySink#OnDataplaneUpdate is called
// When a dataplane is removed the DiscoverySink#OnDataplaneDelete is called
type storePollingSource struct {
	store             store.ResourceStore
	currentDataplanes dataplanesByKey
	interval          time.Duration
	core_discovery.DiscoverySink
}

func (s *storePollingSource) Start(stop <-chan struct{}) error {
	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ticker.C:
			if err := s.detectChanges(); err != nil {
				log.Error(err, "error while detecting changes")
			}
		case <-stop:
			return nil
		}
	}
}

func newStorePollingSource(store store.ResourceStore, interval time.Duration) *storePollingSource {
	return &storePollingSource{
		store,
		make(dataplanesByKey),
		interval,
		core_discovery.DiscoverySink{},
	}
}

func (s *storePollingSource) detectChanges() error {
	dataplanes, err := s.fetchDataplanes()
	if err != nil {
		return errors.Wrap(err, "could not fetch dataplanes")
	}

	var multiErr error
	for _, dataplane := range newOrChangedDataplanes(s.currentDataplanes, dataplanes) {
		if err := s.OnDataplaneUpdate(dataplane); err != nil {
			multiErr = multierr.Combine(multiErr, errors.Wrap(err, "OnDataplaneUpdate callback returned an error"))
		}
	}

	for _, key := range deletedDataplanes(s.currentDataplanes, dataplanes) {
		if err := s.OnDataplaneDelete(key); err != nil {
			multiErr = multierr.Combine(multiErr, errors.Wrap(err, "OnDataplaneDelete callback returned an error"))
		}
	}
	s.currentDataplanes = dataplanes
	return multiErr
}

func (s *storePollingSource) fetchDataplanes() (dataplanesByKey, error) {
	dataplanesList := core_resource.DataplaneResourceList{}
	if err := s.store.List(context.Background(), &dataplanesList); err != nil {
		return nil, err
	}
	dataplanes := make(dataplanesByKey)
	for _, dataplane := range dataplanesList.Items {
		key := model.MetaToResourceKey(dataplane.Meta)
		dataplanes[key] = dataplane
	}
	return dataplanes, nil
}

func newOrChangedDataplanes(currentDataplanes dataplanesByKey, newDataplanes dataplanesByKey) []*core_resource.DataplaneResource {
	var result []*core_resource.DataplaneResource
	for key, dataplane := range newDataplanes {
		curResource, exists := currentDataplanes[key]
		if !exists || curResource.Meta.GetVersion() != dataplane.Meta.GetVersion() {
			result = append(result, dataplane)
		}
	}
	return result
}

func deletedDataplanes(currentDataplanes dataplanesByKey, newDataplanes dataplanesByKey) []model.ResourceKey {
	var result []model.ResourceKey
	for key := range currentDataplanes {
		_, exist := newDataplanes[key]
		if !exist {
			result = append(result, key)
		}
	}
	return result
}
