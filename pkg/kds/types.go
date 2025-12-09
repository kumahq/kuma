package kds

// Test change to verify KDS-specific copilot instructions pickup
const (
	googleApis = "type.googleapis.com/"

	// KumaResource is the type URL of the KumaResource protobuf.
	KumaResource = googleApis + "kuma.mesh.v1alpha1.KumaResource"

	MetadataFieldConfig    = "config"
	MetadataFieldVersion   = "version"
	MetadataFeatures       = "features"
	MetadataControlPlaneId = "control-plane-id"
)
