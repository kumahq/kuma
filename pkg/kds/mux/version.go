package mux

const (
	KDSVersionHeaderKey = "kds-version"
	KDSVersionV3        = "v3"

	// ZoneVersionHeaderKey marks the zone CP's KDS generation so the global CP
	// can route it. Set unconditionally by clients built from this code line.
	ZoneVersionHeaderKey = "zone-version"
	ZoneVersionV3        = "v3"
)
