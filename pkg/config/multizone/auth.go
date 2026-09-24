package multizone

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/pkg/errors"

	config_types "github.com/kumahq/kuma/v3/pkg/config/types"
)

type KDSAuthType string

const (
	// KDSAuthNone disables authentication of Zone CPs.
	KDSAuthNone KDSAuthType = "none"
	// KDSAuthZoneToken requires Zone CPs to present a Zone Token with the "cp" scope.
	KDSAuthZoneToken KDSAuthType = "zoneToken"
)

// KDSServerAuthConfig defines how Global CP authenticates Zone CPs connecting over KDS.
type KDSServerAuthConfig struct {
	// Type of authentication. Available values: "none", "zoneToken", plus the types
	// a distribution registers on the KDS context. A type nothing authenticates with
	// is rejected when the KDS server starts, not here.
	Type KDSAuthType `json:"type" envconfig:"kuma_multizone_global_kds_auth_type"`
	// Configuration for the "zoneToken" authentication type.
	ZoneToken KDSZoneTokenAuthConfig `json:"zoneToken"`
}

func (c KDSServerAuthConfig) Validate() error {
	if c.Type == "" {
		return errors.Errorf(".Type cannot be empty. Use %q to disable authentication of Zone CPs", KDSAuthNone)
	}
	if err := c.ZoneToken.Validator.Validate(); err != nil {
		return errors.Wrap(err, ".ZoneToken.Validator is not valid")
	}
	return nil
}

type KDSZoneTokenAuthConfig struct {
	// If true the Global CP issues Zone Tokens over its API. Set it to false when
	// all the tokens are signed offline with the private key.
	EnableIssuer bool `json:"enableIssuer" envconfig:"kuma_multizone_global_kds_auth_zone_token_enable_issuer"`
	// Zone Token validator configuration
	Validator KDSZoneTokenValidatorConfig `json:"validator"`
}

type KDSZoneTokenValidatorConfig struct {
	// If true then Kuma secrets with prefix "zone-token-signing-key" are considered as signing keys.
	UseSecrets bool `json:"useSecrets" envconfig:"kuma_multizone_global_kds_auth_zone_token_validator_use_secrets"`
	// List of public keys used to validate the token
	PublicKeys []config_types.PublicKey `json:"publicKeys"`
}

func (c KDSZoneTokenValidatorConfig) Validate() error {
	for i, key := range c.PublicKeys {
		if err := key.Validate(); err != nil {
			return errors.Wrapf(err, ".PublicKeys[%d] is not valid", i)
		}
	}
	return nil
}

// KDSClientAuthConfig defines the credentials Zone CP presents to Global CP over KDS.
type KDSClientAuthConfig struct {
	// TokenInline is a Zone Token provided as a string.
	TokenInline string `json:"tokenInline" envconfig:"kuma_multizone_zone_kds_auth_token_inline"`
	// TokenPath is a path to a file with a Zone Token. It takes precedence over TokenInline.
	// The file is read on every new KDS stream, so the token can be rotated without restarting Zone CP.
	TokenPath string `json:"tokenPath" envconfig:"kuma_multizone_zone_kds_auth_token_path"`
}

func (c KDSClientAuthConfig) HasToken() bool {
	return c.TokenInline != "" || c.TokenPath != ""
}

// LoadToken returns the configured token or an empty string when none is configured.
func (c KDSClientAuthConfig) LoadToken() (string, error) {
	token := c.TokenInline
	if c.TokenPath != "" {
		// the checks are inline rather than in a helper, SAST tools recognize them better
		cleanPath := filepath.Clean(c.TokenPath)
		if slices.Contains(strings.Split(cleanPath, string(filepath.Separator)), "..") {
			return "", errors.Errorf("invalid zone token path: the path contains a traversal sequence: %s", c.TokenPath)
		}
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return "", errors.Wrapf(err, "invalid zone token path: could not resolve an absolute path for: %s", c.TokenPath)
		}
		bytes, err := os.ReadFile(absPath)
		if err != nil {
			return "", errors.Wrapf(err, "could not read zone token from file %s", c.TokenPath)
		}
		token = string(bytes)
	}
	// gRPC rejects metadata values with new lines or other non printable chars. https://github.com/grpc/grpc-go/issues/1888
	token = strings.TrimFunc(token, func(r rune) bool {
		return !unicode.IsGraphic(r) || unicode.IsSpace(r)
	})
	if token == "" {
		if c.TokenPath != "" {
			return "", errors.Errorf("zone token in file %s is empty", c.TokenPath)
		}
		if c.TokenInline != "" {
			return "", errors.New("zone token in .TokenInline is empty")
		}
	}
	return token, nil
}
