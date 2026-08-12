package xds

import (
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
)

// ForEachOrigin walks the ResourceSet's resources indexed by their origin,
// applying filters the same way rs.IndexByOrigin does.
func ForEachOrigin(
	rs *core_xds.ResourceSet,
	fn func(kri.Identifier, core_xds.ResourcesByType) error,
	filters ...func(*core_xds.Resource) bool,
) error {
	for uri, resType := range rs.IndexByOrigin(filters...) {
		if err := fn(uri, resType); err != nil {
			return err
		}
	}
	return nil
}

// ForEachOutboundRule walks the ResourceSet by origin like ForEachOrigin, computing the
// outbound rule for each origin and skipping origins with no computed rule or no conf.
func ForEachOutboundRule[T any](
	rs *core_xds.ResourceSet,
	rules outbound.ResourceRules,
	reader kri.ResourceReader,
	fn func(kri.Identifier, T, core_xds.ResourcesByType) error,
	filters ...func(*core_xds.Resource) bool,
) error {
	return ForEachOrigin(rs, func(uri kri.Identifier, resType core_xds.ResourcesByType) error {
		computed := rules.Compute(uri, reader)
		if computed == nil || len(computed.Conf) == 0 {
			return nil
		}
		return fn(uri, computed.Conf[0].(T), resType)
	}, filters...)
}
