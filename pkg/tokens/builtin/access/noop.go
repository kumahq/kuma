package access

import "github.com/kumahq/kuma/pkg/core/user"

type NoopDpTokenAccess struct{}

var _ DataplaneTokenAccess = NoopDpTokenAccess{}

func (n NoopDpTokenAccess) ValidateGenerateDataplaneToken(name string, mesh string, tags map[string][]string, user user.User) error {
	return nil
}

func (n NoopDpTokenAccess) ValidateGenerateZoneIngressToken(zone string, user user.User) error {
	return nil
}
