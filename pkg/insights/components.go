package insights

import (
	"github.com/kumahq/kuma/pkg/core/runtime"
)

func Setup(rt runtime.Runtime) error {
	resyncer := NewResyncer(&Config{
		ResourceManager:  rt.ResourceManager(),
		EventReader:      rt.EventReaderFactory().New(),
		MinResyncTimeout: rt.Config().Metrics.Mesh.MinResyncTimeout,
		MaxResyncTimeout: rt.Config().Metrics.Mesh.MaxResyncTimeout,
	})
	return rt.Add(resyncer)
}
