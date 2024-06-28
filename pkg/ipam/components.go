package ipam

import (
	"slices"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
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
	allocator, err := vip.NewAllocator(
		logger,
		rt.Config().IPAM.AllocationInterval.Duration,
		map[string]core_model.ResourceTypeDescriptor{
			rt.Config().IPAM.MeshService.CIDR:          meshservice_api.MeshServiceResourceTypeDescriptor,
			rt.Config().IPAM.MeshExternalService.CIDR:  meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			rt.Config().IPAM.MeshMultiZoneService.CIDR: meshmzservice_api.MeshMultiZoneServiceResourceTypeDescriptor,
		},
		rt.Metrics(),
		rt.ResourceManager(),
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
