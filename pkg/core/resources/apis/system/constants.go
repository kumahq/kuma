package system

const (
	// AdminUserToken is the name of the global secret holding the token for the admin user
	AdminUserToken = "admin-user-token"
	// EnvoyAdminCA is the name of the global secret holding the CA for the Envoy Admin
	EnvoyAdminCA = "envoy-admin-ca"
	// InterCpCA is the name of the global secret holding the CA for the inter control plane communication
	InterCpCA = "inter-cp-ca"
	// ZoneTokenSigningKeyPrefix is the prefix for the global secret holding the zone ingress token signing key
	ZoneTokenSigningKeyPrefix = "zone-token-signing-key"
	// ZoneTokenSigningPublicKeyPrefix is the prefix for the global secret holding the zone ingress token signing public key
	ZoneTokenSigningPublicKeyPrefix = "zone-token-signing-public-key"
	// ZoneTokenRevocations is the name of the global secret holding the zone token revocations
	ZoneTokenRevocations = "zone-token-revocations" // #nosec G101 -- this is not a credential

	// UserTokenSigningKeyPrefix is the prefix for the global secret holding the user token signing key
	UserTokenSigningKeyPrefix = "user-token-signing-key"
	// UserTokenRevocations is the name of the global secret holding the user token revocations
	UserTokenRevocations = "user-token-revocations" // #nosec G101 -- this is not a credential

	// DataplaneTokenSigningKeyPrefix is the prefix for the secret holding the dataplane token signing key
	DataplaneTokenSigningKeyPrefix = "dataplane-token-signing-key-"
	// DataplaneTokenRevolationsPrefix is the prefix for the secret holding the dataplane token revocations
	DataplaneTokenRevocationsPrefix = "dataplane-token-revocations-"
)

func DataplaneTokenSigningKey(mesh string) string {
	return DataplaneTokenSigningKeyPrefix + mesh
}

func DataplaneTokenRevocations(mesh string) string {
	return DataplaneTokenRevocationsPrefix + mesh
}
