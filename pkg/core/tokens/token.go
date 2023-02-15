package tokens

import "github.com/golang-jwt/jwt/v4"

type Token = string

type Claims interface {
	jwt.Claims
	ID() string
	SetRegisteredClaims(claims jwt.RegisteredClaims)
}

type KeyID = string

const KeyIDFallbackValue = "0"

type KeyIDFallback interface {
	// KeyIDFallback Marker function to indicate this can be used for tokens with v0
	// This will be removed with https://github.com/kumahq/kuma/issues/5519
	KeyIDFallback()
}

const KeyIDHeader = "kid" // standard JWT header that indicates which signing key we should use
