package access

import (
	"github.com/kumahq/kuma/pkg/core/user"
)

type DataplaneTokenAccess interface {
	ValidateGenerateDataplaneToken(name string, mesh string, tags map[string][]string, user user.User) error
	ValidateGenerateZoneIngressToken(zone string, user user.User) error
}
