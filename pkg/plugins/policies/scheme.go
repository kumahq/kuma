package policies

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kumahq/kuma/pkg/plugins/policies/core"
)

func AddToScheme(s *runtime.Scheme) error {
	for i := range core.AllSchemes {
		if err := core.AllSchemes[i](s); err != nil {
			return err
		}
	}
	return nil
}
