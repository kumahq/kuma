package mesh_test

import (
	"context"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/client"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/model"
	. "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/model/mesh"
)

func Prototype() {
	observe := func(_ model.Resource) {
	}

	trr := TrafficRouteResource{
		Spec: TrafficRoute{},
	}

	observe(&trr)

	var c client.ResourceClient
	tr := &TrafficRouteResource{}
	if err := c.Get(context.Background(), tr, client.GetByName("konvoy-system", "global-route")); err != nil {
		tr.Spec.Path = "somethig else"
	}

	tocreate := &TrafficRouteResource{
		Spec: TrafficRoute{
			Path: "some path",
		},
	}
	_ = c.Create(context.Background(), tocreate, client.CreateByName("konvoy-system", "global-route"))

	items := &TrafficRouteResourceList{}
	c.List(context.Background(), items, client.ListByNamespace("ns-1"))
}
