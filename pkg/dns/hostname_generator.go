package dns

import (
	"slices"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/hostname"
	mes_hostname "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/hostname"
	ms_hostname "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/hostname"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

func SetupHostnameGenerator(rt runtime.Runtime) error {
	if rt.GetMode() == config_core.Global {
		return nil
	}
	logger := core.Log.WithName("hostnamegenerator")
	if !slices.Contains(rt.Config().CoreResources.Enabled, "hostnamegenerators") {
		logger.Info("HostnameGenerator is not enabled, not starting generator")
		return nil
	}
	mesGenerator := mes_hostname.NewMeshExternalServiceHostnameGenerator(rt.ResourceManager())
	msGenerator := ms_hostname.NewMeshServiceHostnameGenerator(rt.ResourceManager())
	generator, err := hostname.NewGenerator(
		logger,
		rt.Metrics(),
		rt.ResourceManager(),
		rt.Config().IPAM.AllocationInterval.Duration,
		[]hostname.HostnameGenerator{
			mesGenerator, msGenerator,
		},
	)
	if err != nil {
		return err
	}
	return rt.Add(component.NewResilientComponent(
		logger,
		generator,
		rt.Config().General.ResilientComponentBaseBackoff.Duration,
		rt.Config().General.ResilientComponentMaxBackoff.Duration,
	))
}
