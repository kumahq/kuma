package insights

import (
	"golang.org/x/time/rate"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

func Setup(rt runtime.Runtime) error {
	if rt.Config().Mode == config_core.Zone {
		return nil
	}
	resyncer := NewResyncer(&Config{
		ResourceManager:    rt.ResourceManager(),
		EventReaderFactory: rt.EventReaderFactory(),
		MinResyncTimeout:   rt.Config().Metrics.Mesh.MinResyncTimeout,
		MaxResyncTimeout:   rt.Config().Metrics.Mesh.MaxResyncTimeout,
		RateLimiterFactory: func() *rate.Limiter {
			return rate.NewLimiter(rate.Every(rt.Config().Metrics.Mesh.MinResyncTimeout), 0)
		},
		Registry: registry.Global(),
	})
	return rt.Add(component.NewResilientComponent(log, resyncer))
}
