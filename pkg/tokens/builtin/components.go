package builtin

import (
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/Kong/kuma/pkg/tokens/builtin/issuer"
)

func NewDataplaneTokenIssuer(rt runtime.Runtime) (issuer.DataplaneTokenIssuer, error) {
	key, err := issuer.GetSigningKey(rt.ResourceManager())
	if err != nil {
		return nil, err
	}
	return issuer.NewDataplaneTokenIssuer(key), nil
}
