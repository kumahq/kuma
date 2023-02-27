package tokens

import (
	"context"
	"crypto/rsa"
	errors2 "errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	store_config "github.com/kumahq/kuma/pkg/config/core/resources/store"
)

type Validator interface {
	// ParseWithValidation parses token and fills data in provided Claims.
	ParseWithValidation(ctx context.Context, token Token, claims Claims) error
}

type jwtTokenValidator struct {
	keyAccessors []SigningKeyAccessor
	revocations  Revocations
	storeType    store_config.StoreType
	log          logr.Logger
}

func NewValidator(log logr.Logger, keyAccessors []SigningKeyAccessor, revocations Revocations, storeType store_config.StoreType) Validator {
	return &jwtTokenValidator{
		log:          log,
		keyAccessors: keyAccessors,
		revocations:  revocations,
		storeType:    storeType,
	}
}

var _ Validator = &jwtTokenValidator{}

func (j *jwtTokenValidator) ParseWithValidation(ctx context.Context, rawToken Token, claims Claims) error {
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		var keyID KeyID
		kid, exists := token.Header[KeyIDHeader]
		if !exists {
			if _, ok := claims.(KeyIDFallback); ok {
				// KID wasn't supported in the past, so we use a marker interface to indicate which tokens were allowed
				// This will be removed with https://github.com/kumahq/kuma/issues/5519
				j.log.Info("[WARNING] Using token with KID header, you should rotate this token as it will not be valid in future versions of Kuma", "claims", claims, KeyIDHeader, 0)
				keyID = KeyIDFallbackValue
			} else {
				return 0, fmt.Errorf("JWT token must have %s header", KeyIDHeader)
			}
		} else {
			keyID = kid.(string)
		}
		switch token.Method.Alg() {
		case jwt.SigningMethodHS256.Name:
			var key []byte
			var err error
			for _, keyAccessor := range j.keyAccessors {
				key, err = keyAccessor.GetLegacyKey(ctx, KeyIDFallbackValue)
				if err == nil {
					return key, nil
				}
			}
			return nil, err
		case jwt.SigningMethodRS256.Name:
			var key *rsa.PublicKey
			var err error
			for _, keyAccessor := range j.keyAccessors {
				key, err = keyAccessor.GetPublicKey(ctx, keyID)
				if err == nil {
					return key, nil
				}
			}
			return nil, err
		default:
			return nil, fmt.Errorf("unsupported token alg %s. Allowed algorithms are %s and %s", token.Method.Alg(), jwt.SigningMethodRS256.Name, jwt.SigningMethodHS256)
		}
	})
	if err != nil {
		signingKeyError := &SigningKeyNotFound{}
		if errors2.As(err, &signingKeyError) {
			return signingKeyError
		}
		if j.storeType == store_config.MemoryStore {
			return errors.Wrap(err, "could not parse token. kuma-cp runs with an in-memory database and its state isn't preserved between restarts."+
				" Keep in mind that an in-memory database cannot be used with multiple instances of the control plane")
		}
		return errors.Wrap(err, "could not parse token")
	}
	if !token.Valid {
		return errors.New("token is not valid")
	}

	revoked, err := j.revocations.IsRevoked(ctx, claims.ID())
	if err != nil {
		return errors.Wrap(err, "could not check if the token is revoked")
	}
	if revoked {
		return errors.New("token is revoked")
	}
	return nil
}
