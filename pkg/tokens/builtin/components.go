package builtin

import (
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

func NewDataplaneTokenIssuer(rt runtime.Runtime) (issuer.DataplaneTokenIssuer, error) {
	return issuer.NewDataplaneTokenIssuer(func() ([]byte, error) {
		return issuer.GetSigningKey(rt.ResourceManager())
	}), nil
}
