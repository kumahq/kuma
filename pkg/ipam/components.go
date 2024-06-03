package ipam

import (
	"slices"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
	mes_vip "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/vip"
	ms_vip "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/vip"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

func Setup(rt runtime.Runtime) error {
	if rt.GetMode() == config_core.Global {
		return nil
	}
	logger := core.Log.WithName("vips").WithName("allocator")
	if !slices.Contains(rt.Config().CoreResources.Enabled, "meshservices") {
		logger.Info("MeshService is not enabled. Skip starting VIP allocator for MeshService.")
		return nil
	}
	meshServiceAllocator, err := ms_vip.NewMeshServiceAllocator(
		core.Log.WithName("vips").WithName("allocator").WithName("mesh-service"),
		rt.Config().IPAM.MeshService.CIDR,
		rt.ResourceManager(),
		rt.Config().IPAM.AllocationInterval.Duration,
		rt.Metrics(),
	)
	if err != nil {
		return err
	}
	meshExternalServiceAllocator, err := mes_vip.NewMeshExternalServiceAllocator(
		core.Log.WithName("vips").WithName("allocator").WithName("mesh-external-service"),
		rt.Config().IPAM.MeshExternalService.CIDR,
		rt.ResourceManager(),
		rt.Config().IPAM.AllocationInterval.Duration,
		rt.Metrics(),
	)
	if err != nil {
		return err
	}
	allocator, err := vip.NewAllocator(
		logger,
		rt.Config().IPAM.AllocationInterval.Duration,
		[]vip.VIPAllocator{
			meshExternalServiceAllocator,
			meshServiceAllocator,
		},
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
