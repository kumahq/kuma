package builtin

import (
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
)

type TokenIssuers struct {
	DataplaneToken   issuer.DataplaneTokenIssuer
	ZoneIngressToken zoneingress.TokenIssuer
	ZoneToken        zone.TokenIssuer
}
