package ipam

import (
	"slices"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/vip"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

func Setup(rt runtime.Runtime) error {
	if !slices.Contains(rt.Config().CoreResources.Enabled, "meshservice") {
		return nil
	}
	logger := core.Log.WithName("meshservice").WithName("vips").WithName("allocator")
	allocator, err := vip.NewAllocator(
		logger,
		rt.Metrics(),
		rt.ResourceManager(),
		rt.Config().IPAM.MeshService.CIDR,
		rt.Config().IPAM.MeshService.AllocationInterval.Duration,
	)
	if err != nil {
		return err
	}
	return rt.Add(component.NewResilientComponent(
		logger,
		allocator,
		rt.Config().General.ResilientComponentBaseBackoff.Duration,
		rt.Config().General.ResilientComponentMaxBackoff.Duration,
	))
}
