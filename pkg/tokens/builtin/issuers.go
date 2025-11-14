package builtin

import (
	"github.com/kumahq/kuma/v2/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/v2/pkg/tokens/builtin/zone"
)

type TokenIssuers struct {
	DataplaneToken issuer.DataplaneTokenIssuer
	ZoneToken      zone.TokenIssuer
}
