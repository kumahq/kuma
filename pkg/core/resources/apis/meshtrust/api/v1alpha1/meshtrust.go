// +kubebuilder:object:generate=true
package v1alpha1

// MeshTrust defines trusted Certificate Authority (CA) bundles for a trust domain in the mesh. It establishes trust relationships for service-to-service mTLS authentication by specifying which CA certificates are trusted to verify service identities, supporting PEM-encoded CA bundles and enabling secure cross-service communication within the trust domain.
// +kuma:policy:is_policy=false
// +kuma:policy:allowed_on_system_namespace_only=true
// +kuma:policy:kds_flags=model.GlobalToZonesFlag | model.ZoneToGlobalFlag
// +kuma:policy:short_name=mtrust
// +kuma:policy:register_generator=true
// +kuma:policy:has_status=true
type MeshTrust struct {
	// Origin specifies whether the resource was created from a MeshIdentity.
	//
	// Deprecated: use Status.Origin instead
	Origin *Origin `json:"origin,omitempty"`
	// TrustDomain is the trust domain associated with this resource.
	// +required
	// +kubebuilder:validation:MaxLength=253
	TrustDomain string `json:"trustDomain"`
	// CABundles contains a list of CA bundles supported by this TrustDomain.
	// At least one CA bundle must be specified.
	// +required
	// +kubebuilder:validation:MinItems=1
	CABundles []CABundle `json:"caBundles"`
}

type MeshTrustStatus struct {
	// Origin specifies whether the resource was created from a MeshIdentity.
	Origin *Origin `json:"origin,omitempty"`
}

type Origin struct {
	// Resource identifier
	KRI *string `json:"kri,omitempty"`
}

// +kubebuilder:validation:Enum=Pem
type CABundleType string

const (
	PemCABundleType CABundleType = "Pem"
)

type CABundle struct {
	// Type specifies the format or source type of the CA bundle.
	// +required
	Type CABundleType `json:"type"`
	// Pem contains the PEM-encoded CA bundle if the Type is set to a PEM-based format.
	PEM *PEM `json:"pem,omitempty"`
}

type PEM struct {
	// Value holds the PEM-encoded CA bundle as a string.
	Value string `json:"value"`
}
