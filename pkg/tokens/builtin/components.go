package builtin

import (
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

func NewDataplaneTokenIssuer(rt runtime.Runtime) (issuer.DataplaneTokenIssuer, error) {
	key, err := issuer.GetSigningKey(rt.ResourceManager())
	if err != nil {
		return nil, err
	}
	return issuer.NewDataplaneTokenIssuer(key), nil
}
