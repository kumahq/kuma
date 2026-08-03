package multizone

type GlobalLabels struct {
	// Labels with any of these prefixes won't be synced between control planes
	SkipPrefixes []string `json:"skipPrefixes,omitempty" envconfig:"kuma_multizone_global_kds_labels_skip_prefixes"`
}
type ZoneLabels struct {
	// Labels with any of these prefixes won't be synced between control planes
	SkipPrefixes []string `json:"skipPrefixes,omitempty" envconfig:"kuma_multizone_zone_kds_labels_skip_prefixes"`
}
